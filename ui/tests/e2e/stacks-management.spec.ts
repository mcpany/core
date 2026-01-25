import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {

  test.beforeEach(async ({ page }) => {
    // Mock Settings API
    await page.route('**/api/v1/settings', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ configured: true, version: '0.1.0' })
      });
    });
  });

  test('should list existing stacks', async ({ page }) => {
    // Mock initial list
    await page.route('**/api/v1/collections', async (route) => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify([
                    { name: 'existing-stack', services: [] }
                ])
            });
        }
    });

    await page.goto('/stacks');
    await expect(page.getByText('existing-stack')).toBeVisible();
  });

  test('should create a new stack', async ({ page }) => {
    // State for this test
    let stacks = [{ name: 'existing-stack', services: [] }];

    await page.route('**/api/v1/collections', async (route) => {
         if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify(stacks)
            });
         } else if (route.request().method() === 'POST') {
             const data = route.request().postDataJSON();
             stacks.push({ name: data.name, services: [] });
             await route.fulfill({
                 status: 200,
                 contentType: 'application/json',
                 body: JSON.stringify({ name: data.name, services: [] })
             });
         }
    });

    await page.goto('/stacks');

    await page.getByRole('button', { name: 'Create Stack' }).first().click();
    await page.getByLabel('Stack Name').fill('new-stack');

    // Click Create in the dialog
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    await expect(page.getByText('new-stack')).toBeVisible();
  });

  test('should delete a stack', async ({ page }) => {
    // State for this test
    let stacks = [{ name: 'existing-stack', services: [] }];

    await page.route('**/api/v1/collections', async (route) => {
         if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify(stacks)
            });
         }
    });

    await page.route('**/api/v1/collections/existing-stack', async (route) => {
        if (route.request().method() === 'DELETE') {
             stacks = []; // Remove from list
             await route.fulfill({ status: 200, body: '{}' });
        }
    });

    await page.goto('/stacks');

    // Trigger dropdown
    const card = page.locator('.group').filter({ hasText: 'existing-stack' });
    await card.hover();

    // Force click to bypass opacity transition issues
    await card.getByRole('button', { name: 'Open menu' }).click({ force: true });

    await page.getByText('Delete Stack').click();

    // Handle Alert Dialog
    await expect(page.getByText('Are you sure?')).toBeVisible();
    await page.getByRole('button', { name: 'Delete', exact: true }).click();

    // Verify the card is gone.
    // We filter for a card (group) that contains the stack name.
    await expect(page.locator('.group').filter({ hasText: 'existing-stack' })).not.toBeVisible();
    await expect(page.getByText('No stacks found')).toBeVisible();
  });

});
