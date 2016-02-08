/*

From the documentation for the go tool, it searches for the following header
when searching for packages:

<meta name="go-import" content="import-prefix vcs repo-root">

ysv stands for You're so Vain, the song by Carly Simon.
*/
package ysv

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

type pkg struct {
	Vcs  vcs    `json":vcs"`
	Path string `json:"path"`
	Repo string `json:"repo"`
}

func (p pkg) String() string {
	return fmt.Sprintf(
		"<meta name=\"go-import\" content=\"%s %s %s\">",
		p.Path,
		p.Vcs,
		p.Repo,
	)
}
