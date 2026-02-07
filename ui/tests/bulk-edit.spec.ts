/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Bulk Edit services updates timeout for multiple services', async ({ page, request }) => {
  // 1. Seed Data
  // Create two services via API
  const service1 = {
    name: 'bulk-test-1',
    version: '1.0.0',
    disable: false,
    priority: 0,
    http_service: { address: 'http://example.com' },
    resilience: { timeout: '10s' }
  };
  const service2 = {
    name: 'bulk-test-2',
    version: '1.0.0',
    disable: false,
    priority: 0,
    http_service: { address: 'http://example.org' },
    resilience: { timeout: '20s' }
  };

  // Cleanup potential previous run
  await request.delete('/api/v1/services/bulk-test-1').catch(() => {});
  await request.delete('/api/v1/services/bulk-test-2').catch(() => {});

  const r1 = await request.post('/api/v1/services', { data: service1 });
  expect(r1.ok()).toBeTruthy();
  const r2 = await request.post('/api/v1/services', { data: service2 });
  expect(r2.ok()).toBeTruthy();

  // 2. Navigate to Services Page
  await page.goto('/upstream-services');

  // 3. Select both services
  const row1 = page.getByRole('row').filter({ hasText: 'bulk-test-1' });
  const row2 = page.getByRole('row').filter({ hasText: 'bulk-test-2' });

  // Wait for rows to appear
  await expect(row1).toBeVisible();
  await expect(row2).toBeVisible();

  // Check the checkboxes safely
  await row1.getByRole('checkbox').check();
  await row2.getByRole('checkbox').check();

  // 4. Open Bulk Edit Dialog
  // Ensure the button is visible and active (it depends on selection state)
  const bulkEditBtn = page.getByRole('button', { name: 'Bulk Edit' });
  await expect(bulkEditBtn).toBeVisible();
  await bulkEditBtn.click();

  const dialog = page.getByRole('dialog', { name: 'Bulk Edit Services' });
  await expect(dialog).toBeVisible();

  // 5. Switch to Configuration Tab
  await dialog.getByRole('tab', { name: 'Configuration' }).click();

  // 6. Set Timeout
  await dialog.getByLabel('Timeout').fill('60s');

  // 7. Apply Changes
  await dialog.getByRole('button', { name: 'Apply Changes' }).click();

  // 8. Verify dialog closes
  await expect(dialog).toBeHidden();

  // 9. Verify Backend State with retries (in case of async update lag)
  await expect(async () => {
      const verify1 = await request.get('/api/v1/services/bulk-test-1');
      const json1 = await verify1.json();
      const s1 = json1.service || json1;
      expect(s1.resilience.timeout).toBe('60s');
  }).toPass({ timeout: 5000 });

  await expect(async () => {
      const verify2 = await request.get('/api/v1/services/bulk-test-2');
      const json2 = await verify2.json();
      const s2 = json2.service || json2;
      expect(s2.resilience.timeout).toBe('60s');
  }).toPass({ timeout: 5000 });

  // Cleanup
  await request.delete('/api/v1/services/bulk-test-1');
  await request.delete('/api/v1/services/bulk-test-2');
});
