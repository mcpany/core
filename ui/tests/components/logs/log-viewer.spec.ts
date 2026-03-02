/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('LogViewer Component Documentation Types', () => {
    test('renders LogViewer correctly without crashing', async ({ page }) => {
        // Just verify it doesn't crash from the TSDoc change
        await page.goto('/logs');
        await expect(page.getByText('Log Stream', { exact: false })).toBeVisible();
    });
});
