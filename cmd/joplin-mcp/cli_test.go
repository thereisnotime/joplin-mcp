package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadJSONSource_Literal(t *testing.T) {
	got, err := readJSONSource(`{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"a":1}` {
		t.Errorf("got %q", got)
	}
}

func TestReadJSONSource_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.json")
	if err := os.WriteFile(path, []byte(`{"hi":"there"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readJSONSource("@" + path)
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"hi":"there"}` {
		t.Errorf("got %q", got)
	}
}

func TestReadJSONSource_FileMissing(t *testing.T) {
	_, err := readJSONSource("@/nonexistent/joplin-mcp-test")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read /nonexistent") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReadJSONSource_Stdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = saved })
	go func() {
		_, _ = w.Write([]byte(`{"from":"stdin"}`))
		_ = w.Close()
	}()
	got, err := readJSONSource("-")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"from":"stdin"}` {
		t.Errorf("got %q", got)
	}
}
