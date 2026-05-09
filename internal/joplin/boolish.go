package joplin

import (
	"fmt"
	"strings"
)

// Joplin's SQLite-backed REST API returns boolean columns as integers (0/1),
// not JSON booleans. Boolish accepts either form on the wire and presents
// itself as a normal bool to Go code.
type Boolish bool

func (b *Boolish) UnmarshalJSON(data []byte) error {
	switch s := strings.TrimSpace(string(data)); s {
	case "true", "1":
		*b = true
	case "false", "0", "null", `""`:
		*b = false
	default:
		return fmt.Errorf("joplin: cannot unmarshal %s as boolean", s)
	}
	return nil
}

func (b Boolish) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (b Boolish) Bool() bool { return bool(b) }
