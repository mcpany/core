import { test, expect } from '@playwright/test';

test('Stacks Management Flow', async ({ page }) => {
  // 1. Go to Stacks page
  await page.goto('/stacks');

  // 2. Create Stack
  await page.getByRole('button', { name: 'Create Stack' }).click();
  await page.getByLabel('Name').fill('e2e-stack');
  await page.getByRole('button', { name: 'Create', exact: true }).click();

  // 3. Verify redirection to editor
  await expect(page).toHaveURL(/\/stacks\/e2e-stack/);

  // 4. Edit YAML (Monaco is tricky to interact with, we might need to use page.evaluate or wait for it)
  // Wait for editor to be visible
  await page.waitForSelector('.monaco-editor');

  // Click Save
  await page.getByRole('button', { name: 'Save Changes' }).click();

  // Verify toast
  await expect(page.getByText('Stack Saved').first()).toBeVisible();

  // 5. Go back to list
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page).toHaveURL(/\/stacks$/);

  // 6. Verify stack in list
  await expect(page.getByText('e2e-stack').first()).toBeVisible();

  // 7. Cleanup (Delete stack)
  // Hover over the card to reveal delete button? Or just click it if accessible?
  // The delete button appears on hover.
  const stackCard = page.locator('.group').filter({ hasText: 'e2e-stack' });
  await stackCard.hover();

  // Handle confirm dialog
  page.on('dialog', dialog => dialog.accept());

  await stackCard.getByRole('button').click();

  // Verify deletion
  await expect(page.getByText('Stack Deleted').first()).toBeVisible();
  // Verify card is gone
  await expect(page.locator('.group').filter({ hasText: 'e2e-stack' })).not.toBeVisible();
});
