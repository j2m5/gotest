// Package testenv enforces prerequisites shared by integration test fixtures.
package testenv

import (
	"os"
	"strings"
	"testing"
)

func DatabaseURL(t testing.TB) string {
	t.Helper()
	return setting(t, "TEST_DATABASE_URL", "REQUIRE_DATABASE_TESTS")
}

func Browser(t testing.TB) string {
	t.Helper()
	return setting(t, "CHARACTER_BROWSER_BINARY", "REQUIRE_BROWSER_TESTS")
}

func setting(t testing.TB, name, requirement string) string {
	t.Helper()
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	ci := strings.TrimSpace(os.Getenv("CI"))
	if strings.EqualFold(ci, "true") || ci == "1" || os.Getenv(requirement) == "1" {
		t.Fatalf("%s is required in CI or when %s=1; integration tests must not be skipped", name, requirement)
	}
	t.Skipf("%s is unset; configure it to run this integration test", name)
	return ""
}
