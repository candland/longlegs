.PHONY: build clean deploy gomodgen

build: gomodgen
	# env GOOS=linux go build -ldflags="-s -w" -o bin/pr cmd/cli/main.go

local: gomodgen
	env go build -ldflags="-s -w" -o bin/longlegs cmd/longlegs/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

gomodgen:
	export GO111MODULE=on
