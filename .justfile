# blackboard — 開発ワークフロー
# Usage: just <recipe> [args...]
# List all: just --list

# Build the bb CLI
build:
    go build -o dist/bb main.go

# Run all tests
test:
    go test ./...

# Run go vet (lint)
lint:
    go vet ./...

# Format Go source code
fmt:
    go fmt ./...

# Verify Go source is formatted
fmt-check:
    #!/usr/bin/env sh
    set -eu
    files="$(gofmt -l .)"
    if [ -n "$files" ]; then
        printf 'Go files are not formatted:\n%s\n' "$files"
        exit 1
    fi

install:
  just build
  cp dist/bb ~/.go/bin/bb

# Run all verification checks
check: fmt-check lint test
