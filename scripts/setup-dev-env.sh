#!/bin/bash

# SmartHealer Development Environment Setup Script
# This script sets up the development environment and checks all prerequisites

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
INSTALL_MISSING=false
CHECK_ONLY=false
SKIP_BUILD=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install)
            INSTALL_MISSING=true
            shift
            ;;
        --check-only)
            CHECK_ONLY=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --install      Attempt to install missing dependencies"
            echo "  --check-only   Only check prerequisites, don't setup"
            echo "  --skip-build   Skip initial build after setup"
            echo "  --help         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                     # Check and setup development environment"
            echo "  $0 --install           # Install missing dependencies automatically"
            echo "  $0 --check-only        # Only check what's missing"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            print_error "Use --help for usage information"
            exit 1
            ;;
    esac
done

print_status "SmartHealer Development Environment Setup"
print_status "Project root: $PROJECT_ROOT"

# System information
show_system_info() {
    print_status "System Information:"
    echo "  OS: $(uname -s)"
    echo "  Architecture: $(uname -m)"
    echo "  Kernel: $(uname -r)"

    if command -v lsb_release &> /dev/null; then
        echo "  Distribution: $(lsb_release -d | cut -f2)"
    fi
}

# Check Go installation and version
check_go() {
    print_status "Checking Go installation..."

    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go found: version $go_version"

        # Check minimum version (1.19+)
        if [[ $(echo -e "1.19\n$go_version" | sort -V | head -n1) == "1.19" ]]; then
            print_success "Go version is sufficient"
        else
            print_warning "Go version $go_version might be too old (recommended: 1.19+)"
        fi

        # Check GOPATH and GOROOT
        echo "  GOPATH: ${GOPATH:-not set}"
        echo "  GOROOT: ${GOROOT:-$(go env GOROOT)}"

        return 0
    else
        print_error "Go is not installed"
        if [[ "$INSTALL_MISSING" == true ]]; then
            install_go
        else
            print_status "To install Go, visit: https://golang.org/dl/"
        fi
        return 1
    fi
}

# Install Go (basic attempt)
install_go() {
    print_status "Attempting to install Go..."

    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y golang-go
    elif command -v yum &> /dev/null; then
        sudo yum install -y golang
    elif command -v pacman &> /dev/null; then
        sudo pacman -S go
    elif command -v brew &> /dev/null; then
        brew install go
    else
        print_error "Cannot auto-install Go. Please install manually from https://golang.org/dl/"
        return 1
    fi

    print_success "Go installation attempted"
}

# Check Node.js installation and version
check_node() {
    print_status "Checking Node.js installation..."

    if command -v node &> /dev/null; then
        local node_version=$(node --version | sed 's/v//')
        print_success "Node.js found: version $node_version"

        # Check minimum version (16.0.0+)
        if [[ $(echo -e "16.0.0\n$node_version" | sort -V | head -n1) == "16.0.0" ]]; then
            print_success "Node.js version is sufficient"
        else
            print_warning "Node.js version $node_version might be too old (recommended: 16.0.0+)"
        fi

        return 0
    else
        print_error "Node.js is not installed"
        if [[ "$INSTALL_MISSING" == true ]]; then
            install_node
        else
            print_status "To install Node.js, visit: https://nodejs.org/"
        fi
        return 1
    fi
}

# Install Node.js (basic attempt)
install_node() {
    print_status "Attempting to install Node.js..."

    if command -v apt-get &> /dev/null; then
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif command -v yum &> /dev/null; then
        curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
        sudo yum install -y nodejs
    elif command -v pacman &> /dev/null; then
        sudo pacman -S nodejs npm
    elif command -v brew &> /dev/null; then
        brew install node
    else
        print_error "Cannot auto-install Node.js. Please install manually from https://nodejs.org/"
        return 1
    fi

    print_success "Node.js installation attempted"
}

# Check npm and node-gyp
check_npm() {
    print_status "Checking npm and build tools..."

    if command -v npm &> /dev/null; then
        local npm_version=$(npm --version)
        print_success "npm found: version $npm_version"
    else
        print_error "npm is not installed"
        return 1
    fi

    # Check node-gyp
    if command -v node-gyp &> /dev/null; then
        print_success "node-gyp found globally"
    else
        print_warning "node-gyp not found globally (will use npx)"
    fi

    # Check build tools
    if command -v make &> /dev/null; then
        print_success "make found"
    else
        print_warning "make not found (required for building)"
    fi

    if command -v gcc &> /dev/null; then
        print_success "gcc found"
    else
        print_warning "gcc not found (required for native builds)"
    fi

    return 0
}

