# SmartHealer Appium Plugin

An Appium plugin that integrates SmartHealer's intelligent element locator resolution and healing capabilities into your Appium test automation workflows.

## Features

- **Automatic Element Healing**: When element location fails, SmartHealer attempts to find alternative locators using AI-powered analysis
- **Learning Mode**: Successful element finds are saved to improve SmartHealer's knowledge base
- **Multi-platform Support**: Works with Web, Android, and iOS applications
- **Persistent Background Workers**: Background workers run continuously across sessions to process queues
- **Platform-Specific Context Handling**: Automatically uses the right context (activity, URL, or screenshot) based on platform
- **Project Isolation**: Each project maintains its own knowledge base while sharing the same SmartHealer instance

## Installation

```bash
# Install the plugin locally
appium plugin install --source=local /path/to/smarthealer-appium-plugin

# Or install from npm (when published)
appium plugin install smarthealer-appium-plugin
```

## Configuration

SmartHealer is configured at **two levels**:

### 1. Plugin-Level Configuration (Required)

Configure when starting Appium server. Choose one of the following methods:

#### Option 1: Config File (Recommended)

Create `.appiumrc.json` in your project root:

```json
{
  "server": {
    "use-plugins": ["smarthealer"],
    "plugin": {
      "smarthealer": {
        "openai_key": "sk-your-key",
        "sqlite_db_path": "/path/to/smarthealer.db",
        "enabled": true
      }
    }
  }
}
```

Or use `.appiumrc.js` for dynamic configuration:

```javascript
module.exports = {
  server: {
    'use-plugins': ['smarthealer'],
    plugin: {
      smarthealer: {
        openai_key: process.env.SMARTHEALER_OPENAI_KEY || 'sk-fallback-key',
        sqlite_db_path: process.env.SMARTHEALER_DB_PATH,
        enabled: true
      }
    }
  }
};
```

Then start Appium:

```bash
# Automatically loads .appiumrc.json or .appiumrc.js from current directory
appium

# Or specify custom config file
appium --config /path/to/appiumrc.json
```

#### Option 2: Command Line

```bash
# Via command line args
appium --use-plugins=smarthealer \
  --plugin-args='{"smarthealer":{"openai_key":"sk-your-key","sqlite_db_path":"/path/to/db.db"}}'

# Or via environment variable
export SMARTHEALER_OPENAI_KEY="sk-your-openai-key"
appium --use-plugins=smarthealer
```

**Plugin Args:**
- `openai_key` (Required): Your OpenAI API key for AI-powered healing
- `sqlite_db_path` (Optional): Custom database path. Defaults to `~/.smarthealer/smarthealer.db`
- `enabled` (Optional): Enable/disable plugin. Defaults to `true`

**Environment Variables:**
- `SMARTHEALER_OPENAI_KEY`: Alternative to passing `openai_key` in plugin args

### 2. Session-Level Configuration (Required)

Provide `project_id` in capabilities for each test session:

```javascript
const capabilities = {
  'platformName': 'Android', // or 'iOS', 'Web'
  'deviceName': 'Pixel 6',
  'app': '/path/to/app.apk',
  // SmartHealer session config (REQUIRED)
  'smarthealer:config': {
    project_id: 'my-test-project'  // REQUIRED - Unique identifier for your project
  }
};
```

**Why Two Levels?**
- **Plugin-level**: SmartHealer initializes **once** when Appium starts. Background workers start immediately and run continuously across all sessions, processing queues even between test runs.
- **Session-level**: Each test session specifies its `project_id` to maintain separate knowledge bases while sharing the same SmartHealer instance and background workers.

**Key Benefits:**
- ✅ Background workers **never stop** - they process queues continuously
- ✅ No repeated initialization overhead per session
- ✅ OpenAI key and database path configured once for all projects
- ✅ Multiple projects can run simultaneously with proper isolation

## Usage

Once configured, SmartHealer works **automatically and transparently** - no code changes needed!

### How It Works

The plugin intercepts all `findElement` and `findElements` commands:

#### When Element is Found (Learning Mode)
```
Find Element → Success → Save to Database → Queue Description Generation
                                            ↓
                        Background workers process queue continuously
```

