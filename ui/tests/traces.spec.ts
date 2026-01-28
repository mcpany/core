/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

const mockTrace = {
  id: 'trace-123',
  timestamp: new Date().toISOString(),
  totalDuration: 150,
  status: 'success',
  trigger: 'user',
  rootSpan: {
    id: 'span-1',
    name: 'test-tool',
    type: 'tool',
    startTime: 1000,
    endTime: 1150,
    status: 'success',
    input: { query: 'test' },
    output: { result: 'ok' },
    children: [
      {
        id: 'span-2',
        name: 'test-service-call',
        type: 'service',
        serviceName: 'backend-api',
        startTime: 1020,
        endTime: 1100,
        status: 'success'
      }
    ]
  }
};

test.describe('Trace Detail', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the traces API
    await page.route('**/api/traces', async (route) => {
      console.log('Mocking /api/traces');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        json: [mockTrace]
      });
    });
  });

  test('should display sequence diagram tab', async ({ page }) => {
    await page.goto('/traces');

    // Wait for the trace item to appear in the list
    const traceItem = page.getByText('test-tool').first();
    await expect(traceItem).toBeVisible({ timeout: 10000 });

    // Click it to ensure it is selected
    await traceItem.click();

    // Verify Trace Detail view is showing the selected trace
    // The title should be "test-tool"
    await expect(page.locator('h2', { hasText: 'test-tool' })).toBeVisible();

    // Check if "Overview" tab is visible (default)
    await expect(page.getByRole('tab', { name: 'Overview' })).toBeVisible();

    // Check if "Sequence" tab is visible
    const sequenceTab = page.getByRole('tab', { name: 'Sequence' });
    await expect(sequenceTab).toBeVisible();

    // Click it
    await sequenceTab.click();

    // Check if diagram is rendered (check for actors inside the tab panel)
    const tabContent = page.locator('[role="tabpanel"][data-state="active"]');
    await expect(tabContent.getByText('backend-api')).toBeVisible();
    await expect(tabContent.getByText('Client')).toBeVisible();
    await expect(tabContent.getByText('MCP Any')).toBeVisible();
  });
});
