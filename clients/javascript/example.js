const { SmartHealer, Platform, PageType, ComparisonMode } = require('./dist/index.js');

async function testSmartHealer() {
  try {
    console.log('Testing SmartHealer TypeScript wrapper...');

    // Test initialization
    await SmartHealer.init({
      openai_key: 'test-key',
      sqlite_db_path: './test.db'
    });

    console.log('SmartHealer initialized:', SmartHealer.isInitialized);
    console.log('Available constants:', SmartHealer.constants);

    // Test locator resolution (this will fail without valid data, but we can test the API)
    try {
      const result = await SmartHealer.resolveLocator({
        project_id: 'test-project',
        page_source: '<html><body><button>Test</button></body></html>',
        b64_png: 'test-image-base64',
        xpath: '//button',
        context_id: 'test-context',
        platform: Platform.Web,
        page_type: PageType.HTML
      }, {
        comparisionMode: ComparisonMode.Automatic
      });

      console.log('Resolution result:', result);
    } catch (error) {
      console.log('Expected locator resolution error:', error.message);
    }

    // Clean up
    SmartHealer.close();
    console.log('SmartHealer closed. Initialized:', SmartHealer.isInitialized);

  } catch (error) {
    console.error('Test error:', error);
  }
}

testSmartHealer();