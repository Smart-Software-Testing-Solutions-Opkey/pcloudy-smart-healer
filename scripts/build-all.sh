#!/bin/bash

# SmartHealer Multi-Architecture Build Script
# This script builds all components for the current or specified architecture

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Architecture detection
detect_arch() {
    local os_name=$(uname -s)
    local arch_name=$(uname -m)

    case "$os_name" in
        Linux*)
            case "$arch_name" in
                x86_64) echo "linux-x64" ;;
                aarch64|arm64) echo "linux-arm64" ;;
                armv7l) echo "linux-arm" ;;
                *) echo "linux-unknown" ;;
            esac
            ;;
        Darwin*)
            case "$arch_name" in
                x86_64) echo "darwin-x64" ;;
                arm64) echo "darwin-arm64" ;;
                *) echo "darwin-unknown" ;;
            esac
            ;;
        *) echo "unknown-unknown" ;;
    esac
}

# Default values
TARGET_ARCH=$(detect_arch)
BUILD_TYPE="release"
CLEAN_BUILD=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        --debug)
            BUILD_TYPE="debug"
            shift
            ;;
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --arch ARCH     Target architecture (detected: $(detect_arch))"
            echo "                  Supported: linux-x64, linux-arm64, linux-arm, darwin-x64, darwin-arm64"
            echo "  --debug         Build in debug mode (default: release)"
            echo "  --clean         Clean build artifacts before building"
            echo "  --verbose       Enable verbose output"
            echo "  --help          Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                           # Build for current architecture"
            echo "  $0 --arch linux-arm64        # Cross-compile for ARM64"
            echo "  $0 --clean --debug           # Clean debug build"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_error "Use --help for usage information"
            exit 1
            ;;
    esac
done

print_status "Building SmartHealer for architecture: $TARGET_ARCH"
print_status "Build type: $BUILD_TYPE"
print_status "Project root: $PROJECT_ROOT"

# Validate architecture
case "$TARGET_ARCH" in
    linux-x64|linux-arm64|linux-arm|darwin-x64|darwin-arm64)
        print_status "Architecture $TARGET_ARCH is supported"
        ;;
    *)
        print_error "Unsupported architecture: $TARGET_ARCH"
        print_error "Supported architectures: linux-x64, linux-arm64, linux-arm, darwin-x64, darwin-arm64"
        exit 1
        ;;
esac

# Check required tools
check_dependencies() {
    print_status "Checking build dependencies..."

    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check Node.js
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed or not in PATH"
        exit 1
    fi

    # Check npm
    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed or not in PATH"
        exit 1
    fi

    # Check node-gyp
    if ! command -v node-gyp &> /dev/null; then
        print_warning "node-gyp not found globally, will use npx"
    fi

    print_success "All dependencies are available"
}

# Clean build artifacts
clean_build() {
    if [[ "$CLEAN_BUILD" == true ]]; then
        print_status "Cleaning build artifacts..."

        # Clean Go artifacts
        cd "$PROJECT_ROOT/smarthealer"
        make clean 2>/dev/null || true

        # Clean JavaScript client
        cd "$PROJECT_ROOT/clients/javascript"
        npm run clean 2>/dev/null || true
        rm -rf build/ node_modules/ 2>/dev/null || true

        # Clean Appium plugin
        cd "$PROJECT_ROOT/appium-plugin"
        npm run clean 2>/dev/null || true
        rm -rf lib/ node_modules/ 2>/dev/null || true

        print_success "Clean completed"
    fi
}

