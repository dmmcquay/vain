package vain

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Email is a vain type for storing email addresses.
type Email string

// Token is a vain type for an api token.
type Token string

type namespace string
type path string

var vcss = map[string]bool{
	"hg":  true,
	"git": true,
	"bzr": true,
	"svn": true,
}

func valid(vcs string) bool {
	_, ok := vcss[vcs]
	return ok
}

// Package stores the three pieces of information needed to create the meta
// tag. Two of these (Vcs and Repo) are stored explicitly, and the third is
// determined implicitly by the path POSTed to. For more information refer to
// the documentation for the go tool:
//
// https://golang.org/cmd/go/#hdr-Remote_import_paths
type Package struct {
	//Vcs (version control system) supported: "hg", "git", "bzr", "svn"
	Vcs string `json:"vcs"`
	// Repo: the remote repository url
	Repo string `json:"repo"`

	Path string    `json:"path"`
	Ns   namespace `json:"-"`
}

// User stores the information about a user including email used, their
// token, whether they have registerd and the requested timestamp
type User struct {
	Email      Email
	token      Token
	Registered bool
	Requested  time.Time
}

func (p Package) String() string {
	return fmt.Sprintf(
		"<meta name=\"go-import\" content=\"%s %s %s\">",
		p.Path,
		p.Vcs,
		p.Repo,
	)
}

func splitPathHasPrefix(path, prefix []string) bool {
	if len(path) < len(prefix) {
		return false
	}
	for i, p := range prefix {
		if path[i] != p {
			return false
		}
	}
	return true
}

// Valid checks that p will not confuse the go tool if added to packages.
func Valid(p string, packages []Package) bool {
	ps := strings.Split(p, "/")
	for _, pkg := range packages {
		pre := strings.Split(pkg.Path, "/")
		if splitPathHasPrefix(ps, pre) || splitPathHasPrefix(pre, ps) {
			return false
		}
	}
	return true
}

func parseNamespace(path string) (namespace, error) {
	path = strings.TrimLeft(path, "/")
	if path == "" {
		return "", errors.New("path does not contain namespace")
	}
	elems := strings.Split(path, "/")
	return namespace(elems[0]), nil
}

// FreshToken returns a random token string.
func FreshToken() Token {
	buf := &bytes.Buffer{}
	io.Copy(buf, io.LimitReader(rand.Reader, 6))
	s := hex.EncodeToString(buf.Bytes())
	r := []string{}
	for i := 0; i < len(s)/4; i++ {
		r = append(r, s[i*4:(i+1)*4])
	}
	return Token(strings.Join(r, "-"))
}
