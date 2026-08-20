BINARY := prwlrctl

.PHONY: build test lint fmt install

build:
	go build -trimpath -ldflags="-s -w" -o bin/$(BINARY) .

install:
	go install .

test:
	go test ./...

fmt:
	gofmt -l -w .

lint:
	go vet ./...