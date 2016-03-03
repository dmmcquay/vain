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
		if strings.HasPrefix(pkg.path, p) || strings.HasPrefix(p, pkg.path) {
			return false
		}
	}
	return true
}

// Storage is a shim to allow for alternate storage types.
type Storage interface {
	Contains(name string) bool
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

// NewSimpleStore returns a ready-to-use SimpleStore storing json at path.
func NewSimpleStore(path string) *SimpleStore {
	return &SimpleStore{
		path: path,
		p:    make(map[string]Package),
	}
}

func (ms *SimpleStore) Contains(name string) bool {
	_, contains := ms.p[name]
	return contains
}

// Add adds p to the SimpleStore.
func (ss *SimpleStore) Add(p Package) error {
	ss.l.Lock()
	ss.p[p.path] = p
	ss.l.Unlock()
	m := ""
	if err := ss.Save(); err != nil {
		m = fmt.Sprintf("unable to store db: %v", err)
		if err := ss.Remove(p.path); err != nil {
			m = fmt.Sprintf("%s\nto add insult to injury, could not delete package: %v\n", m, err)
		}
		return errors.New(m)
	}
	return nil
}

// Remove removes p from the SimpleStore.
func (ss *SimpleStore) Remove(path string) error {
	ss.l.Lock()
	delete(ss.p, path)
	ss.l.Unlock()
	return nil
}

// All returns all current packages.
func (ss *SimpleStore) All() []Package {
	r := []Package{}
	ss.l.RLock()
	for _, p := range ss.p {
		r = append(r, p)
	}
	ss.l.RUnlock()
	return r
}

// Save writes the db to disk.
func (ss *SimpleStore) Save() error {
	// running in-memory only
	if ss.path == "" {
		return nil
	}
	ss.dbl.Lock()
	defer ss.dbl.Unlock()
	f, err := os.Create(ss.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(ss.p)
}

// Load reads the db from disk and populates ss.
func (ss *SimpleStore) Load() error {
	// running in-memory only
	if ss.path == "" {
		return nil
	}
	ss.dbl.Lock()
	defer ss.dbl.Unlock()
	f, err := os.Open(ss.path)
	if err != nil {
		return err
	}
	defer f.Close()

	in := map[string]Package{}
	if err := json.NewDecoder(f).Decode(&in); err != nil {
		return err
	}
	for k, v := range in {
		v.path = k
		ss.p[k] = v
	}
	return nil
}
