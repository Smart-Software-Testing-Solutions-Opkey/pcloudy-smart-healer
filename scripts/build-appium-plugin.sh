#!/bin/bash

# SmartHealer Appium Plugin Build Script
# This script builds only the Appium plugin component

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Default values
CLEAN_BUILD=false
INSTALL_DEPS=true
WATCH_MODE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --no-install)
            INSTALL_DEPS=false
            shift
            ;;
        --watch)
            WATCH_MODE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --clean        Clean build artifacts before building"
            echo "  --no-install   Skip dependency installation"
            echo "  --watch        Run in watch mode for development"
            echo "  --help         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Standard build"
            echo "  $0 --clean           # Clean build"
            echo "  $0 --watch           # Development mode with file watching"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_error "Use --help for usage information"
            exit 1
            ;;
    esac
done

print_status "Building SmartHealer Appium Plugin"
print_status "Project root: $PROJECT_ROOT"

# Check required tools
check_dependencies() {
    print_status "Checking build dependencies..."

    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed or not in PATH"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        print_error "npm is not installed or not in PATH"
        exit 1
    fi

    # Check if TypeScript is available
    cd "$PROJECT_ROOT/appium-plugin"
    if ! npx tsc --version &> /dev/null; then
        print_warning "TypeScript not found, will be installed with dependencies"
    fi

    print_success "Dependencies check completed"
}

# Clean build artifacts
clean_build() {
    if [[ "$CLEAN_BUILD" == true ]]; then
        print_status "Cleaning build artifacts..."

        cd "$PROJECT_ROOT/appium-plugin"

        # Clean using npm script if available
        if npm run clean &> /dev/null; then
            print_status "Cleaned using npm clean script"
        else
            # Manual cleanup
            rm -rf lib/ 2>/dev/null || true
            print_status "Manually cleaned build artifacts"
        fi

        # Optionally clean node_modules for fresh install
        if [[ "$INSTALL_DEPS" == true ]]; then
            rm -rf node_modules/ package-lock.json 2>/dev/null || true
            print_status "Cleaned node_modules for fresh install"
        fi

        print_success "Clean completed"
    fi
}

# Install dependencies
install_dependencies() {
    if [[ "$INSTALL_DEPS" == true ]]; then
        print_status "Installing dependencies..."

        cd "$PROJECT_ROOT/appium-plugin"

        # Check if package-lock.json exists for faster npm ci
        if [[ -f "package-lock.json" && "$CLEAN_BUILD" != true ]]; then
            npm ci
            print_status "Dependencies installed using npm ci"
        else
            npm install
            print_status "Dependencies installed using npm install"
        fi

        print_success "Dependencies installation completed"
    else
        print_warning "Skipping dependency installation"
    fi
}

# Verify JavaScript client dependency
verify_js_client() {
    print_status "Verifying JavaScript client dependency..."

    local js_client_path="$PROJECT_ROOT/clients/javascript"

    if [[ ! -d "$js_client_path/dist" ]]; then
        print_warning "JavaScript client not built, attempting to build it..."

        cd "$js_client_path"

        # Check if dependencies are installed
        if [[ ! -d "node_modules" ]]; then
            print_status "Installing JavaScript client dependencies..."
            npm install
        fi

        # Build JavaScript client
        print_status "Building JavaScript client..."
        npm run build

        print_success "JavaScript client built successfully"
    else
        print_status "JavaScript client dependency verified"
    fi
}

# Build Appium plugin
build_plugin() {
    print_status "Building Appium plugin..."

    cd "$PROJECT_ROOT/appium-plugin"

    if [[ "$WATCH_MODE" == true ]]; then
        print_status "Starting TypeScript compiler in watch mode..."
        print_status "Press Ctrl+C to stop watching"
        npm run dev
    else
        print_status "Compiling TypeScript..."
        npm run build
        print_success "Appium plugin compiled successfully"
    fi
}

# Validate build output
validate_build() {
    if [[ "$WATCH_MODE" != true ]]; then
        print_status "Validating build output..."

        cd "$PROJECT_ROOT/appium-plugin"

        # Check if lib directory was created
        if [[ ! -d "lib" ]]; then
            print_error "Build failed: lib directory not created"
            exit 1
        fi

        # Check if main plugin file exists
        if [[ ! -f "lib/plugin.js" ]]; then
            print_error "Build failed: plugin.js not found in lib/"
            exit 1
        fi

        # Check if declaration files exist
        if [[ ! -f "lib/plugin.d.ts" ]]; then
            print_warning "Declaration files not generated"
        fi

        print_success "Build output validation completed"
    fi
}

# Show usage instructions
show_usage_instructions() {
    if [[ "$WATCH_MODE" != true ]]; then
        print_success "Appium plugin build completed successfully!"
        print_status ""
        print_status "Usage instructions:"
        print_status "1. Install the plugin in your Appium server:"
        print_status "   appium plugin install --source=local $PROJECT_ROOT/appium-plugin"
        print_status ""
        print_status "2. Use the plugin in your Appium capabilities:"
        print_status "   {'appium:usePlugins': ['smarthealer']}"
        print_status ""
        print_status "3. Configure SmartHealer (optional):"
        print_status "   {'smarthealer:config': {'openai_key': 'your-key', 'sqlite_db_path': '/path/to/db'}}"
        print_status ""
        print_status "Plugin files location: $PROJECT_ROOT/appium-plugin/lib/"
    fi
}

# Main build process
main() {
    print_status "Starting Appium plugin build process..."

    check_dependencies
    clean_build
    install_dependencies
    verify_js_client
    build_plugin
    validate_build
    show_usage_instructions
}

# Handle script interruption
trap 'print_error "Build interrupted"; exit 1' INT TERM

# Run main function
main "$@"