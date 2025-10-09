# SmartHealer JavaScript/TypeScript Client

A TypeScript/JavaScript client for SmartHealer - intelligent web element locator resolution and healing library.

## Features

- 🔧 **TypeScript Support**: Full type definitions for better development experience
- 🛡️ **Error Handling**: Comprehensive error handling with custom error types
- 🚀 **Easy to Use**: Simple async/await API
- 🔍 **Smart Validation**: Input validation with helpful error messages
- 📦 **Zero Dependencies**: No runtime dependencies (except native bindings)

## Installation

```bash
npm install smarthealer-js
```

## Quick Start

```typescript
import { SmartHealer, Platform, PageType, ComparisonMode } from 'smarthealer-js';

async function example() {
  try {
    // Initialize SmartHealer
    await SmartHealer.init({
      openai_key: 'your-openai-api-key',
      sqlite_db_path: './smarthealer.db'
    });

    // Resolve element locator
    const result = await SmartHealer.resolveLocator({
      project_id: 'my-project',
      page_source: '<html>...</html>',
      b64_png: 'base64-encoded-screenshot',
      xpath: '//button[@id="submit"]',
      context_id: 'unique-context-id',
      platform: Platform.Web,
      page_type: PageType.HTML
    }, {
      comparisionMode: ComparisonMode.Automatic
    });

    if (result.success) {
      console.log('New locator:', result.content);
    } else {
      console.log('Resolution failed:', result.reason);
    }

    // Clean up when done
    SmartHealer.close();
  } catch (error) {
    console.error('Error:', error.message);
  }
}
```

## API Reference

### `SmartHealer.init(config: Config): Promise<void>`

Initialize SmartHealer with configuration.

**Parameters:**
- `config.openai_key`: OpenAI API key for AI-powered healing
- `config.sqlite_db_path`: Path to SQLite database file

### `SmartHealer.resolveLocator(info: Info, options: Options): Promise<Result>`

Resolve element locator synchronously.

**Parameters:**
- `info.project_id`: Unique project identifier
- `info.page_source`: HTML/XML source code of the page
- `info.b64_png`: Base64-encoded screenshot of the page
- `info.xpath`: Current XPath locator to heal
- `info.context_id`: Unique context identifier
- `info.platform`: Target platform (`Platform.Android`, `Platform.iOS`, `Platform.Web`)
- `info.page_type`: Page type (`PageType.XML`, `PageType.HTML`)

**Options:**
- `options.comparisionMode`: Comparison mode (`ComparisonMode.Automatic`, `ComparisonMode.Manual`, `ComparisonMode.Screenshot`)

### `SmartHealer.resolveLocatorAsync(info: Info, options: Options): Promise<Result>`

Resolve element locator asynchronously (currently same as sync version).

### `SmartHealer.close(): void`

Clean up SmartHealer resources and close connections.

### Properties

- `SmartHealer.isInitialized`: Check if SmartHealer is initialized
- `SmartHealer.constants`: Access to platform/mode constants

## Types

```typescript
export interface Config {
  openai_key: string;
  sqlite_db_path: string;
}

export interface Info {
  project_id: string;
  page_source: string;
  b64_png: string;
  xpath: string;
  context_id: string;
  platform: Platform;
  page_type: PageType;
}

export interface Options {
  comparisionMode: ComparisonMode;
}

export interface Result {
  success: boolean;
  reason: string;
  content: string;
}

export enum Platform {
  Android = 0,
  iOS = 1,
  Web = 2
}

export enum PageType {
  XML = 0,
  HTML = 1
}

export enum ComparisonMode {
  Automatic = 0,
  Manual = 1,
  Screenshot = 2
}
```

## Error Handling

The library provides a custom `SmartHealerError` class with specific error codes:

```typescript
import { SmartHealerError } from 'smarthealer-js';

try {
  await SmartHealer.init(config);
} catch (error) {
  if (error instanceof SmartHealerError) {
    console.log('Error code:', error.code);
    console.log('Details:', error.details);
  }
}
```

**Error Codes:**
- `INVALID_CONFIG`: Missing or invalid configuration
- `INIT_FAILED`: Initialization failed
- `NOT_INITIALIZED`: SmartHealer not initialized
- `INVALID_INFO`: Invalid or missing Info fields
- `INVALID_PLATFORM`: Invalid platform value
- `INVALID_PAGE_TYPE`: Invalid page type value
- `INVALID_COMPARISON_MODE`: Invalid comparison mode value
- `RESOLVE_ERROR`: Locator resolution error
- `RESOLVE_ASYNC_ERROR`: Async resolution error
- `CLOSE_ERROR`: Error during cleanup

## Platform Support

Currently supports:
- **OS**: Linux x64 only
- **Node.js**: Version 16.0.0 or higher

## Development

```bash
# Install dependencies
npm install

# Build TypeScript
npm run build

# Watch mode for development
npm run dev

# Clean build artifacts
npm run clean
```

## License

MIT