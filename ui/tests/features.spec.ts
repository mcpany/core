
import { test, expect } from '@playwright/test';

test('tools page loads', async ({ page }) => {
  await page.goto('/tools');
  await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();
});

test('resources page loads', async ({ page }) => {
  await page.goto('/resources');
  await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();
});

test('prompts page loads', async ({ page }) => {
  await page.goto('/prompts');
  await expect(page.getByRole('heading', { name: 'Prompts' })).toBeVisible();
});
