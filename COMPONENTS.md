# SmartHealer Components Documentation

## Overview

SmartHealer is an AI-powered element locator resolution and healing system for web automation. The repository contains three main components that work together to provide intelligent element healing capabilities for automation frameworks.

## Repository Structure

```
├── smarthealer/           # Core Go library
├── clients/javascript/    # JavaScript/TypeScript client
├── appium-plugin/        # Appium 2.0+ plugin
├── scripts/              # Build and setup scripts
└── Makefile              # Top-level build system
```

## Core Components

### 1. SmartHealer Core (Go Library) 📁 `smarthealer/`

**Purpose**: The main SmartHealer engine written in Go that provides the core AI-powered element locator resolution functionality.

**Key Features**:
- AI/LLM integration for intelligent element analysis
- SQLite-based data persistence for storing page data and locator mappings
- Background worker processes for async healing tasks
- Sync and async locator resolution APIs
- Cross-platform shared library (.so) compilation

**Key Files**:
- `smarthealer.go:24`: Main `SmartHealer` struct with initialization and resolution methods
- `intelligence/`: AI integration for element analysis and locator suggestions
- `healer/`: Core healing logic and background worker processes
- `store/`: Data persistence layer with SQLite implementation
- `retrieval/`: Page analysis and element retrieval logic
- `config/`: Configuration management for database paths and AI API keys

**Usage**:
```bash
cd smarthealer
make build         # Build shared library
make build-static  # Build static archive
make clean         # Clean build artifacts
```

**Dependencies**: Go 1.25.1+, OpenAI API key, SQLite

### 2. JavaScript Client 📁 `clients/javascript/`

**Purpose**: Node.js bindings for the Go library, providing a TypeScript/JavaScript API for integrating SmartHealer into Node.js applications.

**Key Features**:
- TypeScript interfaces with full type safety
- Native Node.js bindings via `node-gyp`
- Error handling with custom `SmartHealerError` class
- Promise-based async API
- Input validation for configurations and parameters

**Key Files**:
- `src/index.ts:35`: Main `SmartHealer` class with static methods
- `src/types.ts`: TypeScript interfaces and enums
- `binding.gyp`: Node.js native addon configuration

**Usage**:
```bash
cd clients/javascript
npm install        # Build native bindings
npm run build      # Compile TypeScript
npm run dev        # Watch mode for development
```

**API Example**:
```javascript
const { SmartHealer } = require('smarthealer-js');

// Initialize
await SmartHealer.init({
  openai_key: 'your-openai-key',
  sqlite_db_path: '/path/to/database.db'
});

// Resolve locator
const result = await SmartHealer.resolveLocator(info, options);
```

**Requirements**: Node.js 16.0.0+, Linux x64 only

### 3. Appium Plugin 📁 `appium-plugin/`

**Purpose**: An Appium 2.0+ plugin that integrates SmartHealer into Appium test automation, providing automatic element healing during test execution.

**Key Features**:
- **Learning Mode**: Successful element finds trigger async resolution to improve knowledge base
- **Healing Mode**: Failed finds attempt sync resolution with alternative locators
- Intercepts `findElement`/`findElements` commands transparently
- Configuration via Appium capabilities or direct API calls
- Automatic screenshot and page source capture for context

**Key Files**:
- `src/plugin.ts:6`: Main `SmartHealerPlugin` class extending `BasePlugin`
- `src/smarthealer-manager.ts`: Singleton manager for SmartHealer instances
- `src/types.ts`: Plugin-specific type definitions

**Usage**:
```bash
cd appium-plugin
npm install
npm run build
appium plugin install --source=local ./appium-plugin
```

**Configuration**:
```javascript
// Via Appium capabilities
const caps = {
  'smarthealer:config': {
    openai_key: 'your-openai-key',
    sqlite_db_path: '/path/to/database.db'
  }
};
```

**Workflow**:
1. **Success Path**: Element found → Async resolution → Knowledge base update
2. **Failure Path**: Element not found → Sync resolution → Try alternative locator → Return element or original error

## Build System

The repository includes a comprehensive build system supporting multi-architecture compilation:

**Main Commands**:
```bash
make build          # Build all components for current architecture
make clean          # Clean all build artifacts
make setup          # Setup development environment
make cross-compile  # Build for all supported architectures (linux-x64, linux-arm64, darwin-x64, darwin-arm64)
make dev            # Development mode with file watching
make install-plugin # Install Appium plugin locally
make test           # Run tests across all components
```

**Architecture Support**:
- Linux x64/ARM64
- macOS x64/ARM64 (Go core only)
- Multi-architecture compilation via build scripts

## Configuration Requirements

All components require:
- **OpenAI API Key**: For AI-powered element analysis
- **SQLite Database Path**: For storing page data and locator mappings (defaults to `~/.smarthealer/smarthealer.db`)

## Integration Architecture

1. **JavaScript Client**: Wraps Go library via Node.js native bindings
2. **Appium Plugin**: Uses JavaScript client to provide Appium integration
3. **Go Core**: Provides the underlying AI/ML intelligence and data persistence

The system is designed for seamless integration into existing automation frameworks while providing intelligent element healing capabilities powered by AI.

## Dependencies Summary

- **Go**: OpenAI SDK, SQLite driver, XPath/HTML parsing libraries
- **JavaScript**: Node-addon-api, TypeScript, native bindings compilation
- **Appium Plugin**: Appium base plugin framework, SmartHealer JS client

## Platform Support

- **Core Go Library**: Cross-platform (Linux, macOS, Windows)
- **JavaScript Client**: Currently Linux x64 only
- **Appium Plugin**: Follows JavaScript client limitations
- **Node.js**: Version 16.0.0+
- **Appium**: Version 2.0+ (for plugin)