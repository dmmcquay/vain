/*
ysvd, a webserver for hosting go get vanity urls.

The go get command searches for the following header when searching for
packages:

	<meta name="go-import" content="import-prefix vcs repo-root">

this is simply a service for aggregating a collection of prefix, vcs, and
repo-root tuples, and serving the appropriate header over http. For more
information please refer to the documentation for the go tool found at
https://golang.org/cmd/go/#hdr-Remote_import_paths

API

Assume an instance of ysvd at example.org. In order to add a package
example.org/foo that points at bitbucket.org/example/foo (a mercurial
repository) POST the following json object:

	{
		"vcs": "mercurial",
		"repo": "https://bitbucket.org/example/foo"
	}

to https://example.org/foo.

Doing so, then visiting https://example.org/foo?go-get=1 will yield a header
that looks like:


	<meta name="go-import" content="example.org/foo hg https://bitbucket.org/foo">


The json object sent to server can have two fields: "repo" and "vcs". "repo" is
required; leaving off the "vcs" member defaults to "git".

Naming

The "ysv" in ysvd stands for You're so Vain, the song by Carly Simon.
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mcquay.me/vain"

	"github.com/kelseyhightower/envconfig"
)

const usage = `ysvd

environment vars:

YSV_PORT: tcp listen port
YSV_HOST: hostname to use
YSV_DB: path to json database
`

type config struct {
	Port int
	Host string
	DB   string
}

func main() {
	c := &config{
		Port: 4040,
	}
	if err := envconfig.Process("ysv", c); err != nil {
		fmt.Fprintf(os.Stderr, "problem processing environment: %v", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "env", "e", "help", "h":
			fmt.Printf("%s\n", usage)
			os.Exit(0)
		}
	}
	if c.Host == "" {
		log.Printf("must set YSV_HOST; please run $(ysvd env) for more information")
		os.Exit(1)
	}
	if c.DB == "" {
		log.Printf("warning: in-memory db mode; if you do not want this set YSV_DB")
	}
	hostname := "localhost"
	if hn, err := os.Hostname(); err != nil {
		log.Printf("problem getting hostname:", err)
	} else {
		hostname = hn
	}
	log.Printf("serving at: http://%s:%d/", hostname, c.Port)
	sm := http.NewServeMux()
	ms := vain.NewSimpleStore(c.DB)
	if err := ms.Load(); err != nil {
		log.Printf("unable to load db: %v; creating fresh database", err)
	}
	vain.NewServer(sm, ms, c.Host)
	addr := fmt.Sprintf(":%d", c.Port)
	if err := http.ListenAndServe(addr, sm); err != nil {
		log.Printf("problem with http server: %v", err)
		os.Exit(1)
	}
}
