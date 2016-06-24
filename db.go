package vain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	verrors "mcquay.me/vain/errors"
	"mcquay.me/vain/metrics"
)

// NewMemDB returns a functional MemDB.
func NewMemDB(p string) (*MemDB, error) {
	m := &MemDB{
		filename: p,

		Users:      map[Email]User{},
		TokToEmail: map[Token]Email{},

		Packages:   map[path]Package{},
		Namespaces: map[namespace]Email{},
	}

	f, err := os.Open(p)
	if err != nil {
		// file doesn't exist yet
		return m, nil
	}
	err = json.NewDecoder(f).Decode(m)
	return m, err
}

// MemDB implements an in-memory, and disk-backed database for a vain server.
type MemDB struct {
	filename string

	l sync.RWMutex

	Users      map[Email]User
	TokToEmail map[Token]Email

	Packages   map[path]Package
	Namespaces map[namespace]Email
}

// NSForToken creates an entry namespaces with a relation to the token.
func (m *MemDB) NSForToken(ns namespace, tok Token) error {
	m.l.Lock()
	defer m.l.Unlock()

	e, ok := m.TokToEmail[tok]
	if !ok {
		return verrors.HTTP{
			Message: fmt.Sprintf("User for token %q not found", tok),
			Code:    http.StatusNotFound,
		}
	}

	if owner, ok := m.Namespaces[ns]; !ok {
		m.Namespaces[ns] = e
	} else {
		if m.Namespaces[ns] != owner {
			return verrors.HTTP{
				Message: fmt.Sprintf("not authorized against namespace %q", ns),
				Code:    http.StatusUnauthorized,
			}
		}
	}
	return m.flush(m.filename)
}

// Package fetches the package associated with path.
func (m *MemDB) Package(pth string) (Package, error) {
	m.l.RLock()
	pkg, ok := m.Packages[path(pth)]
	m.l.RUnlock()
	var err error
	if !ok {
		err = verrors.HTTP{
			Message: fmt.Sprintf("couldn't find package %q", pth),
			Code:    http.StatusNotFound,
		}
	}
	return pkg, err
}

// AddPackage adds p into packages table.
func (m *MemDB) AddPackage(p Package) error {
	m.l.Lock()
	m.Packages[path(p.Path)] = p
	m.l.Unlock()
	return m.flush(m.filename)
}

// RemovePackage removes package with given path
func (m *MemDB) RemovePackage(pth path) error {
	m.l.Lock()
	delete(m.Packages, pth)
	m.l.Unlock()
	return m.flush(m.filename)
}

// PackageExists tells if a package with path is in the database.
func (m *MemDB) PackageExists(pth path) bool {
	m.l.RLock()
	_, ok := m.Packages[path(pth)]
	m.l.RUnlock()
	return ok
}

// Pkgs returns all packages from the database
func (m *MemDB) Pkgs() []Package {
	ps := []Package{}
	m.l.RLock()
	for _, p := range m.Packages {
		ps = append(ps, p)
	}
	m.l.RUnlock()
	return ps
}

// Register adds email to the database, returning an error if there was one.
func (m *MemDB) Register(e Email) (Token, error) {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.Users[e]; ok {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("duplicate email %q", e),
			Code:    http.StatusConflict,
		}
	}

	tok := FreshToken()
	m.Users[e] = User{
		Email:     e,
		token:     tok,
		Requested: time.Now(),
	}
	m.TokToEmail[tok] = e
	return tok, m.flush(m.filename)
}

// Confirm  modifies the user with the given token. Used on register confirmation.
func (m *MemDB) Confirm(tok Token) (Token, error) {
	m.l.Lock()
	defer m.l.Unlock()

	e, ok := m.TokToEmail[tok]
	if !ok {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("bad token: %s", tok),
			Code:    http.StatusNotFound,
		}
	}

	delete(m.TokToEmail, tok)
	tok = FreshToken()
	u, ok := m.Users[e]
	if !ok {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("inconsistent db; found email for token %q, but no user for email %q", tok, e),
			Code:    http.StatusInternalServerError,
		}
	}
	u.token = tok
	m.Users[e] = u
	m.TokToEmail[tok] = e

	return tok, m.flush(m.filename)
}

// Forgot is used fetch a user's token. It implements rudimentary rate
// limiting.
func (m *MemDB) Forgot(e Email, window time.Duration) (Token, error) {
	m.l.Lock()
	defer m.l.Unlock()

	u, ok := m.Users[e]
	if !ok {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("could not find email %q in db", e),
			Code:    http.StatusNotFound,
		}
	}

	if u.Requested.After(time.Now()) {
		return "", verrors.HTTP{
			Message: fmt.Sprintf("rate limit hit for %q; try again in %0.2f mins", u.Email, u.Requested.Sub(time.Now()).Minutes()),
			Code:    http.StatusTooManyRequests,
		}
	}

	return u.token, nil
}

// Sync takes a lock, and flushes the data to disk.
func (m *MemDB) Sync() error {
	m.l.RLock()
	defer m.l.RUnlock()

	return m.flush(m.filename)
}

// flush writes to disk, but expects the user to have taken the lock.
func (m *MemDB) flush(p string) error {
	defer metrics.DBTime("flush")()
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(&m)
}

func (m *MemDB) addUser(e Email) (Token, error) {
	tok := FreshToken()

	m.l.Lock()
	m.Users[e] = User{
		Email:     e,
		token:     tok,
		Requested: time.Now(),
	}
	m.TokToEmail[tok] = e
	m.l.Unlock()

	return tok, m.flush(m.filename)
}

func (m *MemDB) user(e Email) (User, error) {
	m.l.Lock()
	u, ok := m.Users[e]
	m.l.Unlock()
	var err error
	if !ok {
		err = verrors.HTTP{
			Message: fmt.Sprintf("couldn't find user %q", e),
			Code:    http.StatusNotFound,
		}
	}
	return u, err
}
