# Build variables
BINARY_NAME=gitlab-ci-lint
BUILD_DIR=build
CMD_DIR=cmd/gitlab-ci-lint
MAIN_FILE=$(CMD_DIR)/main.go

# Version variables (injected at build time)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS=-ldflags "-X github.com/InkyQuill/gitlab-ci-lint/pkg/version.Version=$(VERSION) \
                  -X github.com/InkyQuill/gitlab-ci-lint/pkg/version.Commit=$(COMMIT) \
                  -X github.com/InkyQuill/gitlab-ci-lint/pkg/version.BuildDate=$(BUILD_DATE)"

# Build platforms
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/386

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build: clean
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach PLATFORM,$(PLATFORMS),\
		echo "Building for $(PLATFORM)...";\
		GOOS=$(word 1,$(subst /, ,$(PLATFORM))) \
		GOARCH=$(word 2,$(subst /, ,$(PLATFORM))) \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(word 1,$(subst /, ,$(PLATFORM)))-$(word 2,$(subst /, ,$(PLATFORM)))$(if $(filter windows%,$(PLATFORM)),.exe,) $(MAIN_FILE);\
	)
	@echo "Build complete for all platforms"

# Run all tests
.PHONY: test
test:
	@echo "Running all tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

# Run unit tests only
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test -v -race -coverprofile=coverage_unit.out ./internal/... ./pkg/...
	go tool cover -func=coverage_unit.out | grep total

# Run integration tests only
.PHONY: test-integration
test-integration: build
	@echo "Running integration tests..."
	go test -v -race -coverprofile=coverage_integration.out ./tests/...
	go tool cover -func=coverage_integration.out | grep total

# Run all tests with combined coverage
.PHONY: test-all
test-all: test-unit test-integration
	@echo "Combining coverage reports..."
	@echo "mode: set" > coverage_combined.out
	@for file in coverage_unit.out coverage_integration.out; do \
		if [ -f $$file ]; then \
			grep -v "^mode:" $$file >> coverage_combined.out || true; \
		fi; \
	done
	@echo "Total coverage:"
	@go tool cover -func=coverage_combined.out | grep total
	@go tool cover -html=coverage_combined.out -o coverage_combined.html
	@echo "Combined coverage report: coverage_combined.html"

# Run tests with coverage (legacy target, redirects to test-all)
.PHONY: test-coverage
test-coverage: test-all

# Run linter
.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run gofmt
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Install to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_FILE)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Run the application
.PHONY: run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Build for current platform (default)"
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for all platforms"
	@echo "  test          - Run all tests"
	@echo "  test-unit     - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-all      - Run all tests with combined coverage report"
	@echo "  test-coverage - Run tests with coverage report (alias for test-all)"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run (use ARGS='--help' for options)"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION      - Version string (default: git describe or 'dev')"
	@echo "  COMMIT       - Commit hash (default: git rev-parse --short HEAD)"
	@echo "  BUILD_DATE   - Build timestamp (default: current UTC time)"
	@echo "  ARGS         - Arguments for 'run' target"
