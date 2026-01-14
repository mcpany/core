/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Concurrency & Race Conditions', () => {
  test('should handle multiple concurrent user sessions without state verification issues', async ({ browser }) => {
    // Create 3 independent contexts to simulate 3 users
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
      browser.newContext(),
    ]);

    const pages = await Promise.all(contexts.map(context => context.newPage()));

    // Navigate all pages to home concurrently
    await Promise.all(pages.map(page => page.goto('/')));

    // Verify all loaded
    await Promise.all(pages.map(page => expect(page).toHaveTitle(/MCPAny/)));

    // Perform concurrent actions (e.g. navigation or simple interaction if available)
    // For now, we'll verify they can all reach the playground
    // Use a more specific locator if needed, but for now generic link is fine
    await Promise.all(pages.map(async page => {
        const link = page.getByRole('link', { name: 'Playground' }).first();
        await expect(link).toHaveAttribute('href', '/playground');
        await link.click();
        await expect(page).toHaveURL(/.*playground/);
    }));

    // Check that all pages are on the playground
    await Promise.all(pages.map(page => expect(page).toHaveURL(/.*playground/)));

    // Cleanup
    await Promise.all(contexts.map(context => context.close()));
  });
});
