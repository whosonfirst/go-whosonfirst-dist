dist-build:
	OS=darwin make dist-os
	OS=windows make dist-os
	OS=linux make dist-os

dist-os:
	mkdir -p dist/$(OS)
	GOOS=$(OS) GOARCH=386 go build -o dist/$(OS)/wof-build-metafiles cmd/wof-build-metafiles.go
	chmod +x dist/$(OS)/wof-build-metafiles
	cd dist/$(OS) && shasum -a 256 wof-build-metafiles > wof-build-metafiles.sha256

fmt:
	go fmt *.go
	go fmt build/*.go
	# go fmt cmd/*/*.go
	go fmt meta/*.go
	go fmt options/*.go
	go fmt stats/*.go

tools:
	go build -o bin/wof-build-metafiles cmd/wof-build-metafiles/main.go
	go build -o bin/wof-update-metafile cmd/wof-update-metafile/main.go
	go build -o bin/wof-meta-prepare cmd/wof-meta-prepare/main.go
	go build -o bin/wof-meta-stats cmd/wof-meta-stats/main.go

###

CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-meta
	cp -r build src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r meta src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r options src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r stats src/github.com/whosonfirst/go-whosonfirst-meta/
	cp *.go src/github.com/whosonfirst/go-whosonfirst-meta
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi 


bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-build-metafiles cmd/wof-build-metafiles/main.go
	@GOPATH=$(GOPATH) go build -o bin/wof-update-metafile cmd/wof-update-metafile/main.go
	@GOPATH=$(GOPATH) go build -o bin/wof-meta-prepare cmd/wof-meta-prepare/main.go
	@GOPATH=$(GOPATH) go build -o bin/wof-meta-stats cmd/wof-meta-stats/main.go

