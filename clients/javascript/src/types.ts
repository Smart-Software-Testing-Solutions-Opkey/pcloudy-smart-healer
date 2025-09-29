// SmartHealer TypeScript Type Definitions

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

export interface SmartHealerConstants {
  Platform: {
    Android: number;
    Ios: number;
    Web: number;
  };
  PageType: {
    XML: number;
    HTML: number;
  };
  ComparisionMode: {
    Automatic: number;
    Manual: number;
    Screenshot: number;
  };
}

export interface SmartHealerNative {
  initSmartHealer(config: Config): Result;
  resolveLocator(info: Info, options: Options): Result;
  resolveLocatorAsync(info: Info, options: Options): Result;
  close(): void;
  constants: SmartHealerConstants;
}