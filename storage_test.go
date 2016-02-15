package vain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func equiv(a, b map[string]Package) (bool, []error) {
	equiv := true
	errs := []error{}
	if got, want := len(a), len(b); got != want {
		equiv = false
		errs = append(errs, fmt.Errorf("uncorrect number of elements: got %d, want %d", got, want))
		return false, errs
	}

	for k := range a {
		v, ok := b[k]
		if !ok || v != a[k] {
			errs = append(errs, fmt.Errorf("missing key: %s", k))
			equiv = false
			break
		}
	}
	return equiv, errs
}

func TestSimpleStorage(t *testing.T) {
	root, err := ioutil.TempDir("", "vain-")
	if err != nil {
		t.Fatalf("problem creating temp dir: %v", err)
	}
	defer func() { os.RemoveAll(root) }()
	db := filepath.Join(root, "vain.json")
	ms := NewSimpleStore(db)
	orig := map[string]Package{
		"foo": {},
		"bar": {},
		"baz": {},
	}
	ms.p = orig
	if err := ms.Save(); err != nil {
		t.Errorf("should have been able to Save: %v", err)
	}
	ms.p = map[string]Package{}
	if err := ms.Load(); err != nil {
		t.Errorf("should have been able to Load: %v", err)
	}
	if ok, errs := equiv(orig, ms.p); !ok {
		for _, err := range errs {
			t.Error(err)
		}
	}
}

func TestRemove(t *testing.T) {
	root, err := ioutil.TempDir("", "vain-")
	if err != nil {
		t.Fatalf("problem creating temp dir: %v", err)
	}
	defer func() { os.RemoveAll(root) }()
	db := filepath.Join(root, "vain.json")
	ms := NewSimpleStore(db)
	ms.p = map[string]Package{
		"foo": {},
		"bar": {},
		"baz": {},
	}
	if err := ms.Remove("foo"); err != nil {
		t.Errorf("unexpected error during remove: %v", err)
	}
	want := map[string]Package{
		"bar": {},
		"baz": {},
	}
	if ok, errs := equiv(ms.p, want); !ok {
		for _, err := range errs {
			t.Error(err)
		}
	}
}
