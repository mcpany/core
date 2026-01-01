import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';

const AUDIT_DIR = path.join(process.cwd(), '.audit/ui', new Date().toISOString().split('T')[0]);

// Ensure audit directory exists
if (!fs.existsSync(AUDIT_DIR)) {
  fs.mkdirSync(AUDIT_DIR, { recursive: true });
}

test.describe('MCP Any UI E2E', () => {
  test('Dashboard loads and displays charts', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/MCPAny Manager/);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    await expect(page.getByText('Total Requests')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard.png') });
  });

  test('Services page loads and lists services', async ({ page }) => {
    await page.goto('/services');
    await expect(page.getByText('Manage your upstream MCP services')).toBeVisible();
    await expect(page.getByText('github-service')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services.png') });
  });

  test('Tools page loads', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.getByText('Explore available MCP tools')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png') });
  });

  test('Resources page loads', async ({ page }) => {
    await page.goto('/resources');
    await expect(page.getByText('Manage and view resources')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'resources.png') });
  });

  test('Prompts page loads', async ({ page }) => {
    await page.goto('/prompts');
    await expect(page.getByText('Manage and view prompts')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'prompts.png') });
  });

  test('Profiles page loads', async ({ page }) => {
    await page.goto('/profiles');
    await expect(page.getByText('Manage execution profiles')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'profiles.png') });
  });

  test('Middleware page loads', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.getByText('Middleware Pipeline')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png') });
  });

  test('Webhooks page loads', async ({ page }) => {
    await page.goto('/webhooks');
    await expect(page.getByText('Manage and test webhook integrations')).toBeVisible();

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks.png') });
  });
});
