# Contributing

We love pull requests from everyone. By participating in this project, you
agree to abide by the golang [code of conduct].

[code of conduct]: https://golang.org/conduct

I'm trying an experiment in running as much of my own infrastructure as
I possibly can, which includes hosting my git repos using [gogs] at
https://s.mcquay.me/. As such contributing will be slightly more complex than
a simple pull request on github. Sorry for the inconvenience.

[gogs]: https://gogs.io

The basic steps begin with a clone this repo:

	$ git clone https://s.mcquay.me/sm/vain.git $GOPATH/src/mcquay.me/vain

then set up to push up to your hosting service of choice, e.g.:

    $ cd $GOPATH/src/mcquay.me/vain
    $ git remote add mine git@github.com/you/vain.git

add a feature and some tests then run the tests, check your formatting:

	$ go test mcquay.me/vain
	$ go vet mcquay.me/vain
	$ golint mcquay.me/vain

If things look good and tests pass commit and push to your remote:

	$ git add (files you changed)
	$ git commit -m "Job's done"
	$ git push mine feature


Push to your fork and email a pull request to stephen at mcquay dot me with
a link to your branch, of the form:

    https://github.com/you/vain/tree/branch-name

At this point you're waiting on us. We will comment on the pull request request
within three business days (and, typically, one business day). We may suggest
some changes or improvements or alternatives.

Some things that will increase the chance that your pull request is accepted:

* Write tests.
* follow good Go style, including [effective go], running [go vet] and [golint].
* Write a [good commit message][commit].

[effective go]: https://golang.org/doc/effective_go.html
[go vet]: https://golang.org/cmd/vet/
[golint]: https://github.com/golang/lint
[commit]: http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html

contribution guidelines borrowed from [factory girl rails].

[factory girl rails]: https://github.com/thoughtbot/factory_girl_rails
