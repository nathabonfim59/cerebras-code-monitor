# Makefile for Cerebras Code Monitor

# Variables
BINARY_NAME=cerebras-code-monitor
MAIN_FILE=cmd/main.go
BUILD_DIR=build
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.PHONY: all
all: build

# Generate SQL code
.PHONY: sqlc
sqlc:
	sqlc generate

# Build the project
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

# Build the project for production
.PHONY: build-prod
build-prod:
	go build -tags prod $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_FILE)

# Install dependencies
.PHONY: deps
deps:
	go mod tidy

# Run tests
.PHONY: test
test:
	go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -cover ./...

# Lint the project
.PHONY: lint
lint:
	golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	go fmt ./...

# Format uncommitted Go files
.PHONY: fmt-uncommitted
fmt-uncommitted:
	gofmt -w $$(git diff --name-only --diff-filter=ACMR | grep '\.go$$' | xargs)

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Build for multiple platforms with CGO support
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)

# Install the binary
.PHONY: install
install: build
	go install ./cmd/main.go

# Create GitHub release (manual build process)
.PHONY: release
release:
	@echo "Releases are now created via GitHub Actions when you push a tag."
	@echo "To create a release:"
	@echo "  1. git tag v1.0.0"
	@echo "  2. git push origin v1.0.0"
	@echo "The GitHub Action will automatically build and release for all platforms."

# Build local snapshot for testing
.PHONY: snapshot
snapshot: build-all
	@echo "Local snapshot built in $(BUILD_DIR)/"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Build the monitor (default)"
	@echo "  build         - Build the monitor"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Lint the project"
	@echo "  fmt           - Format the code"
	@echo "  fmt-uncommitted - Format uncommitted Go files"
	@echo "  clean         - Clean build artifacts"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  install       - Install the binary"
	@echo "  release       - Show instructions for creating a GitHub release"
	@echo "  snapshot      - Build local snapshot for testing"
	@echo "  help          - Show this help message"