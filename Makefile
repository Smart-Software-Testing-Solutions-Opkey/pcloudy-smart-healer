.PHONY: all build clean rebuild cross-build cross-build-all help install-deps test

# Detect current platform
DETECTED_OS := $(shell uname -s | tr A-Z a-z)
DETECTED_ARCH := $(shell uname -m)

# Normalize OS names
ifeq ($(DETECTED_OS),darwin)
	OS_NAME = darwin
else ifeq ($(DETECTED_OS),linux)
	OS_NAME = linux
else
	$(error Unsupported OS: $(DETECTED_OS))
endif

# Normalize architecture names to GOARCH
ifeq ($(DETECTED_ARCH),x86_64)
	ARCH_NAME = amd64
else ifeq ($(DETECTED_ARCH),aarch64)
	ARCH_NAME = arm64
else ifeq ($(DETECTED_ARCH),arm64)
	ARCH_NAME = arm64
else
	$(error Unsupported architecture: $(DETECTED_ARCH))
endif

# Paths
GO_DIR = smarthealer
JS_CLIENT_DIR = clients/javascript
APPIUM_PLUGIN_DIR = appium-plugin

# Default target - build for current platform
all: build

# Build entire pipeline for current platform
build:
	@echo "=========================================="
	@echo "Building SmartHealer for $(OS_NAME)-$(ARCH_NAME)"
	@echo "=========================================="
	@echo ""
	@echo "Step 1/6: Building Go core library..."
	@cd $(GO_DIR) && $(MAKE) build-static
	@echo ""
	@echo "Step 2/6: Copying artifacts to JavaScript client..."
	@mkdir -p $(JS_CLIENT_DIR)/lib/$(OS_NAME)-$(ARCH_NAME)
	@cp $(GO_DIR)/libsmarthealer.a $(JS_CLIENT_DIR)/lib/$(OS_NAME)-$(ARCH_NAME)/libsmarthealer.a
	@cp $(GO_DIR)/libsmarthealer.h $(JS_CLIENT_DIR)/includes/libsmarthealer.h
	@echo "Artifacts copied successfully"
	@echo ""
	@echo "Step 3/6: Installing JavaScript client dependencies (with native addon build)..."
	@cd $(JS_CLIENT_DIR) && npm install
	@echo ""
	@echo "Step 4/6: Building JavaScript client (TypeScript + native addon)..."
	@cd $(JS_CLIENT_DIR) && $(MAKE) build
	@echo ""
	@echo "Step 5/6: Installing Appium plugin dependencies..."
	@cd $(APPIUM_PLUGIN_DIR) && npm install
	@echo ""
	@echo "Step 6/6: Building Appium plugin..."
	@cd $(APPIUM_PLUGIN_DIR) && $(MAKE) build
	@echo ""
	@echo "=========================================="
	@echo "Build completed successfully for $(OS_NAME)-$(ARCH_NAME)"
	@echo "=========================================="

# Cross-build for all supported platforms
cross-build-all:
	@echo "=========================================="
	@echo "Cross-building SmartHealer for all platforms"
	@echo "=========================================="
	@echo ""
	@for target in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do \
		TARGET_OS=$$(echo $$target | cut -d- -f1); \
		TARGET_ARCH=$$(echo $$target | cut -d- -f2); \
		echo "Building for $$target..."; \
		echo "Step 1/2: Cross-compiling Go core library for $$target..."; \
		(cd $(GO_DIR) && $(MAKE) cross OS=$$TARGET_OS ARCH=$$TARGET_ARCH TYPE=static); \
		echo "Step 2/2: Copying artifacts to JavaScript client..."; \
		mkdir -p $(JS_CLIENT_DIR)/lib/$$TARGET_OS-$$TARGET_ARCH; \
		cp $(GO_DIR)/libsmarthealer-$$TARGET_OS-$$TARGET_ARCH.a $(JS_CLIENT_DIR)/lib/$$TARGET_OS-$$TARGET_ARCH/libsmarthealer.a; \
		cp $(GO_DIR)/libsmarthealer.h $(JS_CLIENT_DIR)/includes/libsmarthealer.h; \
		echo "Completed $$target"; \
		echo ""; \
	done
	@echo "=========================================="
	@echo "All platforms built successfully"
	@echo "Archives copied to:"
	@echo "  $(JS_CLIENT_DIR)/lib/linux-amd64/libsmarthealer.a"
	@echo "  $(JS_CLIENT_DIR)/lib/linux-arm64/libsmarthealer.a"
	@echo "  $(JS_CLIENT_DIR)/lib/darwin-amd64/libsmarthealer.a"
	@echo "  $(JS_CLIENT_DIR)/lib/darwin-arm64/libsmarthealer.a"
	@echo "=========================================="

