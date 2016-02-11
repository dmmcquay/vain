/*

From the documentation for the go tool, it searches for the following header
when searching for packages:

<meta name="go-import" content="import-prefix vcs repo-root">

this is simply a service for aggregating a collection of prefix, vcs, and
repo-root tuples, and serving the appropriate header over http.

API

In order to add a new package POST a json object to the following route:

POST /v0/package/

A sample json object:

	{
		"path": "mcquay.me/vain",
		"repo": "https://s.mcquay.me/sm/vain"
	}

Naming

the "ysv" in ysvd stands for You're so Vain, the song by Carly Simon.
*/
package vain

import "fmt"

type vcs int

const (
	Git vcs = iota
	Hg
)

var vcss = [...]string{
	"git",
	"mercurial",
}

var labelToVcs = map[string]vcs{
	"git":       Git,
	"mercurial": Hg,
	"hg":        Hg,
}

// String returns the name of the vcs ("git", "mercurial", ...).
func (v vcs) String() string { return vcss[v] }

type Package struct {
	Vcs  vcs    `json":vcs"`
	Path string `json:"path"`
	Repo string `json:"repo"`
}

func (p Package) String() string {
	return fmt.Sprintf(
		"<meta name=\"go-import\" content=\"%s %s %s\">",
		p.Path,
		p.Vcs,
		p.Repo,
	)
}
