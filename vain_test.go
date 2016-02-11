package vain

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	p := Package{
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

func TestVcsStrings(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{fmt.Sprintf("%+v", Git), "git"},
		{fmt.Sprintf("%+v", Hg), "mercurial"},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("incorrect conversion of vain.Vcs -> string; got %q, want %q", test.got, test.want)
		}
	}
}
