.PHONY: build build-all test test-coverage install clean lint benchmark info help build-lambda package-lambda deploy-lambda clean-lambda lambda-logs release update-changelog sync-version version

# Binary name
BINARY_NAME=gimage

# Version: 1.1.[build_number] where build_number is the git commit count
BUILD_NUMBER=$(shell git rev-list --count HEAD 2>/dev/null || echo "0")
VERSION?=1.1.$(BUILD_NUMBER)

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
LDFLAGS=-ldflags "-X github.com/apresai/gimage/internal/cli.version=$(VERSION)"

# Installation directory
INSTALL_DIR=/usr/local/bin

# Default target
all: build

## help: Display this help message
help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Gimage - AI Image Generation CLI & MCP"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔨 Build Commands:"
	@echo "  build            - Build binary for current platform"
	@echo "  build-all        - Build binaries for all platforms"
	@echo "  install          - Install binary to $(INSTALL_DIR)"
	@echo "  clean            - Remove build artifacts"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  test             - Run ALL tests (unit + E2E) with coverage"
	@echo "  test-unit        - Run unit tests only (fast)"
	@echo "  test-e2e         - Run all E2E tests (CLI + Generate)"
	@echo "  test-cli-e2e     - Run CLI E2E tests only (FREE)"
	@echo "  test-generate-e2e- Run Generate Image E2E tests (costs \$$)"
	@echo "  test-coverage    - Generate coverage report from coverage.out"
	@echo ""
	@echo "☁️  Lambda Commands:"
	@echo "  build-lambda     - Build Lambda function for AWS ARM64"
	@echo "  package-lambda   - Package Lambda function for deployment"
	@echo "  deploy-lambda    - Deploy Lambda function using CDK"
	@echo "  clean-lambda     - Remove Lambda build artifacts"
	@echo "  lambda-logs      - Tail Lambda function logs"
	@echo ""
	@echo "📦 Release Commands:"
	@echo "  version          - Display current version"
	@echo "  sync-version     - Sync version to package.json"
	@echo "  update-changelog - Update CHANGELOG.md"
	@echo "  release          - Create and publish new release"
	@echo "  info             - Display version and release notes"
	@echo ""
	@echo "🔍 Quality Commands:"
	@echo "  lint             - Run linter"
	@echo "  benchmark        - Run benchmarks"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "💡 Quick Start: make test"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

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

## test: Run all tests (unit + E2E) and generate coverage report
test: build
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "                        🧪 GIMAGE COMPLETE TEST SUITE"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Running unit tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./... > /tmp/gimage-unit-tests.log 2>&1 || true
	@echo "Running CLI E2E tests (free)..."
	@$(GOTEST) -v -tags=e2e ./test/integration/cli_e2e_test.go > /tmp/gimage-cli-e2e-tests.log 2>&1 || true
	@echo "Running Generate Image E2E tests (costs money)..."
	@$(GOTEST) -v -tags=e2e ./test/integration/generate_e2e_test.go > /tmp/gimage-generate-e2e-tests.log 2>&1 || true
	@echo ""
	@go run cmd/test-summary/main.go /tmp/gimage-unit-tests.log /tmp/gimage-cli-e2e-tests.log /tmp/gimage-generate-e2e-tests.log || TEST_FAILED=1
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "                              📈 COVERAGE REPORT"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Generating coverage reports..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html 2>/dev/null
	@go run cmd/coverage-report/main.go coverage.out 2>/dev/null
	@echo ""
	@echo "Core Packages (Unit Test Coverage):"
	@go tool cover -func=coverage.out | grep "internal/imaging\|internal/generate\|internal/mcp" | grep -v "total:" | awk 'BEGIN{sum=0;count=0}{gsub("%","",$$NF);sum+=$$NF;count++}END{if(count>0)printf "  %.1f%% average coverage (%d files)\n", sum/count, count}'
	@echo ""
	@go run cmd/test-summary/main.go --cli-coverage /tmp/gimage-cli-e2e-tests.log
	@echo ""
	@echo "HTML Reports Generated:"
	@echo "  • coverage-report.html  (readable summary - OPEN THIS FIRST)"
	@echo "  • coverage.html         (detailed line-by-line)"
	@echo ""
	@if [ "$$TEST_FAILED" = "1" ]; then \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "                              ❌ SOME TESTS FAILED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "📋 Check test logs:"; \
		echo "   cat /tmp/gimage-unit-tests.log"; \
		echo "   cat /tmp/gimage-cli-e2e-tests.log"; \
		echo "   cat /tmp/gimage-generate-e2e-tests.log"; \
		echo ""; \
		exit 1; \
	else \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "                              ✅ ALL TESTS PASSED"; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo ""; \
		echo "🔗 View detailed coverage report:"; \
		echo "   open coverage-report.html"; \
		echo ""; \
	fi

## test-unit: Run unit tests only (no E2E)
test-unit:
	@echo "Running unit tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "✓ Unit tests complete"
	@go tool cover -func=coverage.out | grep "total:"

