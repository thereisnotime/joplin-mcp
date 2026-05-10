package joplin

import (
	"encoding/json"
	"testing"
)

func TestBoolish_Unmarshal(t *testing.T) {
	cases := map[string]struct {
		want  bool
		errOK bool
	}{
		"true":  {want: true},
		"false": {want: false},
		"1":     {want: true},
		"0":     {want: false},
		"null":  {want: false},
		`""`:    {want: false},
		`"yes"`: {errOK: true},
		"42":    {errOK: true},
	}
	for in, tc := range cases {
		var b Boolish
		err := json.Unmarshal([]byte(in), &b)
		if tc.errOK {
			if err == nil {
				t.Errorf("input %s: expected error, got nil", in)
			}
			continue
		}
		if err != nil {
			t.Errorf("input %s: %v", in, err)
			continue
		}
		if bool(b) != tc.want {
			t.Errorf("input %s: got %v, want %v", in, bool(b), tc.want)
		}
	}
}

func TestBoolish_Marshal(t *testing.T) {
	for _, tc := range []struct {
		v    Boolish
		want string
	}{
		{true, "true"},
		{false, "false"},
	} {
		got, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != tc.want {
			t.Errorf("Marshal(%v) = %s, want %s", bool(tc.v), got, tc.want)
		}
	}
}

func TestBoolish_Bool(t *testing.T) {
	if Boolish(true).Bool() != true || Boolish(false).Bool() != false {
		t.Error("Bool() round-trip broken")
	}
}
