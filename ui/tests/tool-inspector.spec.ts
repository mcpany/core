
import { test, expect } from '@playwright/test';

test('Tools page loads and inspector opens', async ({ page }) => {
  await page.goto('/tools');

  // Check if tools are listed
  await expect(page.getByText('get_weather')).toBeVisible();

  // Open inspector for get_weather
  await page.getByRole('row', { name: 'get_weather' }).getByRole('button', { name: 'Inspect' }).click();

  // Check if inspector sheet is open
  await expect(page.getByRole('heading', { name: 'get_weather' })).toBeVisible();

  // Check if tabs exist
  await expect(page.getByRole('tab', { name: 'Test Tool' })).toBeVisible();
  await expect(page.getByRole('tab', { name: 'Schema Definition' })).toBeVisible();

  // Switch to Schema tab
  await page.getByRole('tab', { name: 'Schema Definition' }).click();
  await expect(page.getByText('"type": "object"')).toBeVisible();

  // Switch back to Test tab
  await page.getByRole('tab', { name: 'Test Tool' }).click();

  // Fill in the form
  await page.getByLabel('location').fill('New York');
  await page.getByRole('combobox').click();
  await page.getByRole('option', { name: 'celsius' }).click();

  // Execute tool
  await page.getByRole('button', { name: 'Execute Tool' }).click();

  // Check for result
  // Use a more specific locator or header for "Result"
  await expect(page.getByRole('heading', { name: 'Result' })).toBeVisible();
  await expect(page.getByText('"temperature": 22')).toBeVisible();
  await expect(page.getByText('"location": "New York"')).toBeVisible();
});
