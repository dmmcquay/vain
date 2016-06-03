/*
vaind, a webserver for hosting go get vanity urls.

The go get command searches for the following header when searching for
packages:

	<meta name="go-import" content="import-prefix vcs repo-root">

this is simply a service for aggregating a collection of prefix, vcs, and
repo-root tuples, and serving the appropriate header over http. For more
information please refer to the documentation for the go tool found at
https://golang.org/cmd/go/#hdr-Remote_import_paths

API

Assume an instance of vaind at example.org. In order to add a package
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

In order to delete a package:

	DELETE /<package name>
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"mcquay.me/vain"

	"github.com/kelseyhightower/envconfig"
)

const usage = "vaind [init] <dbname>"

type config struct {
	Port     int
	Insecure bool

	Cert string
	Key  string

	Static string

	EmailTimeout time.Duration `envconfig:"email_timeout"`

	SMTPHost string `envconfig:"smtp_host"`
	SMTPPort int    `envconfig:"smtp_port"`

	From string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}

	if os.Args[1] == "init" {
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "missing db name: %s\n", usage)
			os.Exit(1)
		}

		db, err := vain.NewDB(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't open db: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		if err := db.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "problem initializing the db: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	db, err := vain.NewDB(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't open db: %v\n", err)
		os.Exit(1)
	}

	c := &config{
		Port:         4040,
		EmailTimeout: 5 * time.Minute,
		SMTPPort:     25,
	}
	if err := envconfig.Process("vain", c); err != nil {
		fmt.Fprintf(os.Stderr, "problem processing environment: %v", err)
		os.Exit(1)
	}
	log.Printf("%+v", c)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "env", "e", "help", "h":
			fmt.Printf("%s\n", usage)
			os.Exit(0)
		}
	}

	m, err := vain.NewEmail(c.From, c.SMTPHost, c.SMTPPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem initializing mailer: %v", err)
		os.Exit(1)
	}

	hostname := "localhost"
	if hn, err := os.Hostname(); err != nil {
		log.Printf("problem getting hostname: %v", err)
	} else {
		hostname = hn
	}
	log.Printf("serving at: http://%s:%d/", hostname, c.Port)
	sm := http.NewServeMux()
	vain.NewServer(sm, db, m, c.Static, c.EmailTimeout, c.Insecure)
	addr := fmt.Sprintf(":%d", c.Port)

	if c.Cert == "" || c.Key == "" {
		log.Printf("INSECURE MODE")
		if err := http.ListenAndServe(addr, sm); err != nil {
			log.Printf("problem with http server: %v", err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServeTLS(addr, c.Cert, c.Key, sm); err != nil {
			log.Printf("problem with http server: %v", err)
			os.Exit(1)
		}
	}
}