#### When Element is Not Found (Healing Mode)
```
Find Element → Fails → Look for Similar Pages in Database
                     ↓
              Found Similar Page? → Try Alternative Locators
                     ↓                        ↓
                    No                     Success → Return Element
                     ↓                        ↓
              Try AI Generation            Failure → Return Original Error
```

### Code Example

```javascript
// Your test code remains unchanged
const element = await driver.findElement('xpath', '//button[@id="submit"]');

// SmartHealer works behind the scenes:
// 1. First attempt with original locator
// 2. On failure: search similar pages in database
// 3. Try alternative locators if found
// 4. On success: save page/locator for future reference
// 5. Queue description generation (processed by background workers)
```

## Platform-Specific Context Handling

SmartHealer automatically detects the platform and uses the appropriate context for page matching:

| Platform | Context Used | Screenshot | Page Matching Strategy |
|----------|-------------|------------|------------------------|
| **Android Native** | Current Activity | ❌ Not needed | Activity-based (fast) |
| **iOS Native** | None | ✅ Required | Screenshot-based (visual) |
| **WebView/Browser** | Current URL | ❌ Not needed | URL-based (precise) |

### How Context Matching Works

#### Android Native
```
1. Capture current activity: "com.myapp.LoginActivity"
2. Search database for pages with same activity + project_id + locator
3. If found: try alternative locators from that page
4. If not found: this is a new page, save it
```

#### iOS Native
```
1. Capture screenshot of current screen
2. Search database for pages with same project_id + locator
3. Compare screenshots using AI to find visually similar pages
4. If found: try alternative locators
5. If not found: save screenshot with page
```

#### WebView/Browser
```
1. Capture current URL: "https://myapp.com/login"
2. Search database for pages with same URL + project_id + locator
3. If found: try alternative locators
4. If not found: save page with URL context
```

## Background Workers

SmartHealer runs **two background workers** continuously:

### Description Queue Worker
- Generates AI descriptions for newly discovered elements
- Runs at ~2 requests/second (rate-limited for OpenAI)
- Processes queue even when no sessions are active

### Healing Queue Worker
- Processes async healing requests (if using async mode)
- Generates new locators when all stored locators fail
- Also rate-limited to respect OpenAI limits

**Lifecycle:**
```
Appium Start → Workers Start → Keep Running → Appium Stop → Workers Stop
     ↓              ↓              ↓                ↓              ↓
  Plugin Init   Process Queue  Process Queue   Finish Job   Graceful Exit
                 (Session 1)   (Between Sessions)
```

**Benefits:**
- ✅ No queue backlog - descriptions generated continuously
- ✅ Better OpenAI API utilization - workers run during idle time
- ✅ Faster test execution - descriptions ready for next test run

## Requirements

- Node.js 16+
- Appium 2.0+
- OpenAI API key (required)
- SmartHealer Go library (included in dependencies)
- SQLite 3 (for database storage)

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

## Examples

### Complete Setup Example

**Step 1: Start Appium with SmartHealer**
```bash
appium --use-plugins=smarthealer \
  --plugin-args='{"smarthealer":{"openai_key":"sk-your-key"}}'
```

**Step 2: Configure Test Capabilities**

#### Android Native App
```javascript
const capabilities = {
  platformName: 'Android',
  deviceName: 'Pixel 6',
  app: '/path/to/app.apk',
  'smarthealer:config': {
    project_id: 'my-android-app'
  }
};

// SmartHealer will use:
// - context_id: Current activity (e.g., "com.myapp.MainActivity")
// - platform: Android
// - page_type: XML
```

#### iOS Native App
```javascript
const capabilities = {
  platformName: 'iOS',
  deviceName: 'iPhone 14',
  app: '/path/to/app.app',
  'smarthealer:config': {
    project_id: 'my-ios-app'
  }
};

// SmartHealer will use:
// - context_id: Empty (screenshot-based matching)
// - platform: iOS
// - page_type: XML
// - Screenshot: Captured automatically
```

#### WebView/Mobile Browser
```javascript
const capabilities = {
  platformName: 'Android',
  deviceName: 'Pixel 6',
  browserName: 'Chrome',
  'smarthealer:config': {
    project_id: 'my-web-app'
  }
};

// SmartHealer will use:
// - context_id: Current URL (e.g., "https://example.com/login")
// - platform: Web
// - page_type: HTML
```

See `example/usage-example.js` for a complete working example.

## License

ISC