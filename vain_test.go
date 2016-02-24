package vain

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	p := Package{
		Vcs:  "git",
		path: "mcquay.me/bps",
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
				{path: ""},
			},
			in:   "bobo",
			want: true,
		},
		{
			pkgs: []Package{
				{path: "bobo"},
			},
			in:   "bobo",
			want: false,
		},
		{
			pkgs: []Package{
				{path: "a/b/c"},
			},
			in:   "a/b/c",
			want: false,
		},
		{
			pkgs: []Package{
				{path: "foo/bar"},
				{path: "foo/baz"},
			},
			in:   "foo",
			want: false,
		},
		{
			pkgs: []Package{
				{path: "bilbo"},
				{path: "frodo"},
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
