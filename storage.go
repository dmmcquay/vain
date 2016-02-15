package vain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Valid checks that p will not confuse the go tool if added to packages.
func Valid(p string, packages []Package) bool {
	for _, pkg := range packages {
		if strings.HasPrefix(pkg.path, p) {
			return false
		}
	}
	return true
}

// Storage is a shim to allow for alternate storage types.
type Storage interface {
	Add(p Package) error
	Remove(path string) error
	All() []Package
}

// SimpleStore implements a simple json on-disk storage.
type SimpleStore struct {
	l sync.RWMutex
	p map[string]Package

	dbl  sync.Mutex
	path string
}

// NewMemStore returns a ready-to-use SimpleStore storing json at path.
func NewMemStore(path string) *SimpleStore {
	return &SimpleStore{
		path: path,
		p:    make(map[string]Package),
	}
}

// Add adds p to the SimpleStore.
func (ms *SimpleStore) Add(p Package) error {
	ms.l.Lock()
	ms.p[p.path] = p
	ms.l.Unlock()
	m := ""
	if err := ms.Save(); err != nil {
		m = fmt.Sprintf("unable to store db: %v", err)
		if err := ms.Remove(p.path); err != nil {
			m = fmt.Sprintf("%s\nto add insult to injury, could not delete package: %v\n", m, err)
		}
		return errors.New(m)
	}
	return nil
}

// Remove removes p from the SimpleStore.
func (ms *SimpleStore) Remove(path string) error {
	ms.l.Lock()
	delete(ms.p, path)
	ms.l.Unlock()
	return nil
}

// All returns all current packages.
func (ms *SimpleStore) All() []Package {
	r := []Package{}
	ms.l.RLock()
	for _, p := range ms.p {
		r = append(r, p)
	}
	ms.l.RUnlock()
	return r
}

// Save writes the db to disk.
func (ms *SimpleStore) Save() error {
	// running in-memory only
	if ms.path == "" {
		return nil
	}
	ms.dbl.Lock()
	defer ms.dbl.Unlock()
	f, err := os.Create(ms.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(ms.p)
}

// Load reads the db from disk and populates ms.
func (ms *SimpleStore) Load() error {
	// running in-memory only
	if ms.path == "" {
		return nil
	}
	ms.dbl.Lock()
	defer ms.dbl.Unlock()
	f, err := os.Open(ms.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&ms.p)
}
