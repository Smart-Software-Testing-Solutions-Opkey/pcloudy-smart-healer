import { Element } from '@appium/types';

export interface SmartHealerPluginConfig {
  openai_key: string;
  sqlite_db_path?: string;
  enabled?: boolean;
}

export interface SmartHealerSessionConfig {
  project_id: string;  // Required from capabilities per session
}

export interface ElementInfo {
  project_id: string;
  page_source: string;
  b64_png: string;
  xpath: string;
  context_id: string;
  platform: Platform;
  page_type: PageType;
}

export interface LocatorStrategy {
  using: string;
  value: string;
}

export interface ElementContext {
  sessionId: string;
  strategy: LocatorStrategy;
  element?: Element;
  screenshot?: string;
  pageSource?: string;
  currentActivity?: string;  // For Android native
  currentUrl?: string;        // For WebView
  isWebView?: boolean;        // To distinguish webview from native
  platformName?: string;      // 'Android', 'iOS', 'Web'
  projectId: string;          // From capabilities
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