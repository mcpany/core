import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  // We assume the backend is running and seeded with 'production-stack'

  test('should list existing stacks', async ({ page }) => {
    // Navigate to Stacks page
    await page.goto('/stacks');

    // Verify the page title
    await expect(page.getByRole('heading', { name: 'Stacks', exact: true })).toBeVisible();

    // Verify the seeded stack card exists
    await expect(page.getByText('production-stack')).toBeVisible();
    await expect(page.getByText('2 services')).toBeVisible();
  });

  test('should create a new stack', async ({ page }) => {
    await page.goto('/stacks');

    // Click Create Stack
    await page.getByRole('button', { name: 'Create Stack' }).click();

    // Fill in name
    const stackName = 'e2e-test-stack-' + Date.now();
    await page.getByLabel('Name').fill(stackName);

    // Click Create
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // Verify redirect to editor
    // URL should contain encoded stack name
    await expect(page).toHaveURL(new RegExp(`/stacks/${encodeURIComponent(stackName)}`));

    // Verify editor loaded
    await expect(page.getByText('config.yaml')).toBeVisible();

    // Go back to list and verify it appears
    await page.goto('/stacks');
    await expect(page.getByText(stackName)).toBeVisible();
  });
});
