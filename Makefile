.PHONY: build test clean

VERSION := 0.1.0
BINARY := denote-tasks

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) cmd/denote-tasks/main.go

test:
	go test ./...

clean:
	rm -f $(BINARY)

install: build
	cp $(BINARY) ~/bin/