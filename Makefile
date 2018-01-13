CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src/github.com/whosonfirst/go-whosonfirst-dist; then rm -rf src/github.com/whosonfirst/go-whosonfirst-dist; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-dist
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

deps:
	@GOPATH=$(GOPATH) go get -u "gopkg.in/src-d/go-git.v4/..."
	# @GOPATH=$(GOPATH) go get -u "github.com/jteeuwen/go-bindata/"
	# @GOPATH=$(GOPATH) go get -u "github.com/dustin/go-humanize"
	# rm -rf src/github.com/jteeuwen/go-bindata/testdata

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-dist-build cmd/wof-dist-build.go
