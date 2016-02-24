package vain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func equiv(a, b map[string]Package) (bool, []error) {
	equiv := true
	errs := []error{}
	if got, want := len(a), len(b); got != want {
		equiv = false
		errs = append(errs, fmt.Errorf("incorrect number of elements: got %d, want %d", got, want))
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
		"foo": {Vcs: "mercurial"},
		"bar": {Vcs: "bzr"},
		"baz": {},
	}
	for k, v := range orig {
		v.path = k
		orig[k] = v
	}
	ms.p = orig
	if err := ms.Save(); err != nil {
		t.Errorf("should have been able to Save: %v", err)
	}
	ms.p = map[string]Package{}
	if err := ms.Load(); err != nil {
		t.Errorf("should have been able to Load: %v", err)
	}
	if ok, errs := equiv(ms.p, orig); !ok {
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

func TestPackageJsonParsing(t *testing.T) {
	tests := []struct {
		input  string
		output string
		parsed Package
	}{
		{
			input:  `{"vcs":"git","repo":"https://s.mcquay.me/sm/ud/"}`,
			output: `{"vcs":"git","repo":"https://s.mcquay.me/sm/ud/"}`,
			parsed: Package{Vcs: "git", Repo: "https://s.mcquay.me/sm/ud/"},
		},
		{
			input:  `{"vcs":"hg","repo":"https://s.mcquay.me/sm/ud/"}`,
			output: `{"vcs":"hg","repo":"https://s.mcquay.me/sm/ud/"}`,
			parsed: Package{Vcs: "hg", Repo: "https://s.mcquay.me/sm/ud/"},
		},
	}

	for _, test := range tests {
		p := Package{}
		if err := json.NewDecoder(strings.NewReader(test.input)).Decode(&p); err != nil {
			t.Error(err)
		}
		if p != test.parsed {
			t.Errorf("got:\n\t%v, want\n\t%v", p, test.parsed)
		}
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(&p); err != nil {
			t.Error(err)
		}
		if got, want := strings.TrimSpace(buf.String()), test.output; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
