package mailer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"gotest/internal/mailer"
	"io"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

type smtpResult struct {
	message string
	err     error
}

func fakeSMTP(t *testing.T, mode string, cert tls.Certificate, disconnectAfterAccept ...bool) (string, <-chan smtpResult) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	result := make(chan smtpResult, 1)
	go func() {
		c, err := listener.Accept()
		if err != nil {
			result <- smtpResult{err: err}
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(3 * time.Second))
		config := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
		if mode == "tls" {
			c = tls.Server(c, config)
		}
		p := textproto.NewConn(c)
		fail := func(err error) { result <- smtpResult{err: err} }
		if err = p.PrintfLine("220 test ESMTP"); err != nil {
			fail(err)
			return
		}
		encrypted := mode == "tls"
		authenticated := false
		for {
			line, err := p.ReadLine()
			if err != nil {
				fail(err)
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO"):
				if !encrypted {
					err = p.PrintfLine("250-test\r\n250 STARTTLS")
				} else {
					err = p.PrintfLine("250-test\r\n250 AUTH PLAIN")
				}
			case line == "STARTTLS":
				if err = p.PrintfLine("220 ready"); err == nil {
					c = tls.Server(c, config)
					err = c.(*tls.Conn).Handshake()
					p = textproto.NewConn(c)
					encrypted = true
				}
			case line == "MAIL FROM:<sender@example.com>" || line == "RCPT TO:<user@example.com>":
				if !authenticated {
					fail(fmt.Errorf("missing SMTP authentication"))
					return
				}
				err = p.PrintfLine("250 ok")
			case line == "AUTH PLAIN AHVzZXIAcGFzcw==":
				if !encrypted {
					fail(fmt.Errorf("unencrypted authentication"))
					return
				}
				err = p.PrintfLine("235 authenticated")
				authenticated = true
			case line == "DATA":
				if err = p.PrintfLine("354 send"); err != nil {
					fail(err)
					return
				}
				body, readErr := io.ReadAll(p.DotReader())
				if readErr != nil {
					fail(readErr)
					return
				}
				if !encrypted {
					fail(fmt.Errorf("unencrypted message"))
					return
				}
				if err = p.PrintfLine("250 accepted"); err != nil {
					fail(err)
					return
				}
				result <- smtpResult{message: string(body)}
				if len(disconnectAfterAccept) > 0 && disconnectAfterAccept[0] {
					return
				}
			case line == "QUIT":
				p.PrintfLine("221 bye")
				return
			default:
				fail(fmt.Errorf("unexpected command %q", line))
				return
			}
			if err != nil {
				fail(err)
				return
			}
		}
	}()
	return listener.Addr().String(), result
}

func TestSMTPRequiresSTARTTLSBeforeSendingCredentials(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() {
		c, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(time.Second))
		p := textproto.NewConn(c)
		p.PrintfLine("220 test ESMTP")
		line, err := p.ReadLine()
		if err != nil {
			done <- err
			return
		}
		if !strings.HasPrefix(line, "EHLO") {
			done <- fmt.Errorf("expected EHLO, got %q", line)
			return
		}
		p.PrintfLine("250-test\r\n250 AUTH PLAIN")
		line, err = p.ReadLine()
		if err == nil {
			done <- fmt.Errorf("sent plaintext command %q", line)
			return
		}
		done <- nil
	}()
	m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: listener.Addr().String(), SMTPUsername: "user", SMTPPassword: "pass", SMTPTLSMode: "starttls", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SendVerification(context.Background(), "user@example.com", "secret"); err == nil {
		t.Fatal("accepted SMTP without STARTTLS")
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
}

func TestSMTPDeliversOverVerifiedTLSAndSTARTTLS(t *testing.T) {
	for _, mode := range []string{"tls", "starttls"} {
		t.Run(mode, func(t *testing.T) {
			address, result := fakeSMTP(t, mode, trustedCertificate)
			m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: address, SMTPUsername: "user", SMTPPassword: "pass", SMTPTLSMode: mode, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if err = m.SendVerification(context.Background(), "user@example.com", "secret"); err != nil {
				t.Fatal(err)
			}
			got := <-result
			if got.err != nil {
				t.Fatal(got.err)
			}
			if !strings.Contains(got.message, "https://example.com/verify?token=3Dsecret") || !strings.Contains(got.message, "MIME-Version: 1.0") {
				t.Fatalf("incorrect message: %s", got.message)
			}
		})
	}
}

func TestSMTPRejectsUntrustedCertificate(t *testing.T) {
	for _, mode := range []string{"tls", "starttls"} {
		t.Run(mode, func(t *testing.T) {
			address, _ := fakeSMTP(t, mode, untrustedCertificate)
			m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: address, SMTPTLSMode: mode, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if err = m.SendVerification(context.Background(), "user@example.com", "secret"); err == nil {
				t.Fatal("accepted untrusted certificate")
			}
		})
	}
}

func TestSMTPAcceptedMessageSurvivesDisconnectBeforeQuit(t *testing.T) {
	for _, mode := range []string{"tls", "starttls"} {
		t.Run(mode, func(t *testing.T) {
			address, result := fakeSMTP(t, mode, trustedCertificate, true)
			m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: address, SMTPUsername: "user", SMTPPassword: "pass", SMTPTLSMode: mode, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if err = m.SendVerification(context.Background(), "user@example.com", "secret"); err != nil {
				t.Fatalf("accepted message reported as undelivered: %v", err)
			}
			if got := <-result; got.err != nil || got.message == "" {
				t.Fatalf("server did not accept message: %+v", got)
			}
		})
	}
}

func TestSMTPCancellationAndTimeoutInterruptBlockedIO(t *testing.T) {
	for _, mode := range []string{"tls", "starttls"} {
		for _, interruption := range []string{"timeout", "parent-deadline", "cancel"} {
			t.Run(fmt.Sprintf("%s/%s", mode, interruption), func(t *testing.T) {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
				accepted := make(chan struct{})
				go func() {
					c, err := listener.Accept()
					if err != nil {
						return
					}
					defer c.Close()
					close(accepted)
					io.Copy(io.Discard, c)
				}()
				timeout := 100 * time.Millisecond
				if interruption == "parent-deadline" {
					timeout = time.Second
				}
				m, err := mailer.New(mailer.Options{Driver: "smtp", From: "sender@example.com", PublicURL: "https://example.com", SMTPAddress: listener.Addr().String(), SMTPTLSMode: mode, Timeout: timeout})
				if err != nil {
					t.Fatal(err)
				}
				var ctx context.Context
				var cancel context.CancelFunc
				if interruption == "parent-deadline" {
					ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
				} else {
					ctx, cancel = context.WithCancel(context.Background())
				}
				defer cancel()
				if interruption == "cancel" {
					go func() { <-accepted; cancel() }()
				}
				start := time.Now()
				err = m.SendVerification(ctx, "user@example.com", "secret")
				want := context.DeadlineExceeded
				if interruption == "cancel" {
					want = context.Canceled
				}
				if !errors.Is(err, want) {
					t.Fatalf("got %v, want %v", err, want)
				}
				if time.Since(start) > time.Second {
					t.Fatal("cancellation did not interrupt blocked connection")
				}
			})
		}
	}
}
