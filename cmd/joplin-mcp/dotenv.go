package main

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"strings"
)

// loadDotEnv reads a .env file in cwd if one exists and sets each KEY=VALUE
// pair as an environment variable. Existing env vars win (so the shell can
// always override the file).
//
// Format is the common subset: blank lines and lines starting with '#' are
// ignored, lines are KEY=VALUE, surrounding single or double quotes on the
// value are stripped. No multiline values, no shell expansion, no exports.
func loadDotEnv(path string) error {
	// #nosec G304,G703 -- path is operator-supplied via --env-file or the documented XDG fallback chain
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return s.Err()
}
