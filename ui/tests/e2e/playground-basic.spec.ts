/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './test-data';

test.describe('Playground Basic Verification', () => {
  test.beforeEach(async ({ page, request }) => {
      await seedServices(request);
      await seedUser(request, "e2e-playground-admin");
      await page.goto('/login');
      await page.fill('input[name="username"]', "e2e-playground-admin");
      await page.fill('input[name="password"]', 'password');
      await Promise.all([
          page.waitForURL('/', { timeout: 30000 }),
          page.click('button[type="submit"]', { force: true })
      ]);
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await cleanupUser(request, "e2e-playground-admin");
  });

  test('should execute calculator tool and verify output', async ({ page }) => {
    // Navigate to playground
    await page.goto('/playground');

    // Waiting for chat input
    const chatInput = page.getByPlaceholder('Enter command or select a tool...');
    await expect(chatInput).toBeVisible({ timeout: 10000 });

    // Type a command
    const msg = 'echo_tool {"test": "echo"}';
    await chatInput.fill(msg);

    // Click Send
    const sendBtn = page.getByRole('button', { name: 'Send' });
    await expect(sendBtn).toBeVisible();
    await sendBtn.click();

    // Assert: Check if message appears
    await expect(page.getByText(msg)).toBeVisible({ timeout: 10000 });

    // Checking layout (Library visible)
    await expect(page.getByText('Library')).toBeVisible();

    // Verify result (Execution happened). echo_tool echoes back the input as JSON string.
    await expect(page.getByText('echoed_output')).toBeVisible({ timeout: 10000 });
  });
});
