.PHONY: build build-all test test-coverage install clean lint benchmark info help

# Binary name
BINARY_NAME=gimage
VERSION?=0.1.1
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build parameters
LDFLAGS=-ldflags "-X github.com/chadneal/gimage/internal/cli.version=$(VERSION)"

# Installation directory
INSTALL_DIR=/usr/local/bin

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary for current platform"
	@echo "  build-all      - Build binaries for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  install        - Install binary to $(INSTALL_DIR)"
	@echo "  clean          - Remove build artifacts"
	@echo "  info           - Display version and release notes"
	@echo "  lint           - Run linter"
	@echo "  benchmark      - Run benchmarks"

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gimage
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "All platform binaries built in $(BUILD_DIR)/"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## install: Install binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@if [ ! -d "$(INSTALL_DIR)" ]; then \
		echo "Creating $(INSTALL_DIR)..."; \
		sudo mkdir -p $(INSTALL_DIR); \
	fi
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Run 'gimage --help' to get started"

## info: Display version and release notes
info:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  $(BINARY_NAME) - AI Image Generation CLI"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo ""
	@if [ -f CHANGELOG.md ]; then \
		echo "Latest Release Notes:"; \
		echo ""; \
		sed -n '/^## \[$(VERSION)\]/,/^## \[/p' CHANGELOG.md | sed '$$d'; \
	else \
		echo "No release notes available (CHANGELOG.md not found)"; \
	fi
	@echo ""

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Module maintenance
.PHONY: mod-tidy mod-verify mod-download

mod-tidy:
	$(GOMOD) tidy

mod-verify:
	$(GOMOD) verify

mod-download:
	$(GOMOD) download
