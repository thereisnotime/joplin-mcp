package joplin

import (
	"reflect"
	"testing"
)

func TestExtractLinkedIDs(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"none", "no links here", nil},
		{"image", "![alt](:/abcdef0123456789abcdef0123456789)", []string{"abcdef0123456789abcdef0123456789"}},
		{"link", "see [doc](:/0123456789abcdef0123456789abcdef)", []string{"0123456789abcdef0123456789abcdef"}},
		{
			"multiple, deduped",
			"first :/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa then :/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb and again :/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			[]string{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			},
		},
		{"too short skipped", ":/short", nil},
		{"uppercase skipped (Joplin always lowercases)", ":/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractLinkedIDs(tc.body)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
