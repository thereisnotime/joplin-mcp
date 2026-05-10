package joplin

import "regexp"

// joplinIDPattern matches Joplin's 32-char lowercase-hex item IDs as they
// appear in markdown links and image embeds (`:/<id>`). Used to extract
// outbound note/resource references from a body without parsing markdown.
var joplinIDPattern = regexp.MustCompile(`:/([a-f0-9]{32})`)

// ExtractLinkedIDs returns the unique Joplin item IDs referenced from a note
// body (via `:/<id>` markdown links / image embeds). Order is the order of
// first occurrence; duplicates are dropped.
func ExtractLinkedIDs(body string) []string {
	matches := joplinIDPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		id := m[1]
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
