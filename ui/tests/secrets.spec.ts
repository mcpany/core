/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';

test.describe('Secrets Manager', () => {
    test('should allow creating, viewing, and deleting a secret', async ({ page }) => {
        // Navigate to settings
        await page.goto('http://localhost:9002/settings');

        // Click on Secrets tab
        await page.click('text=Secrets & Keys');

        // Check empty state or list
        // Assuming empty state initially or list

        // Add Secret
        await page.click('text=Add Secret');
        await page.fill('input[placeholder="e.g. Production OpenAI Key"]', 'E2E Test Secret');
        await page.fill('input[placeholder="e.g. OPENAI_API_KEY"]', 'E2E_TEST_KEY');
        await page.fill('input[placeholder="sk-..."]', 'sk-test-value-123');
        await page.click('button:has-text("Save Secret")');

        // Verify Secret is present
        await expect(page.locator('text=E2E Test Secret')).toBeVisible();
        await expect(page.locator('text=E2E_TEST_KEY')).toBeVisible();

        // Verify secret value is hidden
        const secretValue = page.locator('text=••••••••••••••••••••••••');
        await expect(secretValue).toBeVisible();

        // Reveal secret
        // Click the eye icon. It's inside a button.
        // We can find the button that is a sibling of the masked text
        await page.locator('button:has(.lucide-eye)').click();

        // Verify value is revealed
        await expect(page.locator('text=sk-test-value-123')).toBeVisible();

        // Screenshot verification
        await page.screenshot({ path: '.audit/ui/2025-12-29/api_key_manager.png' });

        // Delete Secret
        await page.click('button[aria-label="Delete secret"]');

        // Verify Secret is gone
        await expect(page.locator('text=E2E Test Secret')).not.toBeVisible();
    });
});
