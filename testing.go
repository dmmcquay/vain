package vain

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestDB returns a populated MemDB in a temp location, as well as a function
// to call at cleanup time.
func TestDB(t *testing.T) (*MemDB, func()) {
	dir, err := ioutil.TempDir("", "vain-testing-")
	if err != nil {
		t.Fatalf("could not create tmpdir for db: %v", err)
		return nil, func() {}
	}
	name := filepath.Join(dir, "test.json")
	db, err := NewMemDB(name)
	if err != nil {
		t.Fatalf("could not create db: %v", err)
		return nil, func() {}
	}
	return db, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("could not clean up tmpdir: %v", err)
		}
	}
}
