// Package vain implements a vanity service for use by the the go tool.
//
// The executable, cmd/vaind, is located in the respective subdirectory.
package vain

import "fmt"

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
