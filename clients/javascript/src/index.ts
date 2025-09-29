import * as path from 'path';
import {
  Config,
  Info,
  Options,
  Result,
  Platform,
  PageType,
  ComparisonMode,
  SmartHealerNative
} from './types';

// Load the native module
const native: SmartHealerNative = require(path.join(__dirname, '../build/Release/smarthealer.node'));

/**
 * SmartHealer Error class for better error handling
 */
export class SmartHealerError extends Error {
  public readonly code: string;
  public readonly details?: string | undefined;

  constructor(message: string, code: string = 'SMARTHEALER_ERROR', details?: string) {
    super(message);
    this.name = 'SmartHealerError';
    this.code = code;
    this.details = details;
    Error.captureStackTrace(this, SmartHealerError);
  }
}

/**
 * Main SmartHealer class providing a clean TypeScript API
 */
export class SmartHealer {
  private static _initialized: boolean = false;

  /**
   * Initialize SmartHealer with configuration
   */
  public static init(config: Config): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        if (!config.openai_key || !config.sqlite_db_path) {
          throw new SmartHealerError(
            'Missing required configuration: openai_key and sqlite_db_path are required',
            'INVALID_CONFIG'
          );
        }

        const result = native.initSmartHealer(config);

        if (result.success) {
          SmartHealer._initialized = true;
          resolve();
        } else {
          throw new SmartHealerError(
            `Failed to initialize SmartHealer: ${result.reason}`,
            'INIT_FAILED',
            result.content
          );
        }
      } catch (error) {
        if (error instanceof SmartHealerError) {
          reject(error);
        } else {
          reject(new SmartHealerError(
            `Initialization error: ${error instanceof Error ? error.message : String(error)}`,
            'INIT_ERROR'
          ));
        }
      }
    });
  }

  /**
   * Resolve element locator synchronously
   */
  public static resolveLocator(info: Info, options: Options): Promise<Result> {
    return new Promise((resolve, reject) => {
      try {
        if (!SmartHealer._initialized) {
          throw new SmartHealerError(
            'SmartHealer not initialized. Call SmartHealer.init() first.',
            'NOT_INITIALIZED'
          );
        }

        SmartHealer.validateInfo(info);
        SmartHealer.validateOptions(options);

        const result = native.resolveLocator(info, options);
        resolve(result);
      } catch (error) {
        if (error instanceof SmartHealerError) {
          reject(error);
        } else {
          reject(new SmartHealerError(
            `Locator resolution error: ${error instanceof Error ? error.message : String(error)}`,
            'RESOLVE_ERROR'
          ));
        }
      }
    });
  }

  /**
   * Resolve element locator asynchronously
   */
  public static resolveLocatorAsync(info: Info, options: Options): Promise<Result> {
    return new Promise((resolve, reject) => {
      try {
        if (!SmartHealer._initialized) {
          throw new SmartHealerError(
            'SmartHealer not initialized. Call SmartHealer.init() first.',
            'NOT_INITIALIZED'
          );
        }

        SmartHealer.validateInfo(info);
        SmartHealer.validateOptions(options);

        const result = native.resolveLocatorAsync(info, options);
        resolve(result);
      } catch (error) {
        if (error instanceof SmartHealerError) {
          reject(error);
        } else {
          reject(new SmartHealerError(
            `Async locator resolution error: ${error instanceof Error ? error.message : String(error)}`,
            'RESOLVE_ASYNC_ERROR'
          ));
        }
      }
    });
  }

  /**
   * Clean up SmartHealer resources
   */
  public static close(): void {
    try {
      native.close();
      SmartHealer._initialized = false;
    } catch (error) {
      throw new SmartHealerError(
        `Error during cleanup: ${error instanceof Error ? error.message : String(error)}`,
        'CLOSE_ERROR'
      );
    }
  }

  /**
   * Check if SmartHealer is initialized
   */
  public static get isInitialized(): boolean {
    return SmartHealer._initialized;
  }

  /**
   * Get available constants
   */
  public static get constants() {
    return native.constants;
  }

  // Validation helpers
  private static validateInfo(info: Info): void {
    const required = ['project_id', 'page_source', 'b64_png', 'xpath', 'context_id'];
    const missing = required.filter(field => !info[field as keyof Info]);

    if (missing.length > 0) {
      throw new SmartHealerError(
        `Missing required Info fields: ${missing.join(', ')}`,
        'INVALID_INFO'
      );
    }

    if (!Object.values(Platform).includes(info.platform)) {
      throw new SmartHealerError(
        `Invalid platform: ${info.platform}. Must be one of: ${Object.values(Platform).filter(v => typeof v === 'number').join(', ')}`,
        'INVALID_PLATFORM'
      );
    }

    if (!Object.values(PageType).includes(info.page_type)) {
      throw new SmartHealerError(
        `Invalid page_type: ${info.page_type}. Must be one of: ${Object.values(PageType).filter(v => typeof v === 'number').join(', ')}`,
        'INVALID_PAGE_TYPE'
      );
    }
  }

  private static validateOptions(options: Options): void {
    if (!Object.values(ComparisonMode).includes(options.comparisionMode)) {
      throw new SmartHealerError(
        `Invalid comparisionMode: ${options.comparisionMode}. Must be one of: ${Object.values(ComparisonMode).filter(v => typeof v === 'number').join(', ')}`,
        'INVALID_COMPARISON_MODE'
      );
    }
  }
}

// Export types and enums
export {
  Config,
  Info,
  Options,
  Result,
  Platform,
  PageType,
  ComparisonMode
};

// Export default
export default SmartHealer;