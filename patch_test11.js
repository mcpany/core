const fs = require('fs');
const path = 'ui/tests/tool-inspector.spec.ts';

const replacement = `
import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page, request }) => {
    // Seed the database with a real tool using the CommandLineUpstreamService schema
    // but correctly wrapped in the Top-level proto object
    const serviceConfig = {
      name: "weather-service-test",
      commandLineService: {
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
    const response = await request.post(apiURL, {
      data: serviceConfig,
      headers: {
         'Content-Type': 'application/json',
      }
    });

    if (!response.ok()) {
      console.error("Failed to seed database", await response.text());
    }

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
