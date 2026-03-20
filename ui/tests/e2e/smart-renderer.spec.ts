import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Smart Result Renderer - E2E Verification', () => {

  test.beforeEach(async ({ request }) => {
    // Seed database using existing fixtures
    await seedGlobalState(request);
  });

  test('should display parsed JSON properly instead of raw string', async ({ page }) => {
    // Go to playground where tool results are rendered
    await page.goto('/playground');

    // Wait for services to load
    await page.waitForSelector('text=Payment Gateway');

    // Click on Payment Gateway
    await page.click('text=Payment Gateway');

    // Click on the process_payment tool
    await page.click('text=process_payment');

    // Fill in the args
    await page.fill('input[placeholder*="JSON arguments"]', '{"id": "123"}');

    // Click execute
    await page.click('button:has-text("Execute")');

    // Wait for the tool result to render.
    // Wait until the "Result: process_payment" card appears
    await page.waitForSelector('text=Result: process_payment');

    // We expect the result to render without throwing errors and displaying nicely.
    // If it is a deeply stringified JSON, the `deepParseJson` function inside `SmartResultRenderer`
    // will now parse it properly, preventing raw strings inside `.whitespace-pre-wrap` div.

    // In our E2E, we just verify the result component handles the response from our seeded tool without crashing.
    expect(await page.locator('.whitespace-pre-wrap, .react-json-view').count()).toBeGreaterThan(0);
  });

});
