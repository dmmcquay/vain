package vain

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func testDB(t *testing.T) (*DB, func()) {
	dir, err := ioutil.TempDir("", "vain-testing-")
	if err != nil {
		t.Fatalf("could not create tmpdir for db: %v", err)
		return nil, func() {}
	}
	name := filepath.Join(dir, "test.db")
	db, err := NewDB(name)
	if err != nil {
		t.Fatalf("could not create db: %v", err)
		return nil, func() {}
	}

	if err := db.Init(); err != nil {
		return nil, func() {}
	}

	return db, func() {
		db.Close()
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("could not clean up tmpdir: %v", err)
		}
	}
}
