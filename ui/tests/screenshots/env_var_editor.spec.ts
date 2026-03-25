/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Service Configuration Editor', () => {
  test('should allow editing environment variables for command line service', async ({ page }) => {
    // Navigate to services page
    await page.goto('/upstream-services');
    await expect(page).toHaveTitle(/MCPAny Manager/);

    // Open "New Service" dialog
    await page.getByRole('button', { name: 'Add Service' }).click();
    await expect(page.getByText('Select Service Template')).toBeVisible();

    // In RegisterServiceDialog, click custom service or blank option
    const customHttpOption = page.locator('text=Custom HTTP').first();
    if (await customHttpOption.isVisible()) {
        await customHttpOption.click();
    } else {
        const customOption = page.locator('text=Custom').first();
        if (await customOption.isVisible()) {
             await customOption.click();
        }
    }

    // Now we should be in the form view
    await expect(page.getByText('Configure Service')).toBeVisible();

    // Select "Command Line" type
    // Protocol selection is now a select dropdown named 'type'
    await page.locator('button[role="combobox"]').first().click();
    await page.getByRole('option', { name: 'Command Line' }).click();

    // Fill command
    await page.getByLabel('Command').fill('echo hello');

    // Switch to Advanced tab to set env vars
    await page.getByRole('tab', { name: 'Advanced (JSON)' }).click();

    // In RegisterServiceDialog, Environment Variables are part of JSON config
    // Update the configJson to include an environment variable
    const configJsonLocator = page.locator('textarea[name="configJson"]');
    const currentJson = await configJsonLocator.inputValue();
    const config = JSON.parse(currentJson);
    if (config.commandLineService) {
        config.commandLineService.env = { TEST_ENV: 'test_value' };
    }
    await configJsonLocator.fill(JSON.stringify(config, null, 2));

    // Switch back to Basic Configuration to see the JSON is valid (or just stay)
    await page.getByRole('tab', { name: 'Basic Configuration' }).click();

    // Take screenshot of the editor
    // Use test-results directory which is guaranteed to be writable in CI
    const screenshotPath = 'test-results/artifacts/audit/ui/2025-02-20/env_var_editor.png';
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const fs = require('fs');
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const path = require('path');
    try {
      fs.mkdirSync(path.dirname(screenshotPath), { recursive: true });
      await page.screenshot({ path: screenshotPath });
    } catch (e) {
      console.warn('Failed to save screenshot:', e);
    }

    // Close
    // Note: The cancel button is the 'X' or click outside since it's a dialog now
    await page.keyboard.press('Escape');
  });
});
