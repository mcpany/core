/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Marketplace Smart Configuration', () => {
  test.use({ bypassCSP: true });

  test('should inject schema for GitHub server', async ({ page }) => {
    // Mock the Awesome List fetch
    await page.route('**/awesome-mcp-servers/**/README.md', async route => {
      const markdown = `
# Awesome MCP Servers

## 📂 Dev Tools
* [GitHub](https://github.com/modelcontextprotocol/server-github) 🐙 - GitHub integration
* [Slack](https://github.com/modelcontextprotocol/server-slack) 💬 - Slack integration
      `;
      await route.fulfill({ status: 200, contentType: 'text/plain', body: markdown });
    });

    await page.goto('/marketplace');

    // Verify page loaded
    await expect(page.getByRole('heading', { name: 'Marketplace' })).toBeVisible({ timeout: 10000 });

    // Click Community Tab
    await page.getByRole('tab', { name: 'Community' }).click();

    // Wait for the list to render
    await expect(page.getByText('Discover 2 servers from the community')).toBeVisible({ timeout: 10000 });

    // Search for GitHub
    await page.getByPlaceholder('Search community servers...').fill('GitHub');

    // Find the card that has "GitHub" text and an "Install" button
    // This is more robust than exact heading match
    const githubCard = page.locator('div').filter({ hasText: /^GitHub/ }).filter({ has: page.getByRole('button', { name: 'Install' }) }).first();
    await expect(githubCard).toBeVisible();

    // Click Install
    await githubCard.getByRole('button', { name: 'Install' }).click();

    // Verify Dialog Open
    await expect(page.getByRole('heading', { name: 'Instantiate Service' })).toBeVisible();

    // Verify Smart Schema Injection
    // Check for the label specifically
    await expect(page.locator('label', { hasText: 'Personal Access Token' })).toBeVisible();

    // Verify Command
    await expect(page.getByText('npx -y @modelcontextprotocol/server-github')).toBeVisible();
  });
});
