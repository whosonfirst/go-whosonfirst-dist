CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src/github.com/whosonfirst/go-whosonfirst-dist; then rm -rf src/github.com/whosonfirst/go-whosonfirst-dist; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-dist
	cp -r git src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r sqlite src/github.com/whosonfirst/go-whosonfirst-dist/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-dist
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "gopkg.in/src-d/go-git.v4/..."
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-sqlite-features"
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-sqlite src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-index src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-log src/github.com/whosonfirst/

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt git/*.go
	go fmt sqlite/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-dist-build cmd/wof-dist-build.go
	@GOPATH=$(GOPATH) go build -o bin/wof-dist-fetch cmd/wof-dist-fetch.go
