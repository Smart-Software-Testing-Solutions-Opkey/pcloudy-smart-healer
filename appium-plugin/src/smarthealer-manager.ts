import { SmartHealer, Platform, PageType, ComparisonMode } from 'smarthealer-js';
import { SmartHealerPluginConfig, ElementInfo, ElementContext } from './types';

export class SmartHealerManager {
  private static instance: SmartHealerManager;
  private initialized = false;

  private constructor() {}

  public static getInstance(): SmartHealerManager {
    if (!SmartHealerManager.instance) {
      SmartHealerManager.instance = new SmartHealerManager();
    }
    return SmartHealerManager.instance;
  }

  public async initialize(config: SmartHealerPluginConfig): Promise<void> {
    if (this.initialized) {
      return;
    }

    try {
      await SmartHealer.init({
        openai_key: config.openai_key,
        sqlite_db_path: config.sqlite_db_path || ''  // Pass empty string, Go will use default
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
      console.log('[SmartHealer] Calling resolveLocator with:', {
        project_id: elementInfo.project_id,
        context_id: elementInfo.context_id,
        xpath: elementInfo.xpath,
        platform: elementInfo.platform,
        page_type: elementInfo.page_type
      });

      const result = await SmartHealer.resolveLocator(elementInfo, {
        comparisionMode: ComparisonMode.Automatic
      });

      console.log('[SmartHealer] resolveLocator result:', {
        success: result.success,
        reason: result.reason,
        hasContent: !!result.content,
        content: result.content
      });

      if (result.success && result.content) {
        return result.content;
      }

      return null;
    } catch (error) {
      console.warn('[SmartHealer] sync resolution failed:', error);
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
    const pageSource = context.pageSource || '';
    let platform: Platform;
    let pageType: PageType;
    let contextId: string;
    let screenshot: string;

    // Determine platform, page type, context_id, and screenshot based on context
    if (context.isWebView) {
      // WebView: Platform=Web, PageType=HTML, contextId=URL, no screenshot
      platform = Platform.Web;
      pageType = PageType.HTML;
      contextId = context.currentUrl || '';
      screenshot = '';
    } else if (context.platformName?.toLowerCase() === 'android') {
      // Android Native: Platform=Android, PageType=XML, contextId=activity, no screenshot
      platform = Platform.Android;
      pageType = PageType.XML;
      contextId = context.currentActivity || '';
      screenshot = '';
    } else if (context.platformName?.toLowerCase() === 'ios') {
      // iOS Native: Platform=iOS, PageType=XML, contextId=empty, screenshot required
      platform = Platform.iOS;
      pageType = PageType.XML;
      contextId = '';
      screenshot = context.screenshot || '';
    } else {
      // Fallback: detect from page source
      platform = this.detectPlatform(context);
      pageType = this.detectPageType(pageSource);
      contextId = context.sessionId;
      screenshot = context.screenshot || '';
    }

    return {
      project_id: context.projectId,
      page_source: pageSource,
      b64_png: screenshot,
      xpath: this.convertLocatorToXPath(context.strategy),
      context_id: contextId,
      platform: platform,
      page_type: pageType
    };
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