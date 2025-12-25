
import { test, expect } from '@playwright/test';
import path from 'path';

const AUDIT_DIR = path.join(__dirname, '../.audit/ui/2025-05-15');

test.describe('Audit Screenshots', () => {

  test('Capture Dashboard', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard.png'), fullPage: true });
  });

  test('Capture Services', async ({ page }) => {
    await page.goto('/services');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services.png'), fullPage: true });
  });

  test('Capture Tools', async ({ page }) => {
    await page.goto('/tools');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
  });

  test('Capture Middleware', async ({ page }) => {
    await page.goto('/middleware');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

    test('Capture Webhooks', async ({ page }) => {
    await page.goto('/webhooks');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks.png'), fullPage: true });
  });

    test('Capture Settings', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(AUDIT_DIR, 'settings.png'), fullPage: true });
  });
});
