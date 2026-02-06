/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('layout smoke test', async ({ page, request }) => {
  // Ensure cleanup of test stack
  await request.delete('/api/v1/collections/smoke-stack').catch(() => {});

  await page.goto('/');

  // Check for Sidebar
  const sidebar = page.locator('text=MCP Any').first();
  await expect(sidebar).toBeVisible();

  // Check for Sidebar links
  await expect(page.locator('a[href="/stacks"]').first()).toBeVisible();
  await expect(page.locator('a[href="/upstream-services"]').first()).toBeVisible();
  await expect(page.locator('a[href="/settings"]').first()).toBeVisible();

  // Navigate to Stacks
  await page.locator('a[href="/stacks"]').first().click();
  await page.waitForURL('**/stacks');
  await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible({ timeout: 10000 });

  // Create a stack for the smoke test
  // Use .first() to avoid ambiguity if empty state also has a button
  await page.getByRole('button', { name: 'Create Stack' }).first().click();

  await expect(page.getByRole('dialog')).toBeVisible();
  await page.getByLabel('Stack Name').fill('smoke-stack');
  await page.getByRole('button', { name: 'Create' }).click();

  // Wait for dialog to close (implies success or at least action taken)
  // Increased timeout for CI stability
  await expect(page.getByRole('dialog')).toBeHidden({ timeout: 15000 });

  // Navigate to Stack Detail (should happen automatically)
  // Increase timeout to handle potentially slow server/network
  await expect(page).toHaveURL(/\/stacks\/smoke-stack/, { timeout: 20000 });

  await expect(page.locator('h2')).toContainText('smoke-stack');
  await expect(page.locator('h2')).toContainText('Stack');

  // Check Tabs
  await expect(page.locator('button[role="tab"]', { hasText: 'Overview & Status' })).toBeVisible();
  await expect(page.locator('button[role="tab"]', { hasText: 'Editor' })).toBeVisible();

  // Cleanup
  await request.delete('/api/v1/collections/smoke-stack').catch(() => {});
});
