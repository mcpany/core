import { test, expect, BrowserContext } from '@playwright/test';
import { seedUser, seedGlobalState } from './test-data';

test.describe('Bulk Actions', () => {
  let context: BrowserContext;

  test.beforeEach(async ({ browser }) => {
    context = await browser.newContext();

    // Seed the database
    const apiContext = context.request;
    await seedGlobalState(apiContext);
    await seedUser(apiContext, 'admin'); // Provide second argument

    // Authenticate
    await context.addCookies([{
      name: 'auth_token',
      value: 'test-token',
      domain: 'localhost',
      path: '/'
    }]);
  });

  test.afterEach(async () => {
    if (context) await context.close();
  });

  test('should perform bulk delete', async () => {
    const page = await context.newPage();
    await page.goto('/services');

    // Wait for the table to load
    await page.waitForSelector('table');

    // Select the first two services
    const checkboxes = await page.locator('table tbody tr td:first-child button[role="checkbox"]').all();
    expect(checkboxes.length).toBeGreaterThanOrEqual(2);

    // Get their names to verify later
    const name1 = await page.locator('table tbody tr:nth-child(1) td:nth-child(3)').innerText();
    const name2 = await page.locator('table tbody tr:nth-child(2) td:nth-child(3)').innerText();

    await checkboxes[0].click();
    await checkboxes[1].click();

    // Verify the bulk actions bar appears
    const bulkBar = page.locator('text=2 selected');
    await expect(bulkBar).toBeVisible();

    // Click bulk delete
    await page.getByRole('button', { name: 'Delete' }).first().click();

    // Confirm deletion
    const dialog = page.locator('[role="dialog"]');
    await expect(dialog).toBeVisible();
    await dialog.getByRole('button', { name: 'Delete' }).click();

    // Verify the bar disappears and services are removed
    await expect(bulkBar).not.toBeVisible();
    await expect(page.locator(`text=${name1}`).first()).not.toBeVisible();
    await expect(page.locator(`text=${name2}`).first()).not.toBeVisible();
  });
});
