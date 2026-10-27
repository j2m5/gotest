// Package mailer delivers account emails through local files or encrypted SMTP.
package mailer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode"
)

type Options struct {
	Driver, From, PublicURL, Directory, SMTPAddress, SMTPUsername, SMTPPassword, SMTPTLSMode string
	Timeout                                                                                  time.Duration
}

type transport interface {
	send(context.Context, string, string, []byte) error
}

type Mailer struct {
	from      string
	publicURL *url.URL
	transport transport
	timeout   time.Duration
}

func New(o Options) (*Mailer, error) {
	if !validAddress(o.From) {
		return nil, errors.New("invalid mail sender")
	}
	u, err := url.Parse(o.PublicURL)
	if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.IndexFunc(o.PublicURL, unicode.IsControl) >= 0 {
		return nil, errors.New("invalid public URL")
	}
	if o.Timeout < 0 {
		return nil, errors.New("invalid mail timeout")
	}
	if o.Timeout == 0 {
		o.Timeout = 10 * time.Second
	}
	m := &Mailer{from: o.From, publicURL: u, timeout: o.Timeout}
	switch o.Driver {
	case "local":
		if o.Directory == "" {
			o.Directory = ".local/mail"
		}
		m.transport = &localTransport{directory: o.Directory}
	case "smtp":
		smtp, err := newSMTPTransport(o)
		if err != nil {
			return nil, err
		}
		m.transport = smtp
	default:
		return nil, errors.New("mail driver must be local or smtp")
	}
	return m, nil
}

func validAddress(value string) bool {
	if value == "" || len(value) > 254 || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return false
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Name == "" && address.Address == value && strings.Contains(value, "@")
}

func (m *Mailer) SendVerification(ctx context.Context, email, token string) error {
	return m.send(ctx, email, "Verify your email address", "deliver verification email", func() (string, error) {
		if token == "" {
			return "", errors.New("empty verification token")
		}
		return fmt.Sprintf("Please confirm your email address by opening this link:\r\n\r\n%s\r\n", m.link("/verify", token)), nil
	})
}

// SendPasswordReset sends a single-use password reset link to email.
func (m *Mailer) SendPasswordReset(ctx context.Context, email, token string) error {
	return m.send(ctx, email, "Reset your password", "deliver password reset email", func() (string, error) {
		if token == "" {
			return "", errors.New("empty password reset token")
		}
		return fmt.Sprintf("Use this single-use link to reset your password:\r\n\r\n%s\r\n\r\nIf you did not request a password reset, you can ignore this email.\r\n", m.link("/reset-password", token)), nil
	})
}

// SendPasswordChanged notifies email that the account password was changed.
func (m *Mailer) SendPasswordChanged(ctx context.Context, email string) error {
	return m.send(ctx, email, "Your password was changed", "deliver password changed email", func() (string, error) {
		return "Your password was changed. All sessions have been closed. Please sign in again.\r\n\r\nIf you did not make this change, contact support.\r\n", nil
	})
}

func (m *Mailer) link(path, token string) string {
	link := *m.publicURL
	link.Path = path
	link.RawPath = ""
	link.RawQuery = url.Values{"token": {token}}.Encode()
	return link.String()
}

func (m *Mailer) send(ctx context.Context, email, subject, failure string, render func() (string, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validAddress(email) {
		return errors.New("invalid mail recipient")
	}
	body, err := render()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return fmt.Errorf("message identifier: %w", err)
	}
	var message bytes.Buffer
	fmt.Fprintf(&message, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s@%s>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", m.from, email, subject, time.Now().UTC().Format(time.RFC1123Z), hex.EncodeToString(id), m.publicURL.Hostname())
	writer := quotedprintable.NewWriter(&message)
	if _, err := writer.Write([]byte(body)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := m.transport.send(ctx, m.from, email, message.Bytes()); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// The socket deadline can fire before the context timer sets ctx.Err().
		// Only normalize a network timeout once the shared deadline has elapsed.
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
				return context.DeadlineExceeded
			}
		}
		return fmt.Errorf("%s: %w", failure, err)
	}
	return nil
}
