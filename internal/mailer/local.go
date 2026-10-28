package mailer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"strings"
)

type localTransport struct{ directory string }

func (t *localTransport) send(ctx context.Context, _, _ string, message []byte) (err error) {
	if err = ctx.Err(); err != nil {
		return err
	}
	preview, err := localPreview(message)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(t.directory, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(t.directory, "verification-*.eml")
	if err != nil {
		return err
	}
	var previewFile *os.File
	defer func() {
		file.Close()
		if previewFile != nil {
			previewFile.Close()
		}
		if err != nil {
			os.Remove(file.Name())
			if previewFile != nil {
				os.Remove(previewFile.Name())
			}
		}
	}()
	if err = ctx.Err(); err != nil {
		return err
	}
	if _, err = file.Write(message); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	previewFile, err = os.OpenFile(strings.TrimSuffix(file.Name(), ".eml")+".txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = previewFile.Write(preview); err != nil {
		return err
	}
	if err = previewFile.Close(); err != nil {
		return err
	}
	return ctx.Err()
}

// The MIME file remains suitable for mail clients; the text copy lets developers
// copy a complete link without interpreting quoted-printable escapes and wraps.
func localPreview(message []byte) ([]byte, error) {
	letter, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil || letter.Header.Get("Content-Transfer-Encoding") != "quoted-printable" {
		return nil, errors.New("invalid local mail preview format")
	}
	body, err := io.ReadAll(quotedprintable.NewReader(letter.Body))
	if err != nil {
		return nil, errors.New("decode local mail preview")
	}
	var preview bytes.Buffer
	for _, header := range []string{"From", "To", "Subject", "Date"} {
		fmt.Fprintf(&preview, "%s: %s\r\n", header, letter.Header.Get(header))
	}
	preview.WriteString("\r\n")
	preview.Write(body)
	return preview.Bytes(), nil
}
