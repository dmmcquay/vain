/*
Package vain implements a vanity service for use by the the go tool

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
		"vcs": "mercurial",
		"path": "mcquay.me/vain",
		"repo": "https://s.mcquay.me/sm/vain"
	}

"path" and "repo" are required; leaving off the "vcs" member defaults to "git".
*/
package vain
