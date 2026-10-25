package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

type smtpTransport struct {
	address, host, username, password, mode string
	tlsConfig                               *tls.Config
}

func newSMTPTransport(o Options) (*smtpTransport, error) {
	host, port, err := net.SplitHostPort(o.SMTPAddress)
	if err != nil || host == "" || strings.ContainsAny(host, "\r\n\t ") {
		return nil, errors.New("SMTP address must contain host and port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return nil, errors.New("invalid SMTP port")
	}
	mode := o.SMTPTLSMode
	if mode == "" {
		mode = "starttls"
	}
	if mode != "tls" && mode != "starttls" {
		return nil, errors.New("SMTP TLS mode must be tls or starttls")
	}
	if o.SMTPUsername == "" && o.SMTPPassword != "" {
		return nil, errors.New("SMTP password requires username")
	}
	return &smtpTransport{address: o.SMTPAddress, host: host, username: o.SMTPUsername, password: o.SMTPPassword, mode: mode, tlsConfig: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}, nil
}

func (t *smtpTransport) send(ctx context.Context, from, to string, message []byte) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", t.address)
	if err != nil {
		return err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		if err = conn.SetDeadline(deadline); err != nil {
			return err
		}
	}
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	active := conn
	if t.mode == "tls" {
		encrypted := tls.Client(conn, t.tlsConfig)
		if err = encrypted.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("SMTP TLS handshake: %w", err)
		}
		active = encrypted
	}
	client, err := smtp.NewClient(active, t.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if t.mode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("SMTP server does not support required STARTTLS")
		}
		if err = client.StartTLS(t.tlsConfig); err != nil {
			return fmt.Errorf("SMTP STARTTLS: %w", err)
		}
	}
	if t.username != "" {
		if err = client.Auth(smtp.PlainAuth("", t.username, t.password, t.host)); err != nil {
			return fmt.Errorf("SMTP authentication: %w", err)
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(message); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	// DATA was accepted: a QUIT failure must not turn successful delivery into
	// a retry that replaces the confirmation link the server already accepted.
	_ = client.Quit()
	return nil
}
