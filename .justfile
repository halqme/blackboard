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

install:
  just build
  cp dist/bb ~/.go/bin/bb


# Run all verification checks (test + lint)
check: lint test