# Check Python (sometimes needed for node-gyp)
check_python() {
    print_status "Checking Python (needed for native builds)..."

    if command -v python3 &> /dev/null; then
        local python_version=$(python3 --version | awk '{print $2}')
        print_success "Python 3 found: version $python_version"
        return 0
    elif command -v python &> /dev/null; then
        local python_version=$(python --version | awk '{print $2}')
        if [[ "$python_version" == 3.* ]]; then
            print_success "Python found: version $python_version"
            return 0
        else
            print_warning "Python 2 found, Python 3 recommended for node-gyp"
        fi
    else
        print_warning "Python not found (may be needed for native builds)"
        if [[ "$INSTALL_MISSING" == true ]]; then
            install_python
        fi
        return 1
    fi
}

# Install Python (basic attempt)
install_python() {
    print_status "Attempting to install Python..."

    if command -v apt-get &> /dev/null; then
        sudo apt-get install -y python3 python3-pip
    elif command -v yum &> /dev/null; then
        sudo yum install -y python3 python3-pip
    elif command -v pacman &> /dev/null; then
        sudo pacman -S python python-pip
    elif command -v brew &> /dev/null; then
        brew install python
    else
        print_error "Cannot auto-install Python. Please install manually"
        return 1
    fi

    print_success "Python installation attempted"
}

# Check cross-compilation tools
check_cross_compilation() {
    print_status "Checking cross-compilation capabilities..."

    # Check if we can cross-compile Go
    if command -v go &> /dev/null; then
        print_status "Supported Go targets:"
        go tool dist list | grep -E "(linux|darwin)/(amd64|arm64|arm)" | head -10
    fi

    # Check for musl-gcc (for Alpine Linux builds)
    if command -v musl-gcc &> /dev/null; then
        print_success "musl-gcc found (good for Alpine Linux builds)"
    else
        print_warning "musl-gcc not found (optional, for Alpine Linux support)"
    fi
}

# Setup project dependencies
setup_dependencies() {
    if [[ "$CHECK_ONLY" == true ]]; then
        print_status "Skipping dependency setup (check-only mode)"
        return 0
    fi

    print_status "Setting up project dependencies..."

    # Setup JavaScript client
    print_status "Setting up JavaScript client dependencies..."
    cd "$PROJECT_ROOT/clients/javascript"

    if [[ -f "package-lock.json" ]]; then
        npm ci
    else
        npm install
    fi
    print_success "JavaScript client dependencies installed"

    # Setup Appium plugin
    print_status "Setting up Appium plugin dependencies..."
    cd "$PROJECT_ROOT/appium-plugin"

    if [[ -f "package-lock.json" ]]; then
        npm ci
    else
        npm install
    fi
    print_success "Appium plugin dependencies installed"
}

# Run initial build
run_initial_build() {
    if [[ "$CHECK_ONLY" == true || "$SKIP_BUILD" == true ]]; then
        print_status "Skipping initial build"
        return 0
    fi

    print_status "Running initial build to verify setup..."

    # Use our build script
    local build_script="$SCRIPT_DIR/build-all.sh"
    if [[ -x "$build_script" ]]; then
        "$build_script"
        print_success "Initial build completed successfully"
    else
        print_warning "Build script not found or not executable"
        # Fallback to manual build
        print_status "Attempting manual build..."

        # Build Go core
        cd "$PROJECT_ROOT/smarthealer"
        make build-static

        # Build JavaScript client
        cd "$PROJECT_ROOT/clients/javascript"
        npm run build

        # Build Appium plugin
        cd "$PROJECT_ROOT/appium-plugin"
        npm run build

        print_success "Manual build completed"
    fi
}

# Show development workflow
show_workflow() {
    if [[ "$CHECK_ONLY" != true ]]; then
        print_success "Development environment setup completed!"
        print_status ""
        print_status "Development Workflow:"
        print_status "1. Build all components:"
        print_status "   ./scripts/build-all.sh"
        print_status ""
        print_status "2. Build only Appium plugin:"
        print_status "   ./scripts/build-appium-plugin.sh"
        print_status ""
        print_status "3. Development mode (watch):"
        print_status "   ./scripts/build-appium-plugin.sh --watch"
        print_status ""
        print_status "4. Cross-compile for different architectures:"
        print_status "   ./scripts/build-all.sh --arch linux-arm64"
        print_status ""
        print_status "Available architectures: linux-x64, linux-arm64, linux-arm, darwin-x64, darwin-arm64"
    fi
}

# Main setup process
main() {
    print_status "Starting development environment setup..."

    local has_errors=false

    show_system_info
    echo

    # Check all prerequisites
    check_go || has_errors=true
    echo

    check_node || has_errors=true
    echo

    check_npm || has_errors=true
    echo

    check_python || has_errors=true
    echo

    check_cross_compilation
    echo

    if [[ "$has_errors" == true && "$INSTALL_MISSING" != true ]]; then
        print_error "Some prerequisites are missing. Run with --install to attempt automatic installation."
        exit 1
    fi

    setup_dependencies
    run_initial_build
    show_workflow
}

# Handle script interruption
trap 'print_error "Setup interrupted"; exit 1' INT TERM

# Run main function
main "$@"