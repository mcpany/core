import { test, expect } from '@playwright/test';

test.describe('Tools Management', () => {
  const serviceName = 'test-tool-mgmt-svc';
  const toolName = 'echo';

  test.beforeEach(async ({ request }) => {
    // 1. Delete service if it exists to ensure a clean state
    await request.delete(`/api/v1/services/${serviceName}`, {
        headers: { 'Authorization': 'Bearer test-key' }
    });

    // 2. Seed the database with a valid service that has an explicit tool
    const config = {
      name: serviceName,
      http_service: {
        address: "http://127.0.0.1:50050/health",
        tools: [
          {
            name: toolName,
            disable: false,
            description: "An echo tool"
          }
        ]
      }
    };

    const response = await request.post('/api/v1/services', {
      data: config,
      headers: {
        'Authorization': 'Bearer test-key',
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok()) {
      console.log(await response.text());
    }
    if (!response.ok()) {
      console.log(await response.text());
    }
    expect(response.ok()).toBeTruthy();

    // Give backend a moment to reload config
    await new Promise(r => setTimeout(r, 1000));
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`, {
        headers: { 'Authorization': 'Bearer test-key' }
    });
  });

  test.skip('can toggle tool disable state and persist it', async ({ page, request }) => {
    // Navigate to Tools page
    await page.goto('/tools');

    // Wait for the tool to be present in the table
    const searchInput = page.getByPlaceholder("Search tools...");
    await expect(searchInput).toBeVisible();
    await searchInput.fill(toolName);

    // Wait for search
    await page.waitForTimeout(1000);

    const toolRow = page.locator(`tr:has-text("${toolName}")`);
    await expect(toolRow).toBeVisible({ timeout: 10000 });

    // Verify it is enabled initially
    const toggleSwitch = toolRow.locator('button[role="switch"]');
    await expect(toggleSwitch).toBeVisible();

    const initialState = await toggleSwitch.getAttribute('aria-checked');
    const expectedState = initialState === 'true' ? 'false' : 'true';

    // Intercept the API call to ensure we click and wait for response
    const requestPromise = page.waitForResponse(response => response.url().includes('/api/v1/tools') && response.status() === 200);

    // Click toggle to disable
    await toggleSwitch.click({ force: true });

    // Wait a moment for network roundtrip to finish saving
    const res = await requestPromise;
    if (!res.ok()) {
       console.log("Failed PUT", await res.text());
    }
    await page.waitForTimeout(500);

    // Verify UI updates optimistic state
    await expect(toggleSwitch).toHaveAttribute('aria-checked', expectedState);

    // Verify the backend API reflects the change (The database was seeded and actually saved)
    const toolsRes = await request.get('/api/v1/tools', {
        headers: { 'Authorization': 'Bearer test-key' }
    });
    expect(toolsRes.ok()).toBeTruthy();
    const tools = await toolsRes.json();
    const myTool = tools.find((t: any) => t.name === toolName && t.serviceId === serviceName);

    expect(myTool).toBeDefined();
    expect(myTool.disable).toBe(expectedState === 'false');

    // Reload the page to verify it persisted across sessions
    await page.reload();

    const reloadedSearchInput = page.getByPlaceholder("Search tools...");
    await expect(reloadedSearchInput).toBeVisible();
    await reloadedSearchInput.fill(toolName);
    await page.waitForTimeout(1000);

    const reloadedRow = page.locator(`tr:has-text("${toolName}")`);
    await expect(reloadedRow).toBeVisible();
    const reloadedToggle = reloadedRow.locator('button[role="switch"]');
    await expect(reloadedToggle).toHaveAttribute('aria-checked', expectedState);
  });
});