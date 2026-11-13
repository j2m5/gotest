package db_test

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCombatLogCleanupCommand(t *testing.T) {
	f := cloneFinishedLog(t, wonLogFixture(t), 40)
	binary := filepath.Join(t.TempDir(), "cleanup.exe")
	build := exec.CommandContext(f.ctx, "go", "build", "-o", binary, "./cmd/cleanup")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build cleanup: %v\n%s", err, output)
	}
	directory := t.TempDir()
	for _, dryRun := range []bool{true, false} {
		args := []string{"-battle-logs", "-batch-size", "2"}
		if dryRun {
			args = append(args, "-dry-run")
		}
		command := exec.CommandContext(f.ctx, binary, args...)
		command.Dir = directory
		command.Env = pathAppEnvironment(map[string]string{"DATABASE_URL": isolatedCombatDSN(t, f.s)})
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("cleanup command: %v\n%s", err, output)
		}
		var event map[string]any
		if err := json.Unmarshal(output, &event); err != nil {
			t.Fatal("invalid CLI JSON", err)
		}
		if event["battle_log_cutoff"] == nil {
			t.Fatal("missing retention cutoff", event)
		}
		if dryRun {
			if event["msg"] != "cleanup_preview" || event["battle_logs_matched"] != float64(1) || event["battle_events_matched"] != float64(2) {
				t.Fatal(event)
			}
			logCounts(t, f, 3, 0)
		} else {
			if event["msg"] != "cleanup_completed" || event["battle_logs_cleared"] != float64(1) || event["battle_events_deleted"] != float64(2) {
				t.Fatal(event)
			}
			logCounts(t, f, 1, 1)
		}
	}
}
