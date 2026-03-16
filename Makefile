BINARY_NAME := cloudflared-project
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-s -w -X github.com/kev/cloudflared-cli/cmd.Version=$(VERSION) -X github.com/kev/cloudflared-cli/cmd.Commit=$(COMMIT) -X github.com/kev/cloudflared-cli/cmd.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint clean install cross install-script

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

cross:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

install-script:
	bash scripts/install.sh
