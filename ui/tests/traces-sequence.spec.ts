/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Trace Sequence Diagram', () => {
    test('should render sequence diagram in trace detail', async ({ page }) => {
        // Mock traces list
        await page.route('/api/traces', async route => {
            await route.fulfill({
                json: [
                    {
                        id: 'trace-123',
                        timestamp: new Date().toISOString(),
                        totalDuration: 100,
                        status: 'success',
                        rootSpan: {
                            id: 'span-1',
                            name: 'test-tool',
                            type: 'tool',
                            startTime: 0,
                            endTime: 100,
                            status: 'success',
                            children: [
                                {
                                    id: 'span-2',
                                    name: 'backend-service',
                                    type: 'service',
                                    startTime: 10,
                                    endTime: 90,
                                    status: 'success'
                                }
                            ]
                        }
                    }
                ]
            });
        });

        await page.goto('/traces');

        // Wait for tab to appear
        await expect(page.getByRole('tab', { name: 'Sequence' })).toBeVisible({ timeout: 10000 });

        // Click tab
        await page.getByRole('tab', { name: 'Sequence' }).click();

        // Check for SVG presence
        const svg = page.locator('svg.font-sans');
        await expect(svg).toBeVisible();

        // Check for Actors inside the SVG
        await expect(svg.getByText('User / Client')).toBeVisible();
        // Use first() because the name appears in the Actor Box and potentially as a Message Label
        await expect(svg.getByText('test-tool').first()).toBeVisible();
        await expect(svg.getByText('backend-service').first()).toBeVisible();
    });
});
