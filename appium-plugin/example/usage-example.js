const { remote } = require('webdriverio');

async function testSmartHealerPlugin() {
  console.log('Testing SmartHealer Appium Plugin...');

  const driver = await remote({
    hostname: 'localhost',
    port: 4723,
    path: '/',
    capabilities: {
      'platformName': 'Web',
      'browserName': 'chrome',
      'appium:chromeOptions': {
        'args': ['--headless']
      },
      // SmartHealer plugin configuration
      'smarthealer:config': {
        openai_key: process.env.OPENAI_API_KEY || 'your-openai-api-key',
        sqlite_db_path: './smarthealer-test.db',
        enabled: true
      }
    },
    plugins: {
      smarthealer: {}
    }
  });

  try {
    // Navigate to a test page
    await driver.url('https://example.com');

    console.log('Testing element finding with SmartHealer...');

    // Test 1: Find an element that exists (should trigger async resolution)
    try {
      const existingElement = await driver.findElement('xpath', '//h1');
      console.log('✓ Found existing element:', existingElement);
    } catch (error) {
      console.log('✗ Failed to find existing element:', error.message);
    }

    // Test 2: Try to find an element that doesn't exist (should trigger sync resolution)
    try {
      const nonExistentElement = await driver.findElement('xpath', '//button[@id="non-existent"]');
      console.log('✓ SmartHealer found alternative for non-existent element:', nonExistentElement);
    } catch (error) {
      console.log('✗ SmartHealer could not heal non-existent element (expected):', error.message);
    }

    // Test 3: Configure SmartHealer via API
    try {
      await driver.execute('smarthealer:configureSmartHealer', {
        config: {
          openai_key: process.env.OPENAI_API_KEY || 'updated-key',
          enabled: true
        }
      });
      console.log('✓ SmartHealer configuration updated via API');
    } catch (error) {
      console.log('✗ Failed to update SmartHealer config:', error.message);
    }

  } catch (error) {
    console.error('Test error:', error);
  } finally {
    await driver.deleteSession();
    console.log('Test completed');
  }
}

// Run the test if this file is executed directly
if (require.main === module) {
  testSmartHealerPlugin().catch(console.error);
}

module.exports = { testSmartHealerPlugin };