# Build Go core library
build_go_core() {
    print_status "Building Go core library for $TARGET_ARCH..."

    cd "$PROJECT_ROOT/smarthealer"

    # Set up cross-compilation environment
    case "$TARGET_ARCH" in
        linux-x64)
            export GOOS=linux
            export GOARCH=amd64
            ;;
        linux-arm64)
            export GOOS=linux
            export GOARCH=arm64
            ;;
        linux-arm)
            export GOOS=linux
            export GOARCH=arm
            export GOARM=7
            ;;
        darwin-x64)
            export GOOS=darwin
            export GOARCH=amd64
            ;;
        darwin-arm64)
            export GOOS=darwin
            export GOARCH=arm64
            ;;
    esac

    export CGO_ENABLED=1

    # Create output directory
    OUTPUT_DIR="../clients/javascript/lib/$TARGET_ARCH"
    mkdir -p "$OUTPUT_DIR"

    # Build static library for linking with Node.js addon
    if [[ "$BUILD_TYPE" == "debug" ]]; then
        go build -buildmode=c-archive -o "$OUTPUT_DIR/libsmarthealer.a" cmd/main.go
    else
        go build -ldflags '-s -w' -buildmode=c-archive -o "$OUTPUT_DIR/libsmarthealer.a" cmd/main.go
    fi

    # Copy header file
    cp "$OUTPUT_DIR/libsmarthealer.h" "../javascript/includes/" 2>/dev/null || true

    print_success "Go core library built successfully"
}

# Build JavaScript client
build_js_client() {
    print_status "Building JavaScript client for $TARGET_ARCH..."

    cd "$PROJECT_ROOT/clients/javascript"

    # Install dependencies
    print_status "Installing JavaScript client dependencies..."
    npm install

    # Update binding.gyp for target architecture
    update_binding_gyp

    # Compile native addon
    print_status "Compiling native addon..."
    if command -v node-gyp &> /dev/null; then
        node-gyp rebuild
    else
        npx node-gyp rebuild
    fi

    # Compile TypeScript
    print_status "Compiling TypeScript..."
    npm run build

    print_success "JavaScript client built successfully"
}

# Update binding.gyp for target architecture
update_binding_gyp() {
    local binding_file="$PROJECT_ROOT/clients/javascript/binding.gyp"
    local temp_file=$(mktemp)

    # Create architecture-specific binding.gyp
    cat > "$temp_file" << EOF
{
  "targets": [
    {
      "target_name": "smarthealer",
      "sources": ["src/smarthealer.cc"],
      "include_dirs": [
        "<!(node -p \"require('node-addon-api').include\")",
        "includes"
      ],
      "dependencies": [
        "<!(node -p \"require('node-addon-api').targets\"):node_addon_api"
      ],
      "cflags!": [ "-fno-exceptions" ],
      "cflags_cc!": [ "-fno-exceptions" ],
      "cflags": [ "-fexceptions" ],
      "cflags_cc": [ "-fexceptions" ],
      "conditions": [
EOF

    case "$TARGET_ARCH" in
        linux-*)
            cat >> "$temp_file" << EOF
          ["OS=='linux'", {
              "libraries": ["<(module_root_dir)/lib/$TARGET_ARCH/libsmarthealer.a"]
          }]
EOF
            ;;
        darwin-*)
            cat >> "$temp_file" << EOF
          ["OS=='mac'", {
              "libraries": ["<(module_root_dir)/lib/$TARGET_ARCH/libsmarthealer.a"]
          }]
EOF
            ;;
    esac

    cat >> "$temp_file" << EOF
      ]
    }
  ]
}
EOF

    mv "$temp_file" "$binding_file"
}

# Build Appium plugin
build_appium_plugin() {
    print_status "Building Appium plugin..."

    cd "$PROJECT_ROOT/appium-plugin"

    # Install dependencies
    print_status "Installing Appium plugin dependencies..."
    npm install

    # Compile TypeScript
    print_status "Compiling TypeScript..."
    npm run build

    print_success "Appium plugin built successfully"
}

# Main build process
main() {
    print_status "Starting SmartHealer build process..."

    check_dependencies
    clean_build
    build_go_core
    build_js_client
    build_appium_plugin

    print_success "Build completed successfully for $TARGET_ARCH!"
    print_status "Built components:"
    print_status "  - Go core library: clients/javascript/lib/$TARGET_ARCH/libsmarthealer.a"
    print_status "  - JavaScript client: clients/javascript/dist/"
    print_status "  - Appium plugin: appium-plugin/lib/"
}

# Handle script interruption
trap 'print_error "Build interrupted"; exit 1' INT TERM

# Run main function
main "$@"