import { test, expect } from '@playwright/test';

test('global search palette works', async ({ page }) => {
  // 1. Go to homepage
  await page.goto('/');

  // 2. Open Command Palette (Cmd+K or click button)
  // We'll click the button since simulating Cmd+K can be flaky across OS
  await page.click('text=Search feature...');

  // 3. Verify it opens
  await expect(page.locator('input[placeholder="Type a command or search..."]')).toBeVisible();

  // 4. Search for "Playground"
  await page.fill('input[placeholder="Type a command or search..."]', 'Playground');

  // 5. Click the result
  await page.click('text=Playground');

  // 6. Verify navigation
  await expect(page).toHaveURL('/playground');
});
