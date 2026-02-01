/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('system uptime chart renders real data', async ({ page }) => {
    // 1. Seed Health Data
    // Generate 20 points, mostly OK, one error
    const now = Date.now();
    const historyPoints = [];
    for (let i = 0; i < 20; i++) {
        historyPoints.push({
            timestamp: now - (20 - i) * 60000, // 1 min intervals
            status: i === 10 ? 'error' : 'ok'
        });
    }

    // Seed via API
    const response = await page.request.post('/api/v1/debug/seed_health', {
        data: {
            "system": historyPoints
        }
    });
    expect(response.status()).toBe(200);

    // 2. Go to Dashboard
    await page.goto('/');

    // 3. Verify Chart
    // The chart title is "System Uptime"
    await expect(page.getByText('System Uptime')).toBeVisible();

    // The uptime percentage should reflect the error.
    // 1 error out of 20 = 5% down => 95% uptime.
    // Logic: (19/20)*100 = 95.0%
    // Note: The Dashboard loads widgets. We might need to wait or drag-and-drop if it's empty?
    // "Default widgets for a fresh dashboard" includes "uptime".
    // Check `dashboard-grid.tsx`:
    /*
    const DEFAULT_LAYOUT: WidgetInstance[] = WIDGET_DEFINITIONS.map(def => ({
        instanceId: crypto.randomUUID(),
        type: def.type,
        title: def.title,
        size: def.defaultSize,
        hidden: false
    }));
    */
    // And `widget-registry.tsx` has `uptime` ("System Uptime").
    // So it should be present by default.

    // Wait for text to appear (handling loading state)
    await expect(page.getByText('95.0% Overall Uptime')).toBeVisible({ timeout: 10000 });
});
