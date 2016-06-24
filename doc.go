/*
Package vain implements a vanity service for use by the the go tool.


The executable, cmd/vaind, is located in the respective subdirectory. vaind,
a webserver for hosting go get vanity urls.

The go get command searches for the following header when searching for
packages:

	<meta name="go-import" content="import-prefix vcs repo-root">

this is simply a service for aggregating a collection of prefix, vcs, and
repo-root tuples, and serving the appropriate header over http. For more
information please refer to the documentation for the go tool found at
https://golang.org/cmd/go/#hdr-Remote_import_paths

For instructions on how to use this service, build the daemon, run it, and
visit the root url.
*/
package vain
