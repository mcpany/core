/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';
import { seedServices, cleanupServices, seedTraffic } from './e2e/test-data';

test.describe('Feature Screenshot', () => {
    // Skip if CAPTURE_SCREENSHOTS is not set
    // Enabled audit screenshots
    // test.skip(process.env.CAPTURE_SCREENSHOTS !== 'true', 'Skipping audit screenshots');

    const date = new Date().toISOString().split('T')[0];
    // Use test-results directory which is writable in CI
    const auditDir = path.join(process.cwd(), 'test-results/artifacts/audit/ui', date);

    test.beforeAll(async ({ request }) => {
        try {
            if (!fs.existsSync(auditDir)) {
                fs.mkdirSync(auditDir, { recursive: true });
            }
        } catch (e) {
            console.warn('Failed to create audit directory:', e);
        }
        await seedServices(request);
        await seedTraffic(request);
    });

    test.afterAll(async ({ request }) => {
        await cleanupServices(request);
    });

  test('Capture Logs', async ({ page }) => {
    await page.goto('/logs');
    // Wait for some logs to appear
    await page.waitForTimeout(3000);
    try {
        await page.screenshot({ path: path.join(auditDir, 'log_stream.png') });
    } catch (e) {
        console.warn('Failed to save screenshot:', e);
    }
  });
});
