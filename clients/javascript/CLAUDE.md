# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the JavaScript/Node.js client wrapper for SmartHealer, which provides intelligent web element locator resolution and healing capabilities. The project bridges a Go-based SmartHealer library with Node.js through native C++ bindings.

## Architecture

**Core Components:**
- `src/smarthealer.cc`: Node.js N-API C++ wrapper that binds to the Go static library
- `includes/libsmarthealer.h`: CGO-generated C header defining the API interface
- `lib/linux-x64/libsmarthealer.a`: Static library containing the Go SmartHealer implementation
- `binding.gyp`: Node-gyp configuration for building the native addon

**Key APIs Exposed:**
- `initSmartHealer(config)`: Initialize the SmartHealer with OpenAI key and SQLite database path
- `resolveLocator(info, options)`: Synchronously resolve element locators
- `resolveLocatorAsync(info, options)`: Asynchronously resolve element locators
- `close()`: Clean up SmartHealer resources

**Data Structures:**
- `Info`: Contains project ID, page source, base64 PNG, XPath, context ID, platform, and page type
- `Options`: Specifies comparison mode (Automatic, Manual, Screenshot)
- `Config`: Holds OpenAI API key and SQLite database path
- `Result`: Returns success status, reason, and content

## Development Commands

**Build the native addon:**
```bash
npm install
```
This automatically runs `node-gyp rebuild` to compile the C++ addon and link against the static library.

**Manual rebuild:**
```bash
node-gyp rebuild
```

## Platform Support

Currently supports Linux x64 only. The static library is platform-specific and located in `lib/linux-x64/libsmarthealer.a`.

## Integration Notes

- The module exports CommonJS format (`type: "commonjs"`)
- Uses Node Addon API v8.5.0+ for N-API bindings
- Memory management includes automatic cleanup via `freeResult()` calls
- The CGO bridge handles Go string marshalling and C struct conversion

## Systematic File Naming
Format: `YYYY-MM-DD-[001-999]-[category]-[four-word-summary].md`
Folder: `docs/work/`
Categories: `bug` | `feature` | `task` | `research` | `learnings`

Examples:
- `2025-07-18-001-feature-user-authentication-system.md`
- `2025-07-18-002-bug-database-connection-timeout.md`

## Communication Style
- **Concise**: No fluff, direct responses
- **Evidence-based**: Show, don't just tell
- **Contextual**: Reference past learnings from `docs/work/`

## Planning Protocol
1. **Context Gathering**: Check `docs/work/` for relevant past decisions
2. **Assumption Documentation**: Explicit assumptions in plan files
3. **Execution Gate**: Only proceed after planning is complete


When context drops below 30%: 

1. Document every decision made
2. List what failed (with code snippets)  
3. Note what worked brilliantly
4. Write handoff notes for next session

Use the `Systematic File Naming` given above.