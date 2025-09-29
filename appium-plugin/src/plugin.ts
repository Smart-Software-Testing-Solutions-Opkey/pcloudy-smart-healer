import { BasePlugin } from '@appium/base-plugin';
import { Driver, Element } from '@appium/types';
import { SmartHealerManager } from './smarthealer-manager';
import { SmartHealerConfig, ElementContext } from './types';

export class SmartHealerPlugin extends BasePlugin {
  private smartHealerManager: SmartHealerManager;
  private config: SmartHealerConfig;

  constructor(pluginName: string = 'smarthealer') {
    super(pluginName);
    this.smartHealerManager = SmartHealerManager.getInstance();
    this.config = {
      openai_key: '',
      sqlite_db_path: '',
      enabled: true
    };
  }

  public static newMethodMap = {
    '/session/:sessionId/smarthealer/config': {
      POST: {
        command: 'configureSmartHealer',
        payloadParams: { required: ['config'] }
      }
    }
  };

  async createSession(next: () => Promise<[string, any]>, driver: Driver, ...args: any[]): Promise<[string, any]> {
    const result = await next();

    // Initialize SmartHealer if config is available from capabilities
    const caps = driver.caps || {};
    const smartHealerConfig = caps['smarthealer:config'] as SmartHealerConfig;

    if (smartHealerConfig && smartHealerConfig.openai_key) {
      try {
        await this.initializeSmartHealer(smartHealerConfig);
        this.logger.info('SmartHealer initialized successfully');
      } catch (error) {
        this.logger.warn('Failed to initialize SmartHealer:', error);
      }
    }

    return result;
  }

  async deleteSession(next: () => Promise<void>, driver: Driver): Promise<void> {
    // Clean up SmartHealer resources
    try {
      this.smartHealerManager.close();
      this.logger.info('SmartHealer resources cleaned up');
    } catch (error) {
      this.logger.warn('Error cleaning up SmartHealer:', error);
    }

    return await next();
  }

  async findElement(next: () => Promise<Element>, driver: Driver, ...args: any[]): Promise<Element> {
    if (!this.config.enabled || !this.smartHealerManager.isInitialized()) {
      return await next();
    }

    const [strategy, selector] = args;

    try {
      // First, try the normal element finding
      const element = await next();

      // If successful, invoke async resolution for learning
      if (element) {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector }, element);
        await this.smartHealerManager.resolveLocatorAsync(context);
        this.logger.debug('SmartHealer async resolution triggered for successful find');
      }

      return element;
    } catch (originalError) {
      // Element not found, try SmartHealer sync resolution
      this.logger.info(`Element not found with ${strategy}:${selector}, attempting SmartHealer resolution`);

      try {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector });
        const healedLocator = await this.smartHealerManager.resolveLocatorSync(context);

        if (healedLocator) {
          this.logger.info(`SmartHealer suggested alternative locator: ${healedLocator}`);

          // Try to find element with the healed locator
          try {
            // Parse the healed locator and determine strategy
            const { strategy: healedStrategy, value: healedValue } = this.parseHealedLocator(healedLocator);
            const healedElement = await driver.executeCommand('findElement', healedStrategy, healedValue);

            if (healedElement) {
              this.logger.info('Successfully found element using SmartHealer suggestion');
              return healedElement;
            }
          } catch (healedError) {
            this.logger.warn('Failed to find element even with SmartHealer suggestion:', healedError);
          }
        }
      } catch (smartHealerError) {
        this.logger.warn('SmartHealer resolution failed:', smartHealerError);
      }

      // If all healing attempts fail, throw the original error
      throw originalError;
    }
  }

  async findElements(next: () => Promise<Element[]>, driver: Driver, ...args: any[]): Promise<Element[]> {
    if (!this.config.enabled || !this.smartHealerManager.isInitialized()) {
      return await next();
    }

    const [strategy, selector] = args;

    try {
      // First, try the normal element finding
      const elements = await next();

      // If successful and elements found, invoke async resolution for learning
      if (elements && elements.length > 0) {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector }, elements[0]);
        await this.smartHealerManager.resolveLocatorAsync(context);
        this.logger.debug('SmartHealer async resolution triggered for successful findElements');
      }

      return elements;
    } catch (originalError) {
      // Elements not found, try SmartHealer sync resolution
      this.logger.info(`Elements not found with ${strategy}:${selector}, attempting SmartHealer resolution`);

      try {
        const context = await this.buildElementContext(driver, { using: strategy, value: selector });
        const healedLocator = await this.smartHealerManager.resolveLocatorSync(context);

        if (healedLocator) {
          this.logger.info(`SmartHealer suggested alternative locator: ${healedLocator}`);

          // Try to find elements with the healed locator
          try {
            const { strategy: healedStrategy, value: healedValue } = this.parseHealedLocator(healedLocator);
            const healedElements = await driver.executeCommand('findElements', healedStrategy, healedValue);

            if (healedElements && healedElements.length > 0) {
              this.logger.info('Successfully found elements using SmartHealer suggestion');
              return healedElements;
            }
          } catch (healedError) {
            this.logger.warn('Failed to find elements even with SmartHealer suggestion:', healedError);
          }
        }
      } catch (smartHealerError) {
        this.logger.warn('SmartHealer resolution failed:', smartHealerError);
      }

      // If all healing attempts fail, throw the original error
      throw originalError;
    }
  }

  async configureSmartHealer(driver: Driver, config: SmartHealerConfig): Promise<{ success: boolean; message: string }> {
    try {
      await this.initializeSmartHealer(config);
      return { success: true, message: 'SmartHealer configured successfully' };
    } catch (error) {
      const message = `Failed to configure SmartHealer: ${error instanceof Error ? error.message : String(error)}`;
      this.logger.error(message);
      return { success: false, message };
    }
  }

  private async initializeSmartHealer(config: SmartHealerConfig): Promise<void> {
    this.config = { ...this.config, ...config };
    await this.smartHealerManager.initialize(this.config);
  }

  private async buildElementContext(
    driver: Driver,
    strategy: { using: string; value: string },
    element?: Element
  ): Promise<ElementContext> {
    let screenshot = '';
    let pageSource = '';

    try {
      // Get screenshot using executeCommand method
      screenshot = await driver.executeCommand('getScreenshot');
    } catch (error) {
      this.logger.warn('Failed to capture screenshot:', error);
    }

    try {
      // Get page source using executeCommand method
      pageSource = await driver.executeCommand('getPageSource');
    } catch (error) {
      this.logger.warn('Failed to get page source:', error);
    }

    return {
      sessionId: driver.sessionId || 'unknown',
      strategy,
      element,
      screenshot,
      pageSource
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