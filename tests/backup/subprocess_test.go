package backup_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type toolCapture struct {
	Args, Environment     []string
	Passfile, PassContent string
	Mode                  uint32
}

// TestMain doubles as a deliberately failing external PostgreSQL tool. It records
// its real process boundary to a private test file, never to public test output.
func TestMain(m *testing.M) {
	if path := os.Getenv("BACKUP_TEST_TOOL_CAPTURE"); path != "" {
		passfile := os.Getenv("PGPASSFILE")
		data, _ := os.ReadFile(passfile)
		stat, _ := os.Stat(passfile)
		record := toolCapture{Args: os.Args[1:], Environment: os.Environ(), Passfile: passfile, PassContent: string(data)}
		if stat != nil {
			record.Mode = uint32(stat.Mode().Perm())
		}
		encoded, _ := json.Marshal(record)
		if os.WriteFile(path, encoded, 0600) != nil {
			os.Exit(20)
		}
		if os.Getenv("BACKUP_TEST_TOOL_MODE") == "wait-run" {
			deadline := time.Now().Add(10 * time.Second)
			for {
				if _, err := os.Stat(os.Getenv("BACKUP_TEST_TOOL_CONTINUE")); err == nil {
					break
				}
				if time.Now().After(deadline) {
					os.Exit(21)
				}
				time.Sleep(10 * time.Millisecond)
			}
			command := exec.Command(os.Getenv("BACKUP_TEST_REAL_TOOL"), os.Args[1:]...)
			command.Env = os.Environ()
			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Stderr = io.Discard
			if command.Run() != nil {
				os.Exit(22)
			}
			os.Exit(0)
		}
		fmt.Fprintln(os.Stdout, "partial archive")
		fmt.Fprintln(os.Stderr, "private-tool-password private-tool-email private-tool-dsn")
		os.Exit(19)
	}
	os.Exit(m.Run())
}

func fakeTools(t *testing.T) string {
	t.Helper()
	binary, err := os.Executable()
	dbFail(t, "test executable", err)
	dir := t.TempDir()
	for _, name := range []string{"pg_dump", "pg_restore"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		source, err := os.Open(binary)
		dbFail(t, "open test executable", err)
		out, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
		dbFail(t, "create fake tool", err)
		_, err = io.Copy(out, source)
		source.Close()
		out.Close()
		dbFail(t, "copy fake tool", err)
	}
	return dir
}

func assertToolCapture(t *testing.T, capture string, restore bool) {
	t.Helper()
	encoded, err := os.ReadFile(capture)
	dbFail(t, "read private tool capture", err)
	var record toolCapture
	dbFail(t, "decode private tool capture", json.Unmarshal(encoded, &record))
	args := strings.Join(record.Args, " ")
	for _, forbidden := range []string{"postgres://", "postgresql://", "password=", "--clean", "--create", "--file"} {
		if strings.Contains(args, forbidden) {
			t.Fatal("unsafe external tool argument")
		}
	}
	for _, arg := range []string{"--no-owner", "--no-privileges", "--no-password"} {
		if !strings.Contains(args, arg) {
			t.Fatal("required tool safety flag absent")
		}
	}
	if restore {
		for _, arg := range []string{"--single-transaction", "--exit-on-error", "--dbname"} {
			if !strings.Contains(args, arg) {
				t.Fatal("required restore safety flag absent")
			}
		}
	} else if !strings.Contains(args, "--snapshot") || !strings.Contains(args, "--format=custom") {
		t.Fatal("dump lost consistency or archive format flag")
	}
	for _, entry := range record.Environment {
		key, value, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		if key == "PGPASSWORD" || key == "DATABASE_URL" || key == "RESTORE_DATABASE_URL" || key == "TEST_DATABASE_URL" || key == "MAIL_QUEUE_KEY" || strings.HasPrefix(key, "SMTP_") {
			t.Fatal("external tool inherited sensitive environment")
		}
		if key == "PGDATABASE" && (strings.Contains(value, "=") || strings.Contains(value, "://")) {
			t.Fatal("tool database name can be interpreted as DSN")
		}
	}
	if record.PassContent == "" {
		t.Fatal("external process has no private password file")
	}
	if runtime.GOOS != "windows" && record.Mode != 0600 {
		t.Fatal("password file permissions too broad")
	}
	if _, err := os.Stat(filepath.Dir(record.Passfile)); !os.IsNotExist(err) {
		t.Fatal("private password file directory was not removed")
	}
}
