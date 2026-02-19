import { test, expect } from '@playwright/test';

test('visualizer renders live agent flow', async ({ page }) => {
  // 1. Seed: Execute a tool in the Playground
  await page.goto('/playground');

  // Wait for tools to load
  // We try to find weather-service.get_weather, assuming the minimal config is loaded.
  // If not, we might need another strategy, but get_weather is standard in this repo's demo.
  await expect(page.getByText('weather-service.get_weather')).toBeVisible({ timeout: 15000 });

  // Type command
  await page.getByPlaceholder('Enter command or select a tool...').fill('weather-service.get_weather {"city": "Paris"}');
  await page.keyboard.press('Enter');

  // Wait for result
  await expect(page.getByText('Result: weather-service.get_weather')).toBeVisible({ timeout: 15000 });

  // 2. Navigate to Visualizer
  await page.goto('/visualizer');

  // 3. Verify Graph
  // We expect "Live Mode" to be on
  await expect(page.getByText('Live Mode')).toBeVisible();

  // Wait for polling to pick up the trace (interval is 2s)
  // We expect a node with label "get_weather" (the helper strips service name if distinct? no, helper logic: if service return serviceName || name. if tool return name.)
  // getParticipantLabel for tool returns span.name. So it will be "weather-service.get_weather"
  await expect(page.getByText('weather-service.get_weather', { exact: true })).toBeVisible({ timeout: 10000 });

  // We also expect "User" node
  await expect(page.getByText('User', { exact: true })).toBeVisible();

  // We also expect "MCP Core" node
  await expect(page.getByText('MCP Core')).toBeVisible();

  // Take screenshot for verification
  await page.screenshot({ path: 'visualizer_verification.png' });
});
