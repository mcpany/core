/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Middleware Page', () => {

  test('should display rate limiter middleware', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.getByText('Rate Limiter').first()).toBeVisible();
    // Verify the card structure (icon, name, type, switch, settings)
    const card = page.locator('.bg-card', { hasText: 'Rate Limiter' }).first();
    await expect(card.getByRole('switch')).toBeVisible();
    await expect(card.getByRole('button').filter({ has: page.locator('svg.lucide-settings') })).toBeVisible();
  });
});
