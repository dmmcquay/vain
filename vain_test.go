package vain

import (
	"errors"
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	p := Package{
		Vcs:  "git",
		Path: "mcquay.me/bps",
		Repo: "https://s.mcquay.me/sm/bps",
	}
	got := fmt.Sprintf("%s", p)
	want := `<meta name="go-import" content="mcquay.me/bps git https://s.mcquay.me/sm/bps">`
	if got != want {
		t.Errorf(
			"incorrect converstion to meta; got %s, want %s",
			got,
			want,
		)
	}
}

func TestSupportedVcsStrings(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"hg", true},
		{"git", true},
		{"bzr", true},

		{"", false},
		{"bazar", false},
		{"mercurial", false},
	}
	for _, test := range tests {
		if got, want := valid(test.in), test.want; got != want {
			t.Errorf("%s: %t is incorrect validity", test.in, got)
		}
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		pkgs []Package
		in   string
		want bool
	}{
		{
			pkgs: []Package{},
			in:   "bobo",
			want: true,
		},
		{
			pkgs: []Package{
				{Path: "bobo"},
			},
			in:   "bobo",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "a/b/c"},
			},
			in:   "a/b/c",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "a/b/c"},
			},
			in:   "a/b",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "name/db"},
				{Path: "name/lib"},
			},
			in:   "name/foo",
			want: true,
		},
		{
			pkgs: []Package{
				{Path: "a"},
			},
			in:   "a/b",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "foo"},
			},
			in:   "foo/bar",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "foo/bar"},
				{Path: "foo/baz"},
			},
			in:   "foo",
			want: false,
		},
		{
			pkgs: []Package{
				{Path: "bilbo"},
				{Path: "frodo"},
			},
			in:   "foo/bar/baz",
			want: true,
		},
	}
	for _, test := range tests {
		got := Valid(test.in, test.pkgs)
		if got != test.want {
			t.Errorf("Incorrect testing of %q against %#v; got %t, want %t", test.in, test.pkgs, got, test.want)
		}
	}
}

func TestNamespaceParsing(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   error
	}{
		{
			input: "/sm/foo",
			want:  "sm",
		},
		{
			input: "/a/b/c/d",
			want:  "a",
		},
		{
			input: "/dm/bar",
			want:  "dm",
		},
		{
			input: "/ud",
			want:  "ud",
		},
		// test stripping
		{
			input: "ud",
			want:  "ud",
		},
		{
			input: "/",
			err:   errors.New("should find no namespace"),
		},
		{
			input: "",
			err:   errors.New("should find no namespace"),
		},
	}
	for _, test := range tests {
		got, err := parseNamespace(test.input)
		if err != nil && test.err == nil {
			t.Errorf("unexpected error parsing %q; got %q, want %q, error: %v", test.input, got, test.want, err)
		}

		if got != test.want {
			t.Errorf("parse failure: got %q, want %q", got, test.want)
		}
	}
}