## test-e2e: Run all E2E tests (CLI + Generate Image)
test-e2e: test-cli-e2e test-generate-e2e

## test-cli-e2e: Run CLI E2E tests (free, no API calls)
test-cli-e2e: build
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running CLI E2E tests (resize, scale, crop)..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@$(GOTEST) -v -tags=e2e ./test/integration/cli_e2e_test.go
	@echo ""
	@echo "✓ CLI E2E tests complete (FREE - no API costs)"

## test-generate-e2e: Run Generate Image E2E tests (costs money!)
test-generate-e2e:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "⚠️  WARNING: E2E tests will make real API calls!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "   - Gemini: Uses free tier quota (FREE)"
	@echo "   - Vertex: ~\$$0.02 per test"
	@echo "   - Bedrock: ~\$$0.04 per test"
	@echo ""
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "Running Generate Image E2E tests..."; \
		$(GOTEST) -v -tags=e2e ./test/integration/generate_e2e_test.go; \
		echo ""; \
		echo "✓ Generate Image E2E tests complete"; \
	else \
		echo "E2E tests cancelled"; \
	fi

## test-coverage: Generate coverage report from existing coverage.out
test-coverage:
	@if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found. Run 'make test' or 'make test-unit' first."; \
		exit 1; \
	fi
	@echo "Generating coverage reports..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@go run cmd/coverage-report/main.go coverage.out
	@echo ""
	@echo "✓ Coverage reports generated:"
	@echo "  • coverage-report.html (readable summary)"
	@echo "  • coverage.html (detailed line-by-line)"

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
	rm -f coverage.out coverage.html coverage-report.html
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

## version: Display current version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Number: $(BUILD_NUMBER)"

## sync-version: Sync version to package.json files
sync-version:
	@echo "Syncing version $(VERSION) to package.json files..."
	@if [ -f package.json ]; then \
		sed -i.bak 's/"version": "[^"]*"/"version": "$(VERSION)"/' package.json && \
		rm package.json.bak && \
		echo "✓ package.json updated to $(VERSION)"; \
	else \
		echo "✗ package.json not found"; \
		exit 1; \
	fi
	@if [ -f npm/package.json ]; then \
		sed -i.bak 's/"version": "[^"]*"/"version": "$(VERSION)"/' npm/package.json && \
		rm npm/package.json.bak && \
		echo "✓ npm/package.json updated to $(VERSION)"; \
	else \
		echo "✗ npm/package.json not found"; \
	fi
	@echo "CLI and MCP versions are now in sync: $(VERSION)"

## update-changelog: Update CHANGELOG.md with new version
update-changelog:
	@echo "Updating CHANGELOG.md with version $(VERSION)..."
	@if [ ! -f scripts/update-changelog.sh ]; then \
		echo "✗ scripts/update-changelog.sh not found"; \
		exit 1; \
	fi
	@bash scripts/update-changelog.sh $(VERSION)

## release: Create and publish a new release
release:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  Creating Release v$(VERSION)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Step 1: Updating CHANGELOG.md..."
	@$(MAKE) update-changelog
	@echo ""
	@echo "Step 2: Syncing version to package.json files..."
	@$(MAKE) sync-version
	@echo ""
	@echo "Step 3: Checking for uncommitted changes..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "✓ Found changes, committing..."; \
		git add CHANGELOG.md package.json npm/package.json; \
		git commit -m "Release v$(VERSION)"; \
		git push origin main; \
		echo "✓ Changes committed and pushed"; \
	else \
		echo "✓ No changes to commit"; \
	fi
	@echo ""
	@echo "Step 4: Creating git tag v$(VERSION)..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)" || (echo "✗ Tag already exists" && exit 1)
	@git push origin v$(VERSION)
	@echo "✓ Tag v$(VERSION) created and pushed"
	@echo ""
	@echo "Step 5: Building and publishing with GoReleaser..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "Setting GITHUB_TOKEN from gh auth..."; \
		export GITHUB_TOKEN=$$(gh auth token); \
	fi; \
	if [ -z "$$HOMEBREW_TAP_TOKEN" ]; then \
		echo "⚠️  Warning: HOMEBREW_TAP_TOKEN not set"; \
		echo "   Homebrew formula will not be updated"; \
		echo "   Set it with: export HOMEBREW_TAP_TOKEN=<token>"; \
	fi; \
	GITHUB_TOKEN=$${GITHUB_TOKEN} HOMEBREW_TAP_TOKEN=$${HOMEBREW_TAP_TOKEN} goreleaser release --clean
	@echo ""
	@echo "Step 6: Publishing npm package..."
	@cd npm && npm publish --access public
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  ✓ Release v$(VERSION) Complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "GitHub Release: https://github.com/apresai/gimage/releases/tag/v$(VERSION)"
	@echo "npm Package: https://www.npmjs.com/package/@apresai/gimage-mcp/v/$(VERSION)"
	@echo ""
	@echo "Installation:"
	@echo "  Homebrew: brew install apresai/tap/gimage"
	@echo "  npm:      npm install -g @apresai/gimage-mcp"
	@echo ""
