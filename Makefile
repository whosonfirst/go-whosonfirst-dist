tools:
	@make cli

cli:
	go build -mod vendor -o bin/wof-dist-build cmd/wof-dist-build/main.go
	go build -mod vendor -o bin/wof-dist-fetch cmd/wof-dist-fetch/main.go
