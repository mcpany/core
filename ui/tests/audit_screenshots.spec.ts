
import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const date = new Date().toISOString().split('T')[0];
const auditDir = path.resolve(__dirname, `../../.audit/ui/${date}`);

if (!fs.existsSync(auditDir)) {
  fs.mkdirSync(auditDir, { recursive: true });
}

test('Audit Screenshots', async ({ page }) => {
  // 1. Dashboard
  await page.goto('/');
  await page.waitForTimeout(1000); // Wait for animations/load
  await page.screenshot({ path: path.join(auditDir, 'dashboard.png'), fullPage: true });

  // 2. Services
  await page.goto('/services');
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });

  // Click edit on first service
  // await page.locator('button').filter({ hasText: 'Edit Service' }).first().click().catch(() => {});
  // Or click the settings icon
  // await page.locator('.lucide-settings').first().click();
  // await page.waitForTimeout(500);
  // await page.screenshot({ path: path.join(auditDir, 'service_edit.png') });
  // Close sheet
  // await page.keyboard.press('Escape');

  // 3. Tools
  await page.goto('/tools');
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });

  // 4. Resources
  await page.goto('/resources');
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'resources.png'), fullPage: true });

  // 5. Prompts
  await page.goto('/prompts');
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'prompts.png'), fullPage: true });

  // 6. Settings - Profiles
  await page.goto('/settings');
  await page.getByRole('tab', { name: 'Profiles' }).click();
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'settings_profiles.png'), fullPage: true });

  // 7. Settings - Webhooks
  await page.getByRole('tab', { name: 'Webhooks' }).click();
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'settings_webhooks.png'), fullPage: true });

  // 8. Settings - Middleware
  await page.getByRole('tab', { name: 'Middleware' }).click();
  await page.waitForTimeout(500);
  await page.screenshot({ path: path.join(auditDir, 'settings_middleware.png'), fullPage: true });
});
