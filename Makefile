CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src/github.com/whosonfirst/go-whosonfirst-dist; then rm -rf src/github.com/whosonfirst/go-whosonfirst-dist; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-dist
	cp -r build src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r compress src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r csv src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r database src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r fs src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r git src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r hash src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r options src/github.com/whosonfirst/go-whosonfirst-dist/
	cp -r utils src/github.com/whosonfirst/go-whosonfirst-dist/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-dist
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	fmt bin

# see notes in git/git.go - this adds a non-trivial set of dependencies
# so we're excluding it for now (20180816/thisisaaronland)
# @GOPATH=$(GOPATH) go get -u "gopkg.in/src-d/go-git.v4/..."

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/jtacoma/uritemplates"
	@GOPATH=$(GOPATH) go get -u "github.com/mholt/archiver"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/pretty"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-meta"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-sqlite-features"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-repo"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-bundles"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/atomicfile"
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-sqlite src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-index src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-log src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-index/vendor/github.com/whosonfirst/go-whosonfirst-crawl src/github.com/whosonfirst/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/tidwall/gjson src/github.com/tidwall/
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/tidwall/match src/github.com/tidwall/
	mv src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/facebookgo src/github.com/
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-index
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-log
	rm -rf src/github.com/whosonfirst/go-whosonfirst-bundles/vendor/github.com/whosonfirst/go-whosonfirst-log
	rm -rf src/github.com/whosonfirst/go-whosonfirst-bundles/vendor/github.com/whosonfirst/go-whosonfirst-meta
	rm -rf src/github.com/whosonfirst/go-whosonfirst-bundles/vendor/github.com/whosonfirst/go-whosonfirst-index
	rm -rf src/github.com/whosonfirst/go-whosonfirst-bundles/vendor/github.com/whosonfirst/go-whosonfirst-sqlite
	rm -rf src/github.com/mholt/archiver/testdata

vendor-deps: rmdeps deps
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt build/*.go
	go fmt cmd/*.go
	go fmt compress/*.go
	go fmt csv/*.go
	go fmt database/*.go
	go fmt fs/*.go
	go fmt git/*.go
	go fmt hash/*.go
	go fmt options/*.go
	go fmt utils/*.go

bin: 	self
	rm -rf bin/*
	@GOPATH=$(GOPATH) go build -o bin/wof-dist-build cmd/wof-dist-build.go
	@GOPATH=$(GOPATH) go build -o bin/wof-dist-fetch cmd/wof-dist-fetch.go
