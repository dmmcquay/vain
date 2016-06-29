package vain

import (
	"errors"
	"testing"
)

func TestPartialPackage(t *testing.T) {
	db, done := TestDB(t)
	if db == nil {
		t.Fatalf("could not create temp db")
	}
	defer done()

	paths := []path{
		"a/b",
		"a/c",
		"a/d/c",
		"a/d/e",

		"f/b/c/d",
		"f/b/c/e",
	}

	for _, p := range paths {
		db.Packages[p] = Package{Path: string(p)}
	}

	tests := []struct {
		pth string
		pkg Package
		err error
	}{
		// obvious
		{
			pth: "a/b",
			pkg: db.Packages["a/b"],
		},
		{
			pth: "a/d/c",
			pkg: db.Packages["a/d/c"],
		},

		// here we exercise the code that matches closest submatch
		{
			pth: "a/b/c",
			pkg: db.Packages["a/b"],
		},
		{
			pth: "f/b/c/d/e/f/g",
			pkg: db.Packages["f/b/c/d"],
		},

		// some errors
		{
			pth: "foo",
			err: errors.New("shouldn't find"),
		},

		{
			pth: "a/d/f",
			err: errors.New("shouldn't find"),
		},
	}

	for _, test := range tests {
		p, err := db.Package(test.pth)

		if got, want := p, test.pkg; got != want {
			t.Errorf("bad package fetched: got %+v, want %+v", got, want)
		}

		got := err
		want := test.err
		if (got == nil) != (want == nil) {
			t.Errorf("unexpected error; got %v, want %v", got, want)
		}
	}

}
