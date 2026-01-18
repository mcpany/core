/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const AUDIT_DIR = path.resolve(__dirname, '../../.audit/ui/2026-01-18');

if (!fs.existsSync(AUDIT_DIR)) {
  fs.mkdirSync(AUDIT_DIR, { recursive: true });
}

test('Capture Tool History Persistence and Replay', async ({ page }) => {
    page.on('console', msg => console.log(`PAGE LOG: ${msg.text()}`));
    page.on('pageerror', err => console.log(`PAGE ERROR: ${err.message}`));

    // Mock System Status to avoid error banner covering UI
    await page.route('**/healthz', async route => {
         await route.fulfill({ status: 200, body: 'ok' });
     });
     await page.route('**/api/v1/health', async route => {
         await route.fulfill({ status: 200, json: { status: 'ok' } });
     });
     await page.route('**/doctor', async route => {
         await route.fulfill({
             status: 200,
             contentType: 'application/json',
             body: JSON.stringify({ status: 'healthy', checks: {} })
         });
     });

    // Mock Tools
    await page.route('**/api/v1/tools', async route => {
         await route.fulfill({
             json: {
                 tools: [
                     { name: 'echo', description: 'Echo back', inputSchema: { type: 'object', properties: { msg: { type: 'string' } } } }
                 ]
             }
         });
    });

    // Mock Execution
    await page.route('**/api/v1/execute', async route => {
         await route.fulfill({
             json: { result: "hello persistence" }
         });
    });

    await page.goto('/playground');

    // Simulate user typing and sending a command
    await page.getByPlaceholder(/Enter command/).fill('echo {"msg": "hello persistence"}');
    await page.keyboard.press('Enter');

    // Wait for the message to appear
    await expect(page.getByText('echo {"msg": "hello persistence"}')).toBeVisible();

    // Reload page to verify persistence
    await page.reload();
    await expect(page.getByText('echo {"msg": "hello persistence"}')).toBeVisible();

    // Hover over the tool call to show Replay button
    const toolCallCard = page.locator('.group\\/card').first(); // The card has 'group/card' class
    await toolCallCard.hover();

    // Wait for hover effect
    await page.waitForTimeout(500);

    // Capture screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'persistent_tool_history.png'), fullPage: true });
});
