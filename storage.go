package vain

import (
	"errors"
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
}

func NewMemStore() *MemStore {
	return &MemStore{
		p: make(map[string]Package),
	}
}

func (ms MemStore) Add(p Package) error {
	ms.l.Lock()
	ms.p[p.Path] = p
	ms.l.Unlock()
	return nil
}

func (ms MemStore) Remove(path string) error {
	ms.l.Lock()
	delete(ms.p, path)
	ms.l.Unlock()
	return nil
}

func (ms MemStore) Save() error {
	return errors.New("save is not implemented")
}

func (ms MemStore) All() []Package {
	r := []Package{}
	ms.l.RLock()
	for _, p := range ms.p {
		r = append(r, p)
	}
	ms.l.RUnlock()
	return r
}
