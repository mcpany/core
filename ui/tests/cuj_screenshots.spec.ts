/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

// Write directly to the repo's docs folder so that Bazel will write them out
// Even better, write them locally or use the artifact directory.
const testOutDir = process.env.TEST_UNDECLARED_OUTPUTS_DIR || process.cwd();

test.describe('Generate Detailed Docs Screenshots CUJs', () => {

    test('Capture Full Page and Features Screenshots CUJ', async ({ page }) => {
        const baseURL = process.env.BACKEND_URL || 'http://127.0.0.1:50050';
        const apiKey = process.env.MCPANY_API_KEY || 'test-token';

        // 1. Seed realistic data into the backend without mocking
        await page.request.post(`${baseURL}/api/v1/services`, {
            headers: { 'X-API-Key': apiKey, 'Content-Type': 'application/json' },
            data: {
                name: "Database Admin",
                mcp_service: {
                    url: "http://postgres-admin:8080"
                }
            }
        });

        // Seed Dynamic Mesh Resilience (DMR) Hub
        await page.request.post(`${baseURL}/api/v1/services`, {
            headers: { 'X-API-Key': apiKey, 'Content-Type': 'application/json' },
            data: {
                name: "Dynamic Mesh Resilience (DMR) Hub",
                mcp_service: {
                    url: "http://dmr-hub:8080"
                }
            }
        });

        // 2. Drive UI
        await page.goto('/upstream-services');
        // Wait for page to fully load and API calls to complete
        await page.waitForTimeout(3000);

        // Screenshot: Service List
        await page.screenshot({ path: path.join(testOutDir, 'cuj_service_list.png'), fullPage: true });

        // Screenshot: Audit Dashboard
        await page.goto('/audit');
        await page.waitForTimeout(2000);
        await page.screenshot({ path: path.join(testOutDir, 'cuj_audit_logs.png'), fullPage: true });

        // Screenshot: Inspector Dashboard
        await page.goto('/inspector');
        await page.waitForTimeout(2000);
        await page.screenshot({ path: path.join(testOutDir, 'cuj_service_inspector.png'), fullPage: true });

        // Screenshot: Connect Client dialog
        await page.goto('/');
        await page.waitForTimeout(2000);
        const connectBtn = page.getByRole('button', { name: 'Connect' }).first();
        if (await connectBtn.isVisible()) {
            await connectBtn.click();
            await page.waitForTimeout(500);
            await page.screenshot({ path: path.join(testOutDir, 'cuj_connect-client.png') });
            await page.keyboard.press('Escape');
        } else {
            await page.screenshot({ path: path.join(testOutDir, 'cuj_connect-client.png'), fullPage: true });
        }

    });

});
