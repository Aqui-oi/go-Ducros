# Makefile for Ducros Stratum Proxy

.PHONY: all build clean install test run help

# Build configuration
BINARY_NAME=stratum-proxy
BUILD_DIR=build
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

all: build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "✅ Multi-platform build complete"

# Install to system
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)
	$(GO) clean
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GO) test -v ./...

# Run the proxy
run: build
	@echo "🚀 Starting proxy..."
	./$(BUILD_DIR)/$(BINARY_NAME) \
		--stratum "0.0.0.0:3333" \
		--geth "http://localhost:8545" \
		--diff 10000 \
		-v

# Format code
fmt:
	@echo "💅 Formatting code..."
	$(GO) fmt ./...

# Lint code
lint:
	@echo "🔍 Linting..."
	golangci-lint run ./...

# Show help
help:
	@echo "Ducros Stratum Proxy - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build      - Build the binary"
	@echo "  make build-all  - Build for multiple platforms"
	@echo "  make install    - Install to /usr/local/bin"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make run        - Build and run with default config"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Lint code"
	@echo "  make help       - Show this help"
