const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page, request }) => {
    // Use an existing valid mock tool endpoint, or use the pre-seeded "test-service"
    // that is available in the backend tests (config.minimal.yaml).
    // We avoid creating/deleting services as the test might time out waiting for discovery/restarts.
    // The pre-configured "test-service" from server/config.minimal.yaml should expose tools!

    // Instead, let's just create a super minimal, valid config and wait.
    const serviceConfig = {
      name: "weather-service-test",
      command_line_service: {
        command: "node",
        args: ["-e", "console.log('ready')"],
        tools: [
          {
             name: "get_weather",
             description: "Get weather for a location",
             schema: {
               type: "object",
               properties: {
                 location: { type: "string" }
               }
             }
          }
        ]
      }
    };

    const apiURL = \`/api/v1/services\`;
    await request.post(apiURL, {
      data: serviceConfig,
      headers: {
         'Content-Type': 'application/json',
      }
    });

    // Let backend settle
    await page.waitForTimeout(2000);

    await page.goto('/tools');

    await expect(page.getByText('get_weather')).toBeVisible();

    await page.locator('tr').filter({ hasText: 'get_weather' }).getByText('Inspect').click();

    await expect(page.getByText('get_weather').first()).toBeVisible();

    await page.getByRole('tab', { name: 'Schema' }).click();

    const schemaPanel = page.getByRole('tabpanel').filter({ hasText: 'Visual' });
    await expect(schemaPanel).toBeVisible();

    await schemaPanel.getByRole('tab', { name: 'JSON' }).click();

    // Check for the rendered elements by syntax highlighter
    await expect(page.locator('.language-json').filter({ hasText: /"location"/ }).first()).toBeVisible();
    await expect(page.locator('.language-json').filter({ hasText: /"type"/ }).first()).toBeVisible();

    await expect(page.locator('div[role="dialog"]').getByText('weather-service-test')).toBeVisible();

    // Clean up
    await request.delete('/api/v1/services/weather-service-test');
  });
});
`;

fs.writeFileSync(path, replacement);
