/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { indexClient } from '../src/lib/index-client';

test.describe('Lazy-MCP Tool Index', () => {
  test.beforeEach(async ({ page }) => {
    // Seed the index via API
    const tools = [
      {
        name: 'test-tool-1',
        description: 'A test tool for lazy loading',
        category: 'Testing',
        installed: false,
        sourceUrl: 'https://example.com/tool1',
        tags: ['test', 'lazy']
      },
      {
        name: 'test-tool-2',
        description: 'Another test tool',
        category: 'Testing',
        installed: true,
        sourceUrl: 'https://example.com/tool2',
        tags: ['test', 'installed']
      }
    ];

    // Direct fetch to seed, bypassing frontend client to avoid complexity in test environment
    const seedRes = await page.request.post('/api/v1/index/seed', {
      data: {
        tools: tools.map(t => ({
            name: t.name,
            description: t.description,
            category: t.category,
            installed: t.installed,
            source_url: t.sourceUrl,
            tags: t.tags
        })),
        clear: true
      }
    });
    expect(seedRes.ok()).toBeTruthy();
  });

  test('should display indexed tools and filter by search', async ({ page }) => {
    await page.goto('/index');

    // Check header
    await expect(page.getByRole('heading', { name: 'Tool Index' })).toBeVisible();

    // Check stats
    await expect(page.getByText('Total Indexed').first()).toBeVisible();
    // Stats might take a moment to update/fetch
    await expect(page.locator('.text-2xl').first()).toHaveText('2'); // Total Indexed

    // Check table content
    await expect(page.getByText('test-tool-1')).toBeVisible();
    await expect(page.getByText('test-tool-2')).toBeVisible();

    // Check Status Badges
    await expect(page.getByText('Lazy')).toBeVisible();
    await expect(page.getByText('Installed')).toBeVisible();

    // Test Search
    const searchInput = page.getByPlaceholder('Search tools');
    await searchInput.fill('lazy');

    // Wait for debounce
    await page.waitForTimeout(500);

    // Should show tool-1 (lazy) and hide tool-2 (installed, no 'lazy' keyword)
    // Wait, tool-2 tags are 'test', 'installed'. 'lazy' is not in tool-2.
    // Tool-1 tags: 'test', 'lazy'.
    await expect(page.getByText('test-tool-1')).toBeVisible();
    await expect(page.getByText('test-tool-2')).not.toBeVisible();

    // Clear search
    await searchInput.fill('');
    await page.waitForTimeout(500);
    await expect(page.getByText('test-tool-2')).toBeVisible();
  });

  test('should show install button for lazy tools', async ({ page }) => {
    await page.goto('/index');

    const row = page.getByRole('row', { name: 'test-tool-1' });
    const installBtn = row.getByRole('button', { name: 'Install' });

    await expect(installBtn).toBeVisible();
    await installBtn.click();

    // Check toast
    await expect(page.getByText('Installation Started')).toBeVisible();
  });
});
