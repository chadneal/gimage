.PHONY: build build-all test test-coverage install clean lint benchmark info help build-lambda package-lambda deploy-lambda clean-lambda lambda-logs

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
	@echo "  build           - Build the binary for current platform"
	@echo "  build-all       - Build binaries for all platforms"
	@echo "  build-lambda    - Build Lambda function for AWS ARM64"
	@echo "  package-lambda  - Package Lambda function for deployment"
	@echo "  deploy-lambda   - Deploy Lambda function using CDK"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  install         - Install binary to $(INSTALL_DIR)"
	@echo "  clean           - Remove build artifacts"
	@echo "  clean-lambda    - Remove Lambda build artifacts"
	@echo "  info            - Display version and release notes"
	@echo "  lint            - Run linter"
	@echo "  benchmark       - Run benchmarks"
	@echo "  lambda-logs     - Tail Lambda function logs"

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

# Lambda-specific targets

## build-lambda: Build Lambda function binary for AWS ARM64
build-lambda:
	@echo "Building Lambda function for AWS ARM64..."
	@mkdir -p $(BUILD_DIR)/lambda
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) \
		-tags lambda.norpc \
		-o $(BUILD_DIR)/lambda/bootstrap \
		./cmd/lambda
	@chmod +x $(BUILD_DIR)/lambda/bootstrap
	@echo "Lambda binary built: $(BUILD_DIR)/lambda/bootstrap ($(shell du -h $(BUILD_DIR)/lambda/bootstrap 2>/dev/null | cut -f1))"

## package-lambda: Package Lambda function for deployment
package-lambda: build-lambda
	@echo "Packaging Lambda function..."
	@cd $(BUILD_DIR)/lambda && zip -q -r ../lambda.zip bootstrap
	@echo "Lambda package created: $(BUILD_DIR)/lambda.zip ($(shell du -h $(BUILD_DIR)/lambda.zip 2>/dev/null | cut -f1))"

## deploy-lambda: Deploy Lambda function using CDK
deploy-lambda: package-lambda
	@echo "Deploying Lambda function with CDK..."
	@if [ -d "infrastructure/cdk" ]; then \
		cd infrastructure/cdk && npm install && npm run build && npm run deploy; \
	else \
		echo "Error: infrastructure/cdk directory not found"; \
		echo "Please create CDK infrastructure first (see lambda.md)"; \
		exit 1; \
	fi

## clean-lambda: Clean Lambda build artifacts
clean-lambda:
	@echo "Cleaning Lambda artifacts..."
	@rm -rf $(BUILD_DIR)/lambda $(BUILD_DIR)/lambda.zip
	@echo "Lambda artifacts cleaned"

## lambda-logs: Tail Lambda function logs
lambda-logs:
	@echo "Tailing Lambda logs..."
	@if command -v aws > /dev/null; then \
		aws logs tail /aws/lambda/gimage-processor --follow; \
	else \
		echo "Error: AWS CLI not installed"; \
		echo "Install with: brew install awscli"; \
		exit 1; \
	fi

## lambda-invoke-local: Test Lambda function locally
lambda-invoke-local: build-lambda
	@echo "Testing Lambda locally..."
	@echo "Set environment variables and run:"
	@echo "  export S3_BUCKET=test-bucket"
	@echo "  export GEMINI_API_KEY=your_key"
	@echo "  cd $(BUILD_DIR)/lambda && ./bootstrap"
