package testenv_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"gotest/tests/testenv"
)

func TestEnvironmentRequirements(t *testing.T) {
	for _, tc := range []struct {
		name, variable, ci, required, value string
		fail, skip                          bool
	}{
		{"optional database", "TEST_DATABASE_URL", "", "", "", false, true},
		{"required database", "TEST_DATABASE_URL", "", "1", "", true, false},
		{"whitespace database", "TEST_DATABASE_URL", "", "1", " \t ", true, false},
		{"CI database", "TEST_DATABASE_URL", "true", "", "", true, false},
		{"numeric CI database", "TEST_DATABASE_URL", "1", "", "", true, false},
		{"configured database", "TEST_DATABASE_URL", "true", "", "postgres://private-test-value", false, false},
		{"optional browser", "CHARACTER_BROWSER_BINARY", "false", "", "", false, true},
		{"required browser", "CHARACTER_BROWSER_BINARY", "", "1", "", true, false},
		{"CI browser", "CHARACTER_BROWSER_BINARY", "true", "", "", true, false},
		{"configured browser", "CHARACTER_BROWSER_BINARY", "true", "", "/test/browser", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary, "-test.run=^TestEnvironmentProbe$", "-test.v")
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				switch strings.ToUpper(key) {
				case "CI", "TEST_DATABASE_URL", "CHARACTER_BROWSER_BINARY", "REQUIRE_DATABASE_TESTS", "REQUIRE_BROWSER_TESTS", "TESTENV_PROBE":
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "CI="+tc.ci, "TESTENV_PROBE="+tc.variable, tc.variable+"="+tc.value,
				"REQUIRE_DATABASE_TESTS="+tc.required, "REQUIRE_BROWSER_TESTS="+tc.required)
			output, err := cmd.CombinedOutput()
			if (err != nil) != tc.fail || strings.Contains(string(output), "--- SKIP:") != tc.skip {
				t.Fatalf("fail=%v skip=%v: %v\n%s", tc.fail, tc.skip, err, output)
			}
			if tc.fail && !strings.Contains(string(output), tc.variable) {
				t.Fatalf("missing actionable setting name: %s", output)
			}
			if strings.Contains(string(output), "private-test-value") {
				t.Fatal("test configuration value leaked")
			}
		})
	}
}

// Child processes exercise real testing.T failure/skip behavior.
func TestEnvironmentProbe(t *testing.T) {
	var value string
	switch os.Getenv("TESTENV_PROBE") {
	case "TEST_DATABASE_URL":
		value = testenv.DatabaseURL(t)
	case "CHARACTER_BROWSER_BINARY":
		value = testenv.Browser(t)
	default:
		return
	}
	if value != strings.TrimSpace(os.Getenv(os.Getenv("TESTENV_PROBE"))) {
		t.Fatal("configured value was not preserved")
	}
}
