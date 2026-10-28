package backup_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gotest/internal/backup"
	"gotest/internal/notification"
)

func TestRestoreRejectsIncompleteAndInvalidManifestBeforeConnection(t *testing.T) {
	key := []byte("01234567890123456789012345678901")
	cipher, _ := notification.NewCipher(key)
	archive := []byte("fixture archive")
	sum := sha256.Sum256(archive)
	good := backup.Manifest{FormatVersion: 1, CreatedAt: time.Now().UTC(), SourceDatabase: "source", MigrationVersion: 1, KeyID: cipher.KeyID(), ArchiveSHA256: hex.EncodeToString(sum[:])}
	encoded := func(m backup.Manifest) []byte { data, _ := json.Marshal(m); return data }
	cases := []struct {
		name string
		data []byte
	}{
		{"absent", nil}, {"truncated", []byte(`{"format_version":1`)}, {"unknown fields", []byte(strings.TrimSuffix(string(encoded(good)), "}") + `,"extra":true}`)},
		{"trailing value", append(encoded(good), []byte(`{}`)...)}, {"oversized trailing content", append(append(encoded(good), []byte(strings.Repeat(" ", 65536))...), []byte(`{}`)...)},
	}
	for _, change := range []struct {
		name   string
		modify func(*backup.Manifest)
	}{
		{"version", func(m *backup.Manifest) { m.FormatVersion = 2 }}, {"date", func(m *backup.Manifest) { m.CreatedAt = time.Time{} }},
		{"source", func(m *backup.Manifest) { m.SourceDatabase = "" }}, {"migration", func(m *backup.Manifest) { m.MigrationVersion = 0 }},
		{"key", func(m *backup.Manifest) { m.KeyID = "private-key" }}, {"hash", func(m *backup.Manifest) { m.ArchiveSHA256 = "not-hex" }},
	} {
		m := good
		change.modify(&m)
		cases = append(cases, struct {
			name string
			data []byte
		}{change.name, encoded(m)})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.data != nil {
				dbFail(t, "write manifest fixture", os.WriteFile(filepath.Join(dir, "manifest.json"), tc.data, 0600))
			}
			dbFail(t, "write archive fixture", os.WriteFile(filepath.Join(dir, "database.dump"), archive, 0600))
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := backup.Restore(ctx, backup.Options{DatabaseURL: "postgres://user@127.0.0.1:1/target", Directory: dir, Key: key})
			var problem *backup.Error
			if !errors.As(err, &problem) || problem.Stage != "manifest" {
				t.Fatal("invalid manifest reached connection stage")
			}
			if strings.Contains(err.Error(), "private-key") {
				t.Fatal("manifest values exposed in diagnostics")
			}
		})
	}
}
