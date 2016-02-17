package vain

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	p := Package{
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

func TestVcsStrings(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{fmt.Sprintf("%+v", Git), "git"},
		{fmt.Sprintf("%+v", Hg), "mercurial"},
		{fmt.Sprintf("%+v", Svn), "svn"},
		{fmt.Sprintf("%+v", Bzr), "bazaar"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("incorrect conversion of vain.Vcs -> string; got %q, want %q", test.got, test.want)
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
