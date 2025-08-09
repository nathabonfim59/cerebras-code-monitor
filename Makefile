# Makefile for Cerebras Code Monitor

# Variables
BINARY_NAME=cerebras-code-monitor
MAIN_FILE=cmd/main.go
BUILD_DIR=build

# Default target
.PHONY: all
all: build

# Build the project
.PHONY: build
build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

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

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)

# Install the binary
.PHONY: install
install: build
	go install ./cmd/main.go

# Release using Goreleaser
.PHONY: release
release:
	goreleaser release --clean

# Dry run release using Goreleaser
.PHONY: release-dry
release-dry:
	goreleaser release --clean --skip-publish

# Snapshot release using Goreleaser
.PHONY: snapshot
snapshot:
	goreleaser release --clean --snapshot

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
	@echo "  release       - Create a new release using Goreleaser"
	@echo "  release-dry   - Run Goreleaser in dry-run mode"
	@echo "  snapshot      - Create a snapshot release"
	@echo "  help          - Show this help message"