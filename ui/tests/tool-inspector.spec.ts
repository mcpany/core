import { test, expect } from '@playwright/test';

test.describe('Tool Inspector', () => {
  test('Tools page loads and inspector opens with real data', async ({ page, request }) => {

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

    const apiURL = `/api/v1/services`;
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

    await page.waitForSelector('table tbody tr', { state: 'visible', timeout: 15000 });

    const expandAccordion = await page.locator('.lucide-chevron-down').first();
    if (await expandAccordion.isVisible()) {
       await expandAccordion.click();
    }

    const inspectBtn = page.locator('button:has-text("Inspect")').first();
    await expect(inspectBtn).toBeVisible({ timeout: 10000 });

    await inspectBtn.evaluate(b => b.click());

    await page.waitForTimeout(1000);

    const visualText = page.locator('*:has-text("Visual")').last();
    await expect(visualText).toBeVisible({ timeout: 10000 });
  });
});
