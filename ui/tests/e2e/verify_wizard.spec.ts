import { test, expect } from '@playwright/test';

test('Verify Wizard UI', async ({ page }) => {
  await page.goto('http://localhost:9002/upstream-services');
  await page.waitForTimeout(500);

  // Click Add Service
  await page.getByRole('button', { name: 'Add Service' }).click();
  await page.waitForTimeout(500);

  // Take screenshot of template selector
  await page.screenshot({ path: '/tmp/wizard-1-templates.png' });

  // Click Time template
  await page.getByText('Check the current time in any timezone').click();
  await page.waitForTimeout(500);

  // Take screenshot of step 1
  await page.screenshot({ path: '/tmp/wizard-2-step1.png' });

  // Fill in name
  await page.getByLabel('Service Name').fill('my-time-service');
  await page.waitForTimeout(500);

  // Click connect
  await page.getByRole('button', { name: 'Connect' }).click();
  await page.waitForTimeout(1000);

  // Take screenshot of connection status
  await page.screenshot({ path: '/tmp/wizard-3-step2-3.png' });
});
