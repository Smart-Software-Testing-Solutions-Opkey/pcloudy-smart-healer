# SmartHealer Appium Plugin

An Appium plugin that integrates SmartHealer's intelligent element locator resolution and healing capabilities into your Appium test automation workflows.

## Features

- **Automatic Element Healing**: When element location fails, SmartHealer attempts to find alternative locators using AI-powered analysis
- **Learning Mode**: Successful element finds trigger async resolution to improve SmartHealer's knowledge base
- **Multi-platform Support**: Works with Web, Android, and iOS applications
- **Configurable**: Enable/disable SmartHealer functionality as needed
- **Runtime Configuration**: Update SmartHealer settings during test execution

## Installation

```bash
# Install the plugin locally
appium plugin install --source=local /path/to/smarthealer-appium-plugin

# Or install from npm (when published)
appium plugin install smarthealer-appium-plugin
```

## Configuration

### Via Capabilities

Configure SmartHealer when creating your Appium session:

```javascript
const capabilities = {
  'platformName': 'Web',
  'browserName': 'chrome',
  // SmartHealer configuration
  'smarthealer:config': {
    openai_key: 'your-openai-api-key',
    sqlite_db_path: './smarthealer.db', // Optional, defaults to ~/.smarthealer/appium-plugin.db
    enabled: true // Optional, defaults to true
  }
};
```

### Via API

Update SmartHealer configuration during test execution:

```javascript
await driver.execute('smarthealer:configureSmartHealer', {
  config: {
    openai_key: 'updated-api-key',
    enabled: false
  }
});
```

## Usage

The plugin automatically intercepts `findElement` and `findElements` commands:

1. **Successful Element Find**: Triggers async resolution to learn and improve locator knowledge
2. **Failed Element Find**: Attempts sync resolution to find alternative locators

```javascript
// Normal element finding - SmartHealer works transparently
const element = await driver.findElement('xpath', '//button[@id="submit"]');

// If the above fails, SmartHealer will:
// 1. Analyze the page source and screenshot
// 2. Use AI to suggest alternative locators
// 3. Attempt to find the element using the suggested locator
// 4. Return the element if found, or throw the original error if not
```

## How It Works

### Learning Mode (Successful Finds)
```
Element Found → Async SmartHealer Resolution → Knowledge Base Update
```

### Healing Mode (Failed Finds)
```
Element Not Found → Sync SmartHealer Resolution → Try Alternative Locator → Return Element or Original Error
```

## Requirements

- Node.js 16+
- Appium 2.0+
- OpenAI API key
- SmartHealer Go library (included in dependencies)

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

## Example

See `example/usage-example.js` for a complete working example.

## License

ISC