.PHONY: all build build-static build-shared clean rebuild cross help

# Library configuration
LIB_NAME = smarthealer
STATIC_LIB = lib$(LIB_NAME).a
SHARED_LIB = lib$(LIB_NAME).so
HEADER = lib$(LIB_NAME).h
GO_SRC = cmd/main.go

# Detect current platform
DETECTED_OS := $(shell uname -s | tr A-Z a-z)
DETECTED_ARCH := $(shell uname -m)

# Normalize architecture names to GOARCH
ifeq ($(DETECTED_ARCH),x86_64)
	DETECTED_ARCH = amd64
endif
ifeq ($(DETECTED_ARCH),aarch64)
	DETECTED_ARCH = arm64
endif

# Default target
all: build-static

# Build static library for current platform
build-static:
	@echo "Building static library for $(DETECTED_OS)-$(DETECTED_ARCH)..."
	go build -ldflags '-s -w' -buildmode=c-archive -o $(STATIC_LIB) $(GO_SRC)
	@echo "Built: $(STATIC_LIB)"

# Build shared library for current platform
build-shared:
	@echo "Building shared library for $(DETECTED_OS)-$(DETECTED_ARCH)..."
	go build -ldflags '-s -w' -buildmode=c-shared -o $(SHARED_LIB) $(GO_SRC)
	@echo "Built: $(SHARED_LIB)"

# Alias for backward compatibility
build: build-static

# Cross-compilation for different OS/ARCH combinations
# Usage: make cross OS=linux ARCH=amd64 TYPE=static
# Usage: make cross OS=darwin ARCH=arm64 TYPE=shared
cross:
	@if [ -z "$(OS)" ] || [ -z "$(ARCH)" ] || [ -z "$(TYPE)" ]; then \
		echo "Error: OS, ARCH, and TYPE must be specified"; \
		echo "Usage: make cross OS=linux ARCH=amd64 TYPE=static"; \
		echo "       make cross OS=darwin ARCH=arm64 TYPE=shared"; \
		echo ""; \
		echo "Supported combinations:"; \
		echo "  OS: linux, darwin"; \
		echo "  ARCH: amd64, arm64"; \
		echo "  TYPE: static, shared"; \
		exit 1; \
	fi
	@echo "Cross-compiling for $(OS)-$(ARCH) ($(TYPE))..."
	@if [ "$(TYPE)" = "static" ]; then \
		GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=1 go build -ldflags '-s -w' -buildmode=c-archive -o lib$(LIB_NAME)-$(OS)-$(ARCH).a $(GO_SRC); \
		echo "Built: lib$(LIB_NAME)-$(OS)-$(ARCH).a"; \
	elif [ "$(TYPE)" = "shared" ]; then \
		GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=1 go build -ldflags '-s -w' -buildmode=c-shared -o lib$(LIB_NAME)-$(OS)-$(ARCH).so $(GO_SRC); \
		echo "Built: lib$(LIB_NAME)-$(OS)-$(ARCH).so"; \
	else \
		echo "Error: TYPE must be 'static' or 'shared'"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(STATIC_LIB) $(SHARED_LIB) $(HEADER)
	rm -f lib$(LIB_NAME)-*.a lib$(LIB_NAME)-*.so lib$(LIB_NAME)-*.h
	@echo "Clean completed"

# Rebuild
rebuild: clean build-static

# Help
help:
	@echo "SmartHealer Go Library Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make                    Build static library for current platform"
	@echo "  make build-static       Build static library for current platform"
	@echo "  make build-shared       Build shared library for current platform"
	@echo "  make cross OS=<os> ARCH=<arch> TYPE=<type>"
	@echo "                          Cross-compile for specific platform"
	@echo "  make clean              Remove all build artifacts"
	@echo "  make rebuild            Clean and rebuild"
	@echo ""
	@echo "Cross-compilation examples:"
	@echo "  make cross OS=linux ARCH=amd64 TYPE=static"
	@echo "  make cross OS=linux ARCH=arm64 TYPE=static"
	@echo "  make cross OS=darwin ARCH=amd64 TYPE=static"
	@echo "  make cross OS=darwin ARCH=arm64 TYPE=static"
	@echo ""
	@echo "Supported platforms:"
	@echo "  OS: linux, darwin"
	@echo "  ARCH: amd64, arm64"
	@echo "  TYPE: static, shared"
	@echo ""
	@echo "Current platform: $(DETECTED_OS)-$(DETECTED_ARCH)"
