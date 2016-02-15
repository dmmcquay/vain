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
	Vcs  vcs    `json:"vcs"`
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
