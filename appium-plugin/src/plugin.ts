import { BasePlugin } from '@appium/base-plugin';
import { Driver, Element } from '@appium/types';
import { SmartHealerManager } from './smarthealer-manager';
import { SmartHealerPluginConfig, SmartHealerSessionConfig, ElementContext } from './types';

export class SmartHealerPlugin extends BasePlugin {
  private smartHealerManager: SmartHealerManager;
  private pluginConfig: SmartHealerPluginConfig;
  private currentSessionProjectId?: string;

  constructor(pluginName: string = 'smarthealer', cliArgs?: any) {
    super(pluginName);
    this.smartHealerManager = SmartHealerManager.getInstance();

    // Get config from plugin args (set when Appium starts)
    this.pluginConfig = {
      openai_key: cliArgs?.openai_key || process.env.SMARTHEALER_OPENAI_KEY || '',
      sqlite_db_path: cliArgs?.sqlite_db_path || '',
      enabled: cliArgs?.enabled !== false
    };

    // Initialize SmartHealer immediately when plugin loads (before any sessions)
    if (this.pluginConfig.enabled && this.pluginConfig.openai_key) {
      this.initializeSmartHealer().catch(error => {
        this.log.error('\x1b[31m✗ Failed to initialize SmartHealer on plugin load:\x1b[0m', error);
      });
    } else if (this.pluginConfig.enabled && !this.pluginConfig.openai_key) {
      this.log.warn('\x1b[33m⚠ SmartHealer enabled but no OpenAI key provided\x1b[0m');
      this.log.warn('Set via --plugins-args or SMARTHEALER_OPENAI_KEY environment variable');
    }
  }

  private async initializeSmartHealer(): Promise<void> {
    try {
      await this.smartHealerManager.initialize(this.pluginConfig);
      this.log.info('\x1b[32m✓ SmartHealer initialized - background workers are running\x1b[0m');
    } catch (error) {
      throw error;
    }
  }

  // No custom API methods needed - configuration is done at plugin level

  async createSession(next: () => Promise<[string, any]>, driver: Driver, ...args: any[]): Promise<[string, any]> {
    const result = await next();

    // In Appium 3, capabilities are in the result object's value property
    const resultObj = result as any;
    const caps = (resultObj.value && Array.isArray(resultObj.value)) ? resultObj.value[1] : {};

    // Get project_id from capabilities for this session
    const sessionConfig = caps['smarthealer:config'] as SmartHealerSessionConfig;

    if (sessionConfig?.project_id) {
      this.currentSessionProjectId = sessionConfig.project_id;
      this.log.info(`\x1b[32m✓ SmartHealer session started for project: ${this.currentSessionProjectId}\x1b[0m`);
    } else {
      this.log.warn('\x1b[33m⚠ No project_id found in smarthealer:config capability\x1b[0m');
      this.log.warn('SmartHealer will use "default" as project_id for this session');
      this.currentSessionProjectId = 'default';
    }

    return result;
  }

  async deleteSession(next: () => Promise<void>, driver: Driver): Promise<void> {
    // Just clear the session project_id, but SmartHealer stays alive
    this.log.info(`\x1b[36mℹ SmartHealer session ended for project: ${this.currentSessionProjectId}\x1b[0m`);
    this.log.info('\x1b[36mℹ Background workers continue running...\x1b[0m');
    this.currentSessionProjectId = undefined;

    return await next();
  }

  async findElement(next: () => Promise<Element>, driver: Driver, ...args: any[]): Promise<Element> {
    if (!this.pluginConfig.enabled || !this.smartHealerManager.isInitialized()) {
      return await next();
    }

    const [strategy, selector] = args;

    try {
      // First, try the normal element finding
      const element = await next();

      // If successful, invoke sync resolution (always call resolveLocator)
      if (element) {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector }, element);
        await this.smartHealerManager.resolveLocatorSync(context);
        this.log.info('SmartHealer sync resolution triggered for successful find');
      }

