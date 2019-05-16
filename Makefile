CWD=$(shell pwd)
GOPATH := $(CWD)

vendor-deps:
	go mod vendor

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

tools:
	go build -o bin/wof-dist-build cmd/wof-dist-build.go
	go build -o bin/wof-dist-fetch cmd/wof-dist-fetch.go
