package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckFormat(t *testing.T) {
	t.Run("rejects unformatted source without rewriting it", func(t *testing.T) {
		dir := verificationFixture(t)
		source := "package example\nfunc value()int{return 1}\n"
		writeFixture(t, dir, "example.go", source)
		output, err := runMake(dir, "check-format")
		if err == nil || !strings.Contains(output, "example.go") {
			t.Fatalf("expected formatting failure naming example.go, got %v: %s", err, output)
		}
		if got := readFixture(t, dir, "example.go"); got != source {
			t.Fatal("check-format rewrote the source")
		}
	})

	t.Run("ignores vendored source", func(t *testing.T) {
		dir := verificationFixture(t)
		writeFixture(t, dir, "example.go", "package example\n")
		writeFixture(t, dir, "vendor/example.go", "package example\nfunc value()int{return 1}\n")
		if output, err := runMake(dir, "check-format"); err != nil {
			t.Fatalf("vendored source failed formatting check: %v: %s", err, output)
		}
	})

	t.Run("rejects invalid Go syntax", func(t *testing.T) {
		dir := verificationFixture(t)
		writeFixture(t, dir, "example.go", "package example\nfunc {\n")
		if output, err := runMake(dir, "check-format"); err == nil {
			t.Fatalf("expected syntax failure: %s", output)
		}
	})
}

func TestIntegrationDatabaseConfiguration(t *testing.T) {
	t.Run("requires explicit test database before invoking Go", func(t *testing.T) {
		dir := verificationFixture(t)
		goPath := writeGoRecorder(t, dir, 0)
		output, err := runMake(dir, "test-integration", "GO="+goPath)
		if err == nil || !strings.Contains(output, "MEMOIR_TEST_DATABASE_URL") {
			t.Fatalf("expected missing test database failure, got %v: %s", err, output)
		}
		if _, err := os.Stat(filepath.Join(dir, "go-calls")); !os.IsNotExist(err) {
			t.Fatalf("Go ran without explicit test database: %v", err)
		}
	})

	t.Run("migrates and tests only the explicit test database", func(t *testing.T) {
		dir := verificationFixture(t)
		goPath := writeGoRecorder(t, dir, 0)
		dsn := "postgresql://localhost/memoir_test?sslmode=disable"
		output, err := runMake(dir, "test-integration", "GO="+goPath, "MEMOIR_TEST_DATABASE_URL="+dsn)
		if err != nil {
			t.Fatalf("integration command failed: %v: %s", err, output)
		}
		calls := readFixture(t, dir, "go-calls")
		if strings.Contains(calls, "unrelated-database") || strings.Count(calls, "database="+dsn) != 2 {
			t.Fatalf("integration command did not isolate the database: %s", calls)
		}
		if !strings.Contains(calls, "tool migrate apply --db postgresql --dsn "+dsn) ||
			!strings.Contains(calls, "--migrations internal/database/migrations") ||
			!strings.Contains(calls, "test -mod=vendor -tags=integration -count=1 ./...") {
			t.Fatalf("missing migration or integration invocation: %s", calls)
		}
	})

	t.Run("does not run tests after migration failure", func(t *testing.T) {
		dir := verificationFixture(t)
		goPath := writeGoRecorder(t, dir, 1)
		output, err := runMake(dir, "test-integration", "GO="+goPath,
			"MEMOIR_TEST_DATABASE_URL=postgresql://localhost/memoir_test")
		if err == nil {
			t.Fatalf("expected migration failure: %s", output)
		}
		if calls := readFixture(t, dir, "go-calls"); strings.Contains(calls, "args=test ") {
			t.Fatalf("tests ran after failed migrations: %s", calls)
		}
	})
}

func verificationFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makefile, err := os.ReadFile("../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, dir, "Makefile", string(makefile))
	return dir
}

func writeGoRecorder(t *testing.T, dir string, migrationExit int) string {
	t.Helper()
	script := "#!/bin/sh\nprintf 'database=%s args=%s\\n' \"${DATABASE_URL:-}\" \"$*\" >> go-calls\n"
	if migrationExit != 0 {
		script += "if [ \"$1\" = tool ]; then exit 1; fi\n"
	}
	writeFixture(t, dir, "go-recorder", script)
	path := filepath.Join(dir, "go-recorder")
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func runMake(dir string, args ...string) (string, error) {
	cmd := exec.Command("make", append([]string{"--no-print-directory", "-s"}, args...)...)
	cmd.Dir = dir
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "DATABASE_URL", "MEMOIR_TEST_DATABASE_URL", "MAKEFLAGS", "MFLAGS", "MAKELEVEL", "MAKEOVERRIDES":
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "DATABASE_URL=unrelated-database", "MEMOIR_TEST_DATABASE_URL=")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readFixture(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
