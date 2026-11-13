package db_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"gotest/tests/testenv"
)

func runCombatBrowserDriver(ctx context.Context, browser, profile string, args ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Go owns Chrome independently of Node, so even TerminateProcess on the driver
	// cannot bypass browser cleanup. The profile belongs exclusively to this test.
	chrome := exec.Command(browser, "--headless=new", "--remote-debugging-port=0", "--user-data-dir="+profile,
		"--disable-background-networking", "--no-first-run", "--no-default-browser-check", "--disable-extensions", "--disable-gpu", "about:blank")
	if err := chrome.Start(); err != nil {
		return nil, err
	}
	defer func() { _ = chrome.Process.Kill(); _ = chrome.Wait() }()
	driver := exec.CommandContext(ctx, "node", args...)
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(name, "COMBAT_BROWSER_PARENT") {
			driver.Env = append(driver.Env, entry)
		}
	}
	driver.Env = append(driver.Env, "COMBAT_BROWSER_PARENT=1")
	return driver.CombinedOutput()
}

func combatBrowserEndpoint(profile string) (string, string, error) {
	data, err := os.ReadFile(filepath.Join(profile, "DevToolsActivePort"))
	if err != nil {
		return "", "", err
	}
	lines := strings.Fields(string(data))
	if len(lines) != 2 {
		return "", "", fmt.Errorf("invalid browser endpoint")
	}
	address := net.JoinHostPort("127.0.0.1", lines[0])
	return address, "ws://" + address + lines[1], nil
}

func TestCombatBrowserDriverCancellationClosesOwnedChrome(t *testing.T) {
	browser := testenv.Browser(t)
	profile, scratch := t.TempDir(), t.TempDir()
	helper, err := filepath.Abs(filepath.Join("..", "frontend", "helpers", "chrome.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	marker, script := filepath.Join(scratch, "ready"), filepath.Join(scratch, "hung.mjs")
	code := `import {pathToFileURL} from 'node:url';
import {writeFile} from 'node:fs/promises';
const {launchChrome}=await import(pathToFileURL(process.argv[4]));
await launchChrome(process.argv[2],process.argv[3]);
await writeFile(process.argv[5],'ready');
setInterval(()=>{},1000);`
	if err = os.WriteFile(script, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		defer close(done)
		_, err := runCombatBrowserDriver(ctx, browser, profile, script, browser, profile, helper, marker)
		done <- err
	}()
	t.Cleanup(func() { cancel(); <-done })
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err = os.Stat(marker); err == nil {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("driver exited before ready: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			cancel()
			<-done
			t.Fatal("browser did not become ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
	address, endpoint, err := combatBrowserEndpoint(profile)
	if err != nil {
		t.Fatal(err)
	}
	// The regression must clean its leaked browser even while the assertion is RED.
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		connection, _, err := websocket.Dial(cleanup, endpoint, nil)
		if err == nil {
			_ = connection.Write(cleanup, websocket.MessageText, []byte(`{"id":1,"method":"Browser.close"}`))
			_, _, _ = connection.Read(cleanup)
			connection.CloseNow()
		}
	}()
	cancel()
	if err = <-done; err == nil {
		t.Fatal("cancelled driver succeeded")
	}
	connection, err := net.DialTimeout("tcp", address, 300*time.Millisecond)
	if err == nil {
		connection.Close()
		t.Fatal("Chrome survived cancellation of its Node driver")
	}
}
