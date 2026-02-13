import { test, expect } from '@playwright/test';

test('Onboarding Wizard Flow', async ({ page, request }) => {
  // 1. Check initial state
  const servicesRes = await request.get('/api/v1/services');
  const services = await servicesRes.json();
  const list = Array.isArray(services) ? services : (services.services || []);
  const initialCount = list.length;

  console.log(`Initial service count: ${initialCount}`);

  await page.goto('/');

  // Allow for loading
  await expect(page.locator('body')).toBeVisible();

  if (initialCount === 0) {
    // Step 1 should be visible
    await expect(page.getByText('Welcome to MCP Any!')).toBeVisible();
    // Use a more specific locator if class assertion is flaky, but checking text is good
    await expect(page.getByText('1. Add Service').first()).toBeVisible();
  } else {
    // Step 2 should be visible (You're almost there!)
    await expect(page.getByText("You're almost there!")).toBeVisible();
    await expect(page.getByText('2. Connect Client').first()).toBeVisible();
  }

  // 2. Add a service if we were in Step 1 to verify transition
  if (initialCount === 0) {
    console.log('Seeding service to test transition...');
    const createRes = await request.post('/api/v1/services', {
      data: {
        id: 'onboarding-test-service',
        name: 'onboarding-test-service',
        http_service: { address: 'https://example.com' }
      }
    });
    expect(createRes.ok()).toBeTruthy();

    await page.reload();
    await expect(page.getByText("You're almost there!")).toBeVisible();
  }

  // 3. Verify Copy Config (Step 2 content)
  // Ensure "Claude Desktop" tab is active (default)
  await expect(page.getByRole('tab', { name: 'Claude Desktop' })).toBeVisible();
  await expect(page.locator('pre').first()).toContainText('server-sse-client');

  // 4. Verify other tabs
  await page.getByRole('tab', { name: 'Cursor' }).click();
  await expect(page.getByText('Copy URL')).toBeVisible();

  // Cleanup
  if (initialCount === 0) {
      console.log('Cleaning up seeded service...');
      await request.delete('/api/v1/services/onboarding-test-service');
  }
});