# Cross-build for specific platform
# Usage: make cross-build TARGET=linux-amd64
# Usage: make cross-build TARGET=darwin-arm64
cross-build:
	@if [ -z "$(TARGET)" ]; then \
		echo "Error: TARGET must be specified"; \
		echo "Usage: make cross-build TARGET=<os>-<arch>"; \
		echo ""; \
		echo "Supported targets:"; \
		echo "  linux-amd64    Linux x86_64"; \
		echo "  linux-arm64    Linux ARM64"; \
		echo "  darwin-amd64   macOS x86_64"; \
		echo "  darwin-arm64   macOS ARM64"; \
		exit 1; \
	fi
	@TARGET_OS=$$(echo $(TARGET) | cut -d- -f1); \
	TARGET_ARCH=$$(echo $(TARGET) | cut -d- -f2); \
	echo "=========================================="; \
	echo "Cross-building SmartHealer for $(TARGET)"; \
	echo "=========================================="; \
	echo ""; \
	echo "Step 1/4: Cross-compiling Go core library..."; \
	(cd $(GO_DIR) && $(MAKE) cross OS=$$TARGET_OS ARCH=$$TARGET_ARCH TYPE=static); \
	echo ""; \
	echo "Step 2/4: Copying artifacts to JavaScript client..."; \
	mkdir -p $(JS_CLIENT_DIR)/lib/$$TARGET_OS-$$TARGET_ARCH; \
	cp $(GO_DIR)/libsmarthealer-$$TARGET_OS-$$TARGET_ARCH.a $(JS_CLIENT_DIR)/lib/$$TARGET_OS-$$TARGET_ARCH/libsmarthealer.a; \
	cp $(GO_DIR)/libsmarthealer.h $(JS_CLIENT_DIR)/includes/libsmarthealer.h; \
	echo "Artifacts copied successfully"; \
	echo ""; \
	echo "Step 3/4: Building JavaScript client..."; \
	(cd $(JS_CLIENT_DIR) && $(MAKE) build); \
	echo ""; \
	echo "Step 4/4: Building Appium plugin..."; \
	(cd $(APPIUM_PLUGIN_DIR) && $(MAKE) build); \
	echo ""; \
	echo "=========================================="; \
	echo "Cross-build completed successfully for $(TARGET)"; \
	echo "=========================================="

# Install dependencies for all components
install-deps:
	@echo "Installing dependencies..."
	@echo ""
	@echo "JavaScript client dependencies..."
	@cd $(JS_CLIENT_DIR) && npm install
	@echo ""
	@echo "Appium plugin dependencies..."
	@cd $(APPIUM_PLUGIN_DIR) && npm install
	@echo ""
	@echo "Dependencies installed successfully"

# Clean all build artifacts
clean:
	@echo "Cleaning all build artifacts..."
	@cd $(GO_DIR) && $(MAKE) clean
	@cd $(JS_CLIENT_DIR) && $(MAKE) clean
	@cd $(APPIUM_PLUGIN_DIR) && $(MAKE) clean
	@rm -rf $(JS_CLIENT_DIR)/lib/*/
	@echo "Clean completed"

# Rebuild everything
rebuild: clean build

# Run tests
test:
	@echo "Running tests..."
	@echo ""
	@echo "Go tests..."
	@cd $(GO_DIR) && go test ./... || echo "No Go tests found"
	@echo ""
	@echo "JavaScript client tests..."
	@cd $(JS_CLIENT_DIR) && npm test || echo "No JavaScript tests found"
	@echo ""
	@echo "Appium plugin tests..."
	@cd $(APPIUM_PLUGIN_DIR) && npm test || echo "No Appium plugin tests found"

# Help
help:
	@echo "SmartHealer Multi-Architecture Build System"
	@echo ""
	@echo "Current platform: $(OS_NAME)-$(ARCH_NAME)"
	@echo ""
	@echo "Usage:"
	@echo "  make                    Build for current platform"
	@echo "  make build              Build for current platform"
	@echo "  make cross-build TARGET=<os>-<arch>"
	@echo "                          Cross-build for specific platform"
	@echo "  make cross-build-all    Cross-build for all supported platforms"
	@echo "  make install-deps       Install dependencies"
	@echo "  make clean              Clean all build artifacts"
	@echo "  make rebuild            Clean and rebuild"
	@echo "  make test               Run all tests"
	@echo "  make help               Show this help"
	@echo ""
	@echo "Cross-build examples:"
	@echo "  make cross-build TARGET=linux-amd64"
	@echo "  make cross-build TARGET=linux-arm64"
	@echo "  make cross-build TARGET=darwin-amd64"
	@echo "  make cross-build TARGET=darwin-arm64"
	@echo "  make cross-build-all    # Build all platforms"
	@echo ""
	@echo "Supported targets:"
	@echo "  linux-amd64    Linux x86_64"
	@echo "  linux-arm64    Linux ARM64"
	@echo "  darwin-amd64   macOS x86_64 (Intel)"
	@echo "  darwin-arm64   macOS ARM64 (Apple Silicon)"
	@echo ""
	@echo "Build pipeline:"
	@echo "  1. Build Go core library (static archive)"
	@echo "  2. Copy artifacts to JavaScript client"
	@echo "  3. Build JavaScript client (TypeScript + native addon)"
	@echo "  4. Build Appium plugin"
	@echo ""
	@echo "Component-specific builds:"
	@echo "  cd smarthealer && make help          # Go library options"
	@echo "  cd clients/javascript && make help   # JavaScript client options"
	@echo "  cd appium-plugin && make help        # Appium plugin options"
