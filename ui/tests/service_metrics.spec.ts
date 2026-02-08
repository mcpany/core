/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('service metrics dashboard', async ({ page, request }) => {
  // 1. Seed Traffic Data
  // We explicitly seed data that will map to "weather-service" in the backend hack
  const seedData = [
    { time: "10:00", requests: 100, errors: 2, latency: 50 },
    { time: "10:01", requests: 120, errors: 0, latency: 45 },
    { time: "10:02", requests: 80, errors: 5, latency: 60 },
  ];

  const response = await request.post('/api/v1/debug/seed_traffic', {
    data: seedData,
  });
  expect(response.ok()).toBeTruthy();

  // 2. Navigate to Weather Service
  await page.goto('/service/weather-service');

  // 3. Click Metrics Tab
  await page.getByRole('tab', { name: 'Metrics' }).click();

  // 4. Verify Charts are present
  await expect(page.getByText('Request Traffic')).toBeVisible();
  await expect(page.getByText('Avg Latency')).toBeVisible();
  await expect(page.getByText('Success vs Errors')).toBeVisible();

  // 5. Verify data is rendered
  // Since we seeded 100+ requests, the Y-axis should show numbers.
  // We can check if tooltips work or if SVG elements exist.
  // The .recharts-surface class indicates charts are rendered.
  // We wait for data loading.
  await expect(page.locator('.recharts-surface').first()).toBeVisible({ timeout: 10000 });

  // Verify that we are not seeing the "No metrics available" empty state
  await expect(page.getByText('No metrics available for this service yet.')).not.toBeVisible();
});
