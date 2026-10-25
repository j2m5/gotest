package mailer

import (
	"context"
	"os"
)

type localTransport struct{ directory string }

func (t *localTransport) send(ctx context.Context, _, _ string, message []byte) (err error) {
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = os.MkdirAll(t.directory, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(t.directory, "verification-*.eml")
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(file.Name())
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
	return ctx.Err()
}
