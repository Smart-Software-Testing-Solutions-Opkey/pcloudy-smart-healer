import { SmartHealer, Platform, PageType, ComparisonMode } from 'smarthealer-js';
import { SmartHealerConfig, ElementInfo, ElementContext } from './types';

export class SmartHealerManager {
  private static instance: SmartHealerManager;
  private initialized = false;
  private config?: SmartHealerConfig;

  private constructor() {}

  public static getInstance(): SmartHealerManager {
    if (!SmartHealerManager.instance) {
      SmartHealerManager.instance = new SmartHealerManager();
    }
    return SmartHealerManager.instance;
  }

  public async initialize(config: SmartHealerConfig): Promise<void> {
    if (this.initialized) {
      return;
    }

    this.config = config;

    try {
      await SmartHealer.init({
        openai_key: config.openai_key,
        sqlite_db_path: config.sqlite_db_path || this.getDefaultDbPath()
      });

      this.initialized = true;
    } catch (error) {
      throw new Error(`Failed to initialize SmartHealer: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  public async resolveLocatorSync(context: ElementContext): Promise<string | null> {
    if (!this.initialized) {
      throw new Error('SmartHealer not initialized');
    }

    try {
      const elementInfo = await this.buildElementInfo(context);
      const result = await SmartHealer.resolveLocator(elementInfo, {
        comparisionMode: ComparisonMode.Automatic
      });

      if (result.success && result.content) {
        return result.content;
      }

      return null;
    } catch (error) {
      console.warn('SmartHealer sync resolution failed:', error);
      return null;
    }
  }

  public async resolveLocatorAsync(context: ElementContext): Promise<void> {
    if (!this.initialized) {
      throw new Error('SmartHealer not initialized');
    }

    try {
      const elementInfo = await this.buildElementInfo(context);
      await SmartHealer.resolveLocatorAsync(elementInfo, {
        comparisionMode: ComparisonMode.Automatic
      });
    } catch (error) {
      console.warn('SmartHealer async resolution failed:', error);
    }
  }

  public isInitialized(): boolean {
    return this.initialized;
  }

  public close(): void {
    if (this.initialized) {
      SmartHealer.close();
      this.initialized = false;
    }
  }

  private async buildElementInfo(context: ElementContext): Promise<ElementInfo> {
    const screenshot = context.screenshot || '';
    const pageSource = context.pageSource || '';

    return {
      project_id: this.generateProjectId(context.sessionId),
      page_source: pageSource,
      b64_png: screenshot,
      xpath: this.convertLocatorToXPath(context.strategy),
      context_id: context.sessionId,
      platform: this.detectPlatform(context),
      page_type: this.detectPageType(pageSource)
    };
  }

  private generateProjectId(sessionId: string): string {
    return `appium-${sessionId}`;
  }

  private convertLocatorToXPath(strategy: { using: string; value: string }): string {
    switch (strategy.using) {
      case 'xpath':
        return strategy.value;
      case 'id':
        return `//*[@id='${strategy.value}']`;
      case 'className':
        return `//*[@class='${strategy.value}']`;
      case 'name':
        return `//*[@name='${strategy.value}']`;
      case 'tagName':
        return `//${strategy.value}`;
      case 'cssSelector':
        return strategy.value;
      case 'linkText':
        return `//a[text()='${strategy.value}']`;
      case 'partialLinkText':
        return `//a[contains(text(), '${strategy.value}')]`;
      default:
        return strategy.value;
    }
  }

  private detectPlatform(context: ElementContext): Platform {
    const sessionId = context.sessionId.toLowerCase();
    if (sessionId.includes('android')) {
      return Platform.Android;
    }
    if (sessionId.includes('ios') || sessionId.includes('safari')) {
      return Platform.iOS;
    }
    return Platform.Web;
  }

  private detectPageType(pageSource: string): PageType {
    if (pageSource.includes('<?xml') || pageSource.includes('<hierarchy')) {
      return PageType.XML;
    }
    return PageType.HTML;
  }

  private getDefaultDbPath(): string {
    const os = require('os');
    const path = require('path');
    return path.join(os.homedir(), '.smarthealer', 'appium-plugin.db');
  }
}