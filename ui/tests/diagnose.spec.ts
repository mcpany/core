/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('Diagnose Network', async ({ page }) => {
    page.on('request', request => console.log('>>', request.method(), request.url()));
    page.on('response', response => console.log('<<', response.status(), response.url()));
    page.on('console', msg => console.log('LOG:', msg.text()));

    await page.goto('/services');

    // Wait for a bit
    await page.waitForTimeout(5000);

    // Check heading
    const heading = page.getByRole('heading', { name: 'Services' });
    console.log('Heading visible:', await heading.isVisible());

    // Check content
    const gateway = page.getByText('Payment Gateway');
    console.log('Gateway visible:', await gateway.isVisible());
});
