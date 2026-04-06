# Harness v4 Go build
#
# Usage:
#   make build          — Build for current platform
#   make install        — Build and copy to ../bin/harness
#   make test           — Run all tests
#   make clean          — Remove build artifacts

VERSION := $(shell cat ../VERSION 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
BINARY  := harness
OUTDIR  := ../bin

.PHONY: build install test clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/harness/

install: build
	mkdir -p $(OUTDIR)
	cp $(BINARY) $(OUTDIR)/$(BINARY)
	@echo "Installed $(OUTDIR)/$(BINARY) ($(VERSION))"

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -f $(OUTDIR)/$(BINARY)
