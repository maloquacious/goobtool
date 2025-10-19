# Goobergine Application Generator — Makefile (v0.1-alpha)
# Simple, portable targets with sensible defaults.
# Usage examples:
#   make build
#   make run PORT=8080 ADMIN_PORT=8383
#   make test
#   make clean

BIN ?= app
PORT ?= 8080
ADMIN_PORT ?= 8383
DIST ?= dist

.PHONY: all build run test tidy clean help

all: build

build:
	mkdir -p $(DIST)
	go build -o $(DIST)/$(BIN) ./cmd/$(BIN)

run:
	go run ./cmd/$(BIN) --port $(PORT) --admin-port $(ADMIN_PORT)

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(DIST)

help:
	@echo "Targets:"
	@echo "  build             Build binary to $(DIST)/$(BIN)"
	@echo "  run               Run with --port $(PORT) and --admin-port $(ADMIN_PORT)"
	@echo "  test              Run all tests"
	@echo "  tidy              Run 'go mod tidy'"
	@echo "  clean             Remove $(DIST) directory"
