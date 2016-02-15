package vain

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

func Valid(p string, packages []Package) bool {
	for _, pkg := range packages {
		if strings.HasPrefix(pkg.Path, p) {
			return false
		}
	}
	return true
}

type MemStore struct {
	l sync.RWMutex
	p map[string]Package

	dbl  sync.Mutex
	path string
}

func NewMemStore(path string) *MemStore {
	return &MemStore{
		path: path,
		p:    make(map[string]Package),
	}
}

func (ms *MemStore) Add(p Package) error {
	ms.l.Lock()
	ms.p[p.Path] = p
	ms.l.Unlock()
	return nil
}

func (ms *MemStore) Remove(path string) error {
	ms.l.Lock()
	delete(ms.p, path)
	ms.l.Unlock()
	return nil
}

func (ms *MemStore) Save() error {
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

func (ms *MemStore) Load() error {
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

func (ms *MemStore) All() []Package {
	r := []Package{}
	ms.l.RLock()
	for _, p := range ms.p {
		r = append(r, p)
	}
	ms.l.RUnlock()
	return r
}
