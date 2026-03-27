VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X github.com/aohoyd/aku/pkg/build.Version=$(VERSION) \
	-X github.com/aohoyd/aku/pkg/build.Commit=$(COMMIT) \
	-X github.com/aohoyd/aku/pkg/build.Date=$(DATE)

.PHONY: build install test release

build:
	go build -ldflags "$(LDFLAGS)" -o aku .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

release:
	goreleaser release --snapshot --clean
