/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// This test assumes the backend is running with wttr.in config
// ./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml

test('Real E2E: Tools page inspector form works with wttr.in', async ({ page }) => {
  // 1. Go to Tools page
  await page.goto('/tools');

  // 2. Wait for tools to load (real network request)
  await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });

  // 3. Open inspector for get_weather (target the wttr.in service one)
  // The service name might be sanitized to wttrin_... so we filter by that or "wttr.in" if it shows up in description
  await page.locator('tr').filter({ hasText: 'wttrin' }).filter({ hasText: 'get_weather' }).getByText('Inspect').click();

  // 4. Verify Inspector is open
  await expect(page.getByRole('dialog')).toBeVisible();
  // Use first() because the title might appear multiple times or be less specific,
  // but we know we clicked the right one.
  await expect(page.getByRole('dialog').getByText('get_weather').first()).toBeVisible();

  // 5. Verify "Form" mode is active by default (DynamicForm)
  await expect(page.getByRole('button', { name: 'Form' })).toHaveClass(/secondary/); // Active
  // Check for the specific fields from wttr.in schema
  await expect(page.getByLabel('location', { exact: true })).toBeVisible();
  await expect(page.getByLabel('lang', { exact: true })).toBeVisible();

  // 6. Fill out the form
  await page.getByLabel('location', { exact: true }).fill('Paris');
  // lang has a default 'en', let's leave it or change it
  await page.getByLabel('lang', { exact: true }).fill('fr');

  // 7. Toggle to JSON to verify sync (Optional, but good for regression)
  await page.getByRole('button', { name: 'JSON' }).click();
  const jsonContent = await page.locator('textarea#args').inputValue();
  const parsed = JSON.parse(jsonContent);
  expect(parsed).toEqual({ location: 'Paris', lang: 'fr' });

  // Switch back to Form
  await page.getByRole('button', { name: 'Form' }).click();

  // 8. Execute
  await page.getByRole('button', { name: 'Execute' }).click();

  // 9. Verify Result
  // It should be real weather data for Paris in French
  await expect(page.getByText('Result')).toBeVisible();
  // We expect a successful JSON response. wttr.in usually returns complex JSON.
  // We check for "weatherCode" which indicates valid weather data.
  await expect(page.locator('pre.text-green-600')).toContainText('weatherCode', { timeout: 15000 });

  await page.screenshot({ path: 'tool-inspector-verification.png' });
});
