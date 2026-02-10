# Port Chaser Makefile
# Development and deployment automation

.PHONY: build test clean install release release-prepare release-upload

# Variables
BINARY_NAME=port-chaser
CMD_DIR=./cmd/port-chaser
BUILD_DIR=./bin
VERSION?=0.1.0
GO_VERSION=1.21

# Build flags
LDFLAGS=-s -w -X main.version=$(VERSION)
BUILD_FLAGS=-ldflags "$(LDFLAGS)"

# Default target
all: build

## build: Build the binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "==> Built $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "==> Running tests..."
	go test -v -race -cover ./...

## test-short: Run short tests only
test-short:
	@echo "==> Running short tests..."
	go test -v -short ./...

## clean: Remove build artifacts
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "==> Clean complete"

## install: Install locally
install: build
	@echo "==> Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "==> Installed successfully"

## lint: Run code linters
lint:
	@echo "==> Running linters..."
	go vet ./...
	golangci-lint run 2>/dev/null || echo "golangci-lint not installed, skipping..."

## fmt: Format code
fmt:
	@echo "==> Formatting code..."
	go fmt ./...
	goimports -w . 2>/dev/null || echo "goimports not installed, skipping..."

## deps: Download dependencies
deps:
	@echo "==> Downloading dependencies..."
	go mod download
	go mod tidy

## run: Build and run the binary
run: build
	@echo "==> Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

## release: Build release binaries for GitHub release (multi-platform)
release: clean
	@echo "==> Building release binaries for v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)

	@echo "==> Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)

	@echo "==> Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

	@echo "==> Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

	@echo "==> Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

	@echo "==> Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

	@echo "==> Creating SHA256 checksums..."
	@cd $(BUILD_DIR) && shasum -a 256 $(BINARY_NAME)* > SHA256SUMS.txt

	@echo "==> Release binaries built successfully!"
	@echo "==> Files in $(BUILD_DIR):"
	@ls -lh $(BUILD_DIR)

## release-prepare: Prepare Homebrew Formula SHA256 update
release-prepare:
	@echo "==> Prepare for Homebrew formula update..."
	@echo "==> After uploading release, update SHA256 in Formula/port-chaser.rb"
	@echo "==> Run 'shasum -a 256 <downloaded-tarball>' to get SHA256"

## help: Display available commands
help:
	@echo "Port Chaser - Available Commands:"
	@echo ""
	@echo "  make build         - Build the binary"
	@echo "  make test          - Run all tests"
	@echo "  make test-short    - Run short tests only"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make install       - Install to /usr/local/bin"
	@echo "  make lint          - Run linters"
	@echo "  make fmt           - Format code"
	@echo "  make deps          - Download dependencies"
	@echo "  make run           - Build and run the binary"
	@echo "  make release       - Build release binaries (multi-platform)"
	@echo "  make release-prepare - Prepare Homebrew formula update"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make                  # Build the binary"
	@echo "  make test             # Run tests"
	@echo "  make VERSION=0.2.0 release  # Build v0.2.0 release"
