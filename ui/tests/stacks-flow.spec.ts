import { test, expect } from '@playwright/test';
import { cleanupCollection } from './e2e/test-data';

test.describe('Stacks Flow', () => {
  const stackName = `e2e-stack-${Date.now()}`;

  test.afterEach(async ({ request }) => {
    await cleanupCollection(stackName, request);
  });

  test('should create and deploy a stack', async ({ page }) => {
    // 1. Navigate to Stacks
    await page.goto('/stacks');

    // 2. Click Create
    await page.getByRole('button', { name: 'New Stack' }).click();

    // 3. Fill name
    await page.getByPlaceholder('my-stack').fill(stackName);
    await page.getByRole('button', { name: 'Create', exact: true }).click();

    // 4. Verify redirection
    await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));

    // 5. Verify Editor loaded
    await expect(page.getByText('config.yaml')).toBeVisible();

    // 6. Deploy (assuming backend mock/real is running)
    // We expect a success toast or at least the button to be clickable.
    await page.getByRole('button', { name: 'Deploy' }).click();

    // 7. Verify Toast (Sonner toast usually has role status or specific class)
    // Wait for text "Stack deployed successfully!"
    await expect(page.getByText('Stack deployed successfully!')).toBeVisible({ timeout: 10000 });
  });
});
