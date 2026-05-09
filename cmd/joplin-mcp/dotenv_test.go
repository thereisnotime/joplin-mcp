package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := `# a comment
JOPLIN_TOKEN=abc123
JOPLIN_BASE_URL="http://example:1234"
QUOTED_SINGLE='hello world'
NOEQ
EMPTY=

# preserve-existing key:
ALREADY_SET=should-not-overwrite
`
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ALREADY_SET", "from-shell")
	t.Setenv("JOPLIN_TOKEN", "")
	_ = os.Unsetenv("JOPLIN_TOKEN")
	_ = os.Unsetenv("JOPLIN_BASE_URL")
	_ = os.Unsetenv("QUOTED_SINGLE")
	_ = os.Unsetenv("EMPTY")

	if err := loadDotEnv(envPath); err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}

	cases := map[string]string{
		"JOPLIN_TOKEN":    "abc123",
		"JOPLIN_BASE_URL": "http://example:1234",
		"QUOTED_SINGLE":   "hello world",
		"EMPTY":           "",
		"ALREADY_SET":     "from-shell",
	}
	for k, want := range cases {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestLoadDotEnv_MissingIsOK(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), "does-not-exist")); err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}
