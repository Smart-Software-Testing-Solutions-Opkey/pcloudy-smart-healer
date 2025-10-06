# SmartHealer Multi-Architecture Build System
# Top-level Makefile for building all components

.PHONY: all build clean setup help appium-plugin js-client go-core cross-compile

# Default target
all: build

# Build all components for current architecture
build:
	@echo "Building SmartHealer for current architecture..."
	@./scripts/build-all.sh

# Clean all build artifacts
clean:
	@echo "Cleaning all build artifacts..."
	@cd smarthealer && make clean || true
	@cd clients/javascript && npm run clean 2>/dev/null || true
	@cd appium-plugin && npm run clean 2>/dev/null || true
	@rm -rf clients/javascript/lib/*/
	@echo "Clean completed"

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@./scripts/setup-dev-env.sh

# Install dependencies and setup with automatic installation
setup-install:
	@echo "Setting up development environment with automatic installation..."
	@./scripts/setup-dev-env.sh --install

# Build only the Appium plugin
appium-plugin:
	@echo "Building Appium plugin..."
	@./scripts/build-appium-plugin.sh

# Build only the JavaScript client
js-client:
	@echo "Building JavaScript client..."
	@cd clients/javascript && npm install && npm run build

# Build only the Go core library
go-core:
	@echo "Building Go core library..."
	@cd smarthealer && make build-static

# Cross-compile for all supported architectures
cross-compile:
	@echo "Cross-compiling for all supported architectures..."
	@./scripts/build-all.sh --arch linux-x64
	@./scripts/build-all.sh --arch linux-arm64
	@./scripts/build-all.sh --arch linux-arm
	@./scripts/build-all.sh --arch darwin-x64
	@./scripts/build-all.sh --arch darwin-arm64
	@echo "Cross-compilation completed"

# Development mode with file watching
dev:
	@echo "Starting development mode..."
	@./scripts/build-appium-plugin.sh --watch

# Build with debug symbols
debug:
	@echo "Building in debug mode..."
	@./scripts/build-all.sh --debug

# Clean build (remove all artifacts before building)
rebuild: clean build

# Install Appium plugin locally
install-plugin: appium-plugin
	@echo "Installing SmartHealer plugin in Appium..."
	@appium plugin install --source=local ./appium-plugin

# Run tests (if available)
test:
	@echo "Running tests..."
	@cd clients/javascript && npm test || echo "No JavaScript tests found"
	@cd appium-plugin && npm test || echo "No Appium plugin tests found"
	@cd smarthealer && go test ./... || echo "No Go tests found"

# Show help
help:
	@echo "SmartHealer Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build all components for current architecture"
	@echo "  clean          Clean all build artifacts"
	@echo "  setup          Setup development environment"
	@echo "  setup-install  Setup with automatic dependency installation"
	@echo "  appium-plugin  Build only the Appium plugin"
	@echo "  js-client      Build only the JavaScript client"
	@echo "  go-core        Build only the Go core library"
	@echo "  cross-compile  Build for all supported architectures"
	@echo "  dev            Start development mode with file watching"
	@echo "  debug          Build in debug mode"
	@echo "  rebuild        Clean and build"
	@echo "  install-plugin Install Appium plugin locally"
	@echo "  test           Run tests"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Architecture-specific builds:"
	@echo "  make build ARCH=linux-arm64    # Build for ARM64 Linux"
	@echo "  make build ARCH=darwin-x64     # Build for x64 macOS"
	@echo ""
	@echo "Examples:"
	@echo "  make setup          # First time setup"
	@echo "  make build          # Standard build"
	@echo "  make dev            # Development mode"
	@echo "  make cross-compile  # Build for all architectures"

# Architecture-specific build (when ARCH is provided)
ifdef ARCH
build:
	@echo "Building SmartHealer for architecture: $(ARCH)"
	@./scripts/build-all.sh --arch $(ARCH)
endif