package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	got := String()
	if !strings.HasPrefix(got, "joplin-mcp ") {
		t.Errorf("String() = %q, want prefix %q", got, "joplin-mcp ")
	}
	for _, want := range []string{Version, Commit, Date} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, missing %q", got, want)
		}
	}
}
