# SmartHealer Build Scripts

This directory contains build scripts for creating SmartHealer components across different architectures.

## Scripts Overview

### `build-all.sh`
Complete build script that compiles all SmartHealer components for a specified architecture.

**Usage:**
```bash
./build-all.sh [OPTIONS]

Options:
  --arch ARCH     Target architecture (auto-detected if not specified)
  --debug         Build in debug mode (default: release)
  --clean         Clean build artifacts before building
  --verbose       Enable verbose output
  --help          Show help message
```

**Supported Architectures:**
- `linux-x64` - Linux x86_64
- `linux-arm64` - Linux ARM64
- `linux-arm` - Linux ARM (ARMv7)
- `darwin-x64` - macOS Intel
- `darwin-arm64` - macOS Apple Silicon

**Examples:**
```bash
# Build for current architecture
./build-all.sh

# Cross-compile for ARM64 Linux
./build-all.sh --arch linux-arm64

# Clean debug build
./build-all.sh --clean --debug

# Verbose build for macOS
./build-all.sh --arch darwin-x64 --verbose
```

### `build-appium-plugin.sh`
Specialized script for building only the Appium plugin component.

**Usage:**
```bash
./build-appium-plugin.sh [OPTIONS]

Options:
  --clean        Clean build artifacts before building
  --no-install   Skip dependency installation
  --watch        Run in watch mode for development
  --help         Show help message
```

**Examples:**
```bash
# Standard build
./build-appium-plugin.sh

# Development mode with file watching
./build-appium-plugin.sh --watch

# Clean build without reinstalling dependencies
./build-appium-plugin.sh --clean --no-install
```

### `setup-dev-env.sh`
Development environment setup and validation script.

**Usage:**
```bash
./setup-dev-env.sh [OPTIONS]

Options:
  --install      Attempt to install missing dependencies
  --check-only   Only check prerequisites, don't setup
  --skip-build   Skip initial build after setup
  --help         Show help message
```

**Examples:**
```bash
# Check and setup development environment
./setup-dev-env.sh

# Auto-install missing dependencies
./setup-dev-env.sh --install

# Only check what's installed
./setup-dev-env.sh --check-only
```

## Build Process Flow

1. **Environment Check**: Validates required tools (Go, Node.js, npm, build tools)
2. **Dependency Installation**: Installs npm dependencies for JavaScript components
3. **Go Core Build**: Compiles the Go library as a static archive (.a file)
4. **JavaScript Client Build**: Compiles native Node.js addon and TypeScript
5. **Appium Plugin Build**: Compiles TypeScript to JavaScript

## Architecture-Specific Building

The build system supports cross-compilation for multiple architectures:

### Linux Targets
- **linux-x64**: Standard Linux on x86_64
- **linux-arm64**: Linux on ARM64 (e.g., AWS Graviton, Raspberry Pi 4+)
- **linux-arm**: Linux on ARMv7 (e.g., older Raspberry Pi)

### macOS Targets
- **darwin-x64**: macOS on Intel processors
- **darwin-arm64**: macOS on Apple Silicon (M1/M2)

## Output Structure

After building, artifacts are organized as follows:

```
clients/javascript/lib/
├── linux-x64/
│   ├── libsmarthealer.a      # Static library for Linux x64
│   └── libsmarthealer.h      # Header file
├── linux-arm64/
│   ├── libsmarthealer.a      # Static library for Linux ARM64
│   └── libsmarthealer.h      # Header file
└── [other architectures...]

clients/javascript/dist/      # Compiled TypeScript
├── index.js
├── index.d.ts
└── ...

appium-plugin/lib/           # Compiled Appium plugin
├── plugin.js
├── plugin.d.ts
└── ...
```

## Development Workflow

### First Time Setup
```bash
# Setup development environment
./setup-dev-env.sh --install

# Or use the Makefile
make setup-install
```

### Regular Development
```bash
# Build everything
./build-all.sh

# Or just the plugin in watch mode
./build-appium-plugin.sh --watch

# Using Makefile
make dev
```

### Cross-Platform Building
```bash
# Build for specific architecture
./build-all.sh --arch linux-arm64

# Build for all architectures
make cross-compile
```

## Troubleshooting

### Common Issues

1. **Go not found**
   ```bash
   # Install Go (Ubuntu/Debian)
   sudo apt install golang-go

   # Or download from https://golang.org/dl/
   ```

2. **Node.js version too old**
   ```bash
   # Install Node.js 16+ (Ubuntu/Debian)
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs
   ```

3. **Build tools missing**
   ```bash
   # Install build essentials (Ubuntu/Debian)
   sudo apt install build-essential

   # Install Xcode command line tools (macOS)
   xcode-select --install
   ```

4. **Cross-compilation issues**
   - Ensure Go supports target architecture: `go tool dist list`
   - Check CGO requirements for target platform
   - Verify cross-compilation toolchain is installed

### Debug Mode

Use debug builds for troubleshooting:
```bash
./build-all.sh --debug --verbose
```

This enables:
- Debug symbols in Go binaries
- Verbose output from all build steps
- Detailed error messages
- Source maps for TypeScript

### Clean Builds

If experiencing build issues, try a clean build:
```bash
./build-all.sh --clean
# or
make rebuild
```

## Integration with IDEs

### VS Code
Add these tasks to `.vscode/tasks.json`:
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Build All",
            "type": "shell",
            "command": "./scripts/build-all.sh",
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always"
            }
        },
        {
            "label": "Build Plugin (Watch)",
            "type": "shell",
            "command": "./scripts/build-appium-plugin.sh",
            "args": ["--watch"],
            "group": "build",
            "isBackground": true
        }
    ]
}
```

## Performance Tips

1. **Use clean builds sparingly** - Only when necessary, as they're slower
2. **Leverage watch mode** - For rapid development of the Appium plugin
3. **Parallel builds** - The scripts use parallel execution where possible
4. **Cache dependencies** - npm ci is used when package-lock.json exists