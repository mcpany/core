import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const date = new Date().toISOString().split('T')[0];
const auditDir = path.join(__dirname, '../../.audit/ui', date);

test.beforeAll(async () => {
  if (!fs.existsSync(auditDir)) {
    fs.mkdirSync(auditDir, { recursive: true });
  }
});

test('Dashboard loads and displays metrics', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('h2')).toContainText('Dashboard');
  await expect(page.getByText('Total Requests')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'dashboard.png'), fullPage: true });
});

test('Services page lists services and allows toggling', async ({ page }) => {
  await page.goto('/services');
  await expect(page.locator('h2')).toContainText('Services');
  await expect(page.getByText('weather-service')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });
});

test('Tools page lists tools', async ({ page }) => {
  await page.goto('/tools');
  await expect(page.locator('h2')).toContainText('Tools');
  await expect(page.getByText('get_weather')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });
});

test('Resources page lists resources', async ({ page }) => {
  await page.goto('/resources');
  await expect(page.locator('h2')).toContainText('Resources');
  await expect(page.getByText('My Notes')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'resources.png'), fullPage: true });
});

test('Prompts page lists prompts', async ({ page }) => {
  await page.goto('/prompts');
  await expect(page.locator('h2')).toContainText('Prompts');
  await expect(page.getByText('summarize_text')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'prompts.png'), fullPage: true });
});

test('Middleware page displays pipeline', async ({ page }) => {
  await page.goto('/middleware');
  await expect(page.locator('h2')).toContainText('Middleware Pipeline');
  await expect(page.getByText('Authentication')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'middleware.png'), fullPage: true });
});

test('Webhooks page allows adding webhook', async ({ page }) => {
  await page.goto('/webhooks');
  await expect(page.locator('h2')).toContainText('Webhooks');

  // Fill input
  await page.getByPlaceholder('https://api.example.com/webhook').fill('https://test.com/hook');
  await page.getByRole('button', { name: 'Add Webhook' }).click();

  await expect(page.getByText('https://test.com/hook')).toBeVisible();
  await page.screenshot({ path: path.join(auditDir, 'webhooks.png'), fullPage: true });
});