      return element;
    } catch (originalError) {
      // Element not found, try SmartHealer sync resolution
      this.log.info(`Element not found with ${strategy}:${selector}, attempting SmartHealer resolution`);

      try {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector });
        const healedLocator = await this.smartHealerManager.resolveLocatorSync(context);

        if (healedLocator) {
          this.log.info(`SmartHealer suggested alternative locator: ${healedLocator}`);

          // Try to find element with the healed locator
          try {
            // Parse the healed locator and determine strategy
            const { strategy: healedStrategy, value: healedValue } = this.parseHealedLocator(healedLocator);
            const healedElement = await driver.executeCommand('findElement', healedStrategy, healedValue);

            if (healedElement) {
              this.log.info('Successfully found element using SmartHealer suggestion');
              return healedElement;
            }
          } catch (healedError) {
            this.log.warn('Failed to find element even with SmartHealer suggestion:', healedError);
          }
        } else {
          this.log.info('SmartHealer did not find an alternative locator (element may be new or no similar pages in database)');
        }
      } catch (smartHealerError) {
        this.log.warn('SmartHealer resolution failed:', smartHealerError);
      }

      // If all healing attempts fail, throw the original error
      throw originalError;
    }
  }



  private async buildElementContext(
    driver: Driver,
    strategy: { using: string; value: string },
    element?: Element
  ): Promise<ElementContext> {
    let screenshot = '';
    let pageSource = '';
    let currentActivity: string | undefined;
    let currentUrl: string | undefined;
    let isWebView = false;
    let platformName: string | undefined;

    // Get platform name
    try {
      const caps = await driver.getSession();
      platformName = caps.platformName || (caps as any).platform;
    } catch (error) {
      this.log.warn('Failed to get platform name:', error);
    }

    // Get page source (always needed)
    try {
      pageSource = await driver.executeCommand('getPageSource');
    } catch (error) {
      this.log.warn('Failed to get page source:', error);
    }

    // Detect if it's a webview
    try {
      const currentContext = await driver.executeCommand('getCurrentContext');
      isWebView = currentContext && currentContext.toString().toLowerCase().includes('webview');
    } catch (error) {
      // If getCurrentContext fails, assume native
      isWebView = false;
    }

    if (isWebView) {
      // WebView: Get current URL, no screenshot needed
      try {
        currentUrl = await driver.executeCommand('getUrl');
      } catch (error) {
        this.log.warn('Failed to get current URL for webview:', error);
      }
    } else if (platformName?.toLowerCase() === 'android') {
      // Android Native: Get current activity, no screenshot needed
      try {
        currentActivity = await driver.executeCommand('getCurrentActivity');
      } catch (error) {
        this.log.warn('Failed to get current activity:', error);
      }
    } else if (platformName?.toLowerCase() === 'ios') {
      // iOS Native: Get screenshot, no context_id
      try {
        screenshot = await driver.executeCommand('getScreenshot');
      } catch (error) {
        this.log.warn('Failed to capture screenshot for iOS:', error);
      }
    }

    return {
      sessionId: driver.sessionId || 'unknown',
      strategy,
      element,
      screenshot,
      pageSource,
      currentActivity,
      currentUrl,
      isWebView,
      platformName,
      projectId: this.currentSessionProjectId || 'default'
    };
  }

  private parseHealedLocator(healedLocator: string): { strategy: string; value: string } {
    // Try to determine the strategy from the healed locator
    if (healedLocator.startsWith('//') || healedLocator.startsWith('/')) {
      return { strategy: 'xpath', value: healedLocator };
    }

    if (healedLocator.startsWith('#')) {
      return { strategy: 'css selector', value: healedLocator };
    }

    if (healedLocator.startsWith('.')) {
      return { strategy: 'css selector', value: healedLocator };
    }

    if (healedLocator.includes('[') && healedLocator.includes(']')) {
      return { strategy: 'css selector', value: healedLocator };
    }

    // Default to xpath if we can't determine
    return { strategy: 'xpath', value: healedLocator };
  }
}

export default SmartHealerPlugin;