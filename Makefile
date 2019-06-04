fmt:
	go fmt *.go
	go fmt build/*.go
	go fmt cmd/wof-dist-build/*.go
	go fmt cmd/wof-dist-fetch/*.go
	go fmt compress/*.go
	go fmt csv/*.go
	go fmt database/*.go
	go fmt fs/*.go
	go fmt git/*.go
	go fmt hash/*.go
	go fmt options/*.go
	go fmt utils/*.go

tools:
	go build -mod vendor -o bin/wof-dist-build cmd/wof-dist-build/main.go
	go build -mod vendor -o bin/wof-dist-fetch cmd/wof-dist-fetch/main.go
