import { test, expect } from '@playwright/test';

test('Prompts CRUD', async ({ page }) => {
  // 1. Navigate to Prompts
  await page.goto('/prompts');
  await expect(page.locator('h3', { hasText: 'Prompt Library' })).toBeVisible();

  // 2. Create Prompt
  await page.getByTitle('New Prompt').click();
  await expect(page.locator('h2', { hasText: 'New Prompt' })).toBeVisible();

  const promptName = `test-prompt-${Date.now()}`;
  await page.fill('#name', promptName);
  await page.fill('#description', 'E2E Test Prompt');

  // Default templates are pre-filled in PromptEditor
  await page.click('button:has-text("Save Prompt")');

  // 3. Verify in list and selected
  await expect(page.getByRole('button', { name: promptName })).toBeVisible();
  await expect(page.locator('h2', { hasText: promptName })).toBeVisible();

  // 4. Run Prompt
  // Default arg is "name"
  await page.fill('#name', 'Universe');
  await page.click('button:has-text("Generate Messages")');

  // Verify output
  await expect(page.locator('.whitespace-pre-wrap')).toContainText('Hello Universe!');

  // 5. Update Prompt
  await page.click('button:has-text("Edit Definition")');
  await page.fill('#description', 'Updated Description');
  await page.click('button:has-text("Save Prompt")');
  await expect(page.getByText('Updated Description')).toBeVisible();

  // 6. Delete Prompt
  await page.click('button:has-text("Edit Definition")');

  page.once('dialog', dialog => dialog.accept());
  await page.click('button:has-text("Delete")');

  // Verify deleted
  await expect(page.getByRole('button', { name: promptName })).not.toBeVisible();
});
