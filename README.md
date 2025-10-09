# SmartHealer

SmartHealer is an AI-powered element locator resolution and healing system for web automation frameworks. It intelligently maintains and heals broken element locators in your automation tests, reducing maintenance overhead and improving test reliability.

## What is SmartHealer?

When UI elements change in your application, automation tests often break because locators (XPath, CSS selectors) no longer work. SmartHealer automatically:

- **Learns** from successful element finds to build a knowledge base
- **Heals** broken locators by finding alternative paths to the same element
- **Adapts** to UI changes using AI-powered analysis
- **Maintains** a history of page structures and element locations

## Key Features

- **Automatic Locator Healing**: When a locator fails, SmartHealer uses AI to generate working alternatives
- **Learning Mode**: Builds knowledge from successful element interactions
- **Multi-Platform Support**: Works with web, mobile web, and native mobile applications
- **Framework Agnostic**: Can be integrated into various automation frameworks
- **Visual Context**: Uses screenshots and page structure for intelligent element matching
- **Zero Configuration Fallback**: Works out of the box with sensible defaults

## Architecture

SmartHealer consists of three main components:

1. **Core Go Library** (`smarthealer/`): High-performance healing engine with SQLite-based knowledge store
2. **JavaScript Client** (`clients/javascript/`): Node.js bindings for direct integration
3. **Appium Plugin** (`appium-plugin/`): Ready-to-use plugin for Appium 2.0+ automation

## Getting Started

### Prerequisites

- **Go**: Version 1.25.1+ (for building the core library)
- **Node.js**: Version 16.0.0+ (for JavaScript client and Appium plugin)
- **Appium**: Version 2.0+ (for plugin usage)
- **OpenAI API Key**: Required for AI-powered healing

### Quick Start with Appium Plugin

The fastest way to use SmartHealer is through the Appium plugin:

1. Install the plugin (see `appium-plugin/` for instructions)
2. Configure with your OpenAI API key
3. SmartHealer automatically intercepts element finds and heals broken locators

### Direct Integration with JavaScript Client

For custom integrations, use the JavaScript client directly (see `clients/javascript/` for instructions).

### Building from Source

To build the Go core library (see `smarthealer/` for detailed instructions).

## Configuration

SmartHealer requires minimal configuration:

- **OpenAI API Key**: For AI-powered locator generation
- **Database Path**: Optional SQLite database location (defaults to `~/.smarthealer/smarthealer.db`)

Configuration can be provided through:
- Appium capabilities for plugin usage
- Direct API calls for custom integrations
- Environment variables (planned)

## How It Works

### Learning Phase
When an element is successfully found:
1. SmartHealer captures the page structure and screenshot
2. Stores the locator and contextual information
3. Generates semantic descriptions of elements using AI
4. Builds a searchable knowledge base

### Healing Phase
When a locator fails:
1. SmartHealer retrieves similar pages from the knowledge base
2. Tries known working locators for similar elements
3. If all fail, generates new locators using AI analysis
4. Returns the healed locator to the test
5. Updates knowledge base with the new locator

## Platform Support

- **Operating Systems**: Linux x64 (macOS and Windows support planned)
- **Automation Frameworks**: Appium (Selenium, Playwright support planned)
- **Application Types**: Web, Mobile Web, Native Mobile

## Documentation

Detailed documentation for each component:

- **Core Library**: See `smarthealer/README.md` (coming soon)
- **JavaScript Client**: See `clients/javascript/README.md`
- **Appium Plugin**: See `appium-plugin/README.md` (coming soon)
- **Development Guide**: See `CLAUDE.md` for contributor information

## Examples

Example projects and test suites demonstrating SmartHealer usage are coming soon.

## Contributing

We appreciate your interest in SmartHealer! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for information about how you can participate in this project.

## License

SmartHealer is released under the [MIT License](LICENSE).

## Roadmap

Upcoming features and improvements:

- Multi-platform support (macOS, Windows)
- Additional framework integrations (Selenium, Playwright)
- Enhanced AI models for better healing accuracy
- Cloud-based knowledge sharing (optional)
- Performance optimizations
- Comprehensive test coverage

## Support

- **Bug Reports**: [GitHub Issues](../../issues)
- **Feature Requests**: [GitHub Issues](../../issues)
- **Questions**: [GitHub Issues](../../issues) with "question" label

## Acknowledgments

SmartHealer leverages:
- OpenAI GPT models for intelligent element analysis
- SQLite for efficient local data storage
- Appium ecosystem for test automation integration

---

**Note**: SmartHealer is under active development. APIs and features may change before the 1.0 release.
