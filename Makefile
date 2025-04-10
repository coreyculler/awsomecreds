# Project variables
BINARY_NAME=awsomecreds
PACKAGE=github.com/coreyculler/awsomecreds
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=build

.PHONY: all build clean test coverage deps tidy vet fmt lint help

# Default target
all: clean test build

# Build the application
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the application
install:
	@echo "Installing ${BINARY_NAME}..."
	$(GOBUILD) $(LDFLAGS) -o ${GOPATH}/bin/$(BINARY_NAME)
	@echo "Installation complete: ${GOPATH}/bin/$(BINARY_NAME)"

# Clean build files
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./... -tags=skipintegration
	@echo "Tests complete"

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@echo "Note: Integration tests require the following environment variables to be set:"
	@echo "  - TEST_ROLE_ARN: The ARN of the role to assume (required)"
	@echo "  - TEST_SOURCE_PROFILE: The AWS profile to use (optional)"
	@echo "  - TEST_MFA_TOKEN: MFA token if required by the role (optional)"
	@echo "  - TEST_REGION: AWS region (optional)"
	@echo ""
	RUN_INTEGRATION_TESTS=1 $(GOTEST) -v ./...
	@echo "Integration tests complete"

# Generate test coverage
coverage:
	@echo "Generating test coverage..."
	$(GOTEST) -coverprofile=coverage.out ./... -tags=skipintegration
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...
	@echo "Dependencies downloaded"

# Tidy go.mod
tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy
	@echo "Modules tidied"

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Formatting complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...
	@echo "Lint complete"

# Cross-compile for different platforms
cross-build:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Cross-compilation complete"

# Help command
help:
	@echo "Available commands:"
	@echo "  make all             - Clean, test, and build the application"
	@echo "  make build           - Build the application"
	@echo "  make install         - Install the application to GOPATH/bin"
	@echo "  make clean           - Clean build files"
	@echo "  make test            - Run tests (excluding integration tests)"
	@echo "  make integration-test - Run all tests including integration tests"
	@echo "  make coverage        - Generate test coverage report"
	@echo "  make deps            - Download dependencies"
	@echo "  make tidy            - Tidy go.mod file"
	@echo "  make vet             - Run go vet"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter"
	@echo "  make cross-build     - Cross-compile for multiple platforms" 