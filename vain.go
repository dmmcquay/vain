// Package vain implements a vanity service for use by the the go tool.
//
// The executable, cmd/vaind, is located in the respective subdirectory.
package vain

import (
	"encoding/json"
	"fmt"
)

type vcs int

const (
	// Git is the default Vcs.
	Git vcs = iota

	// Hg is mercurial
	Hg

	// Svn
	Svn

	// Bazaar
	Bzr
)

var vcss = [...]string{
	"git",
	"mercurial",
	"svn",
	"bazaar",
}

var labelToVcs = map[string]vcs{
	"git":       Git,
	"mercurial": Hg,
	"hg":        Hg,
	"svn":       Svn,
	"bazaar":    Bzr,
}

// String returns the name of the vcs ("git", "mercurial", ...).
func (v vcs) String() string { return vcss[v] }

// Package stores the three pieces of information needed to create the meta
// tag. Two of these (Vcs and Repo) are stored explicitly, and the third is
// determined implicitly by the path POSTed to. For more information refer to
// the documentation for the go tool:
//
// https://golang.org/cmd/go/#hdr-Remote_import_paths
type Package struct {
	//Vcs (version control system) supported: "git", "mercurial"
	Vcs vcs `json:"vcs"`
	// Repo: the remote repository url
	Repo string `json:"repo"`

	path string
}

func (p Package) String() string {
	return fmt.Sprintf(
		"<meta name=\"go-import\" content=\"%s %s %s\">",
		p.path,
		p.Vcs,
		p.Repo,
	)
}

func (p *Package) UnmarshalJSON(b []byte) (err error) {
	pkg := struct {
		Vcs  string
		Repo string
	}{}
	err = json.Unmarshal(b, &pkg)
	if err != nil {
		return err
	}
	p.Vcs, p.Repo = labelToVcs[pkg.Vcs], pkg.Repo
	return nil
}
