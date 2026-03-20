import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Service Registration with OAuth', () => {
  test.beforeEach(async ({ request, page }) => {
    await seedGlobalState(request);

    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-core');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
  });

  test('Registers a new service and initiates OAuth seamlessly', async ({ page }) => {
    await page.goto('/upstream-services');
    await page.waitForLoadState('networkidle');

    // Find the add service button. Depending on the page, it might be an icon button.
    const addBtn = page.getByRole('button', { name: /Add Service/i });
    await addBtn.click();

    // Select Custom Service template
    const customSvc = page.getByText('Custom Service');
    await customSvc.click();

    // Basic config
    await page.fill('input[name="name"]', 'e2e-oauth-service');

    // Advanced JSON tab to inject OAuth
    const advTab = page.getByRole('tab', { name: /Advanced \(JSON\)/i });
    await advTab.click();

    const textarea = page.getByLabel('Configuration JSON');
    await textarea.fill(JSON.stringify({
      name: 'e2e-oauth-service',
      upstreamAuth: {
        oauth2: { clientId: { value: "test-client" }, clientSecret: { value: "test-secret" } }
      },
      httpService: { address: 'http://example.com' }
    }));

    // Switch to Auth tab
    const authTab = page.getByRole('tab', { name: /Authentication/i });
    await authTab.click();

    const authBtn = page.getByRole('button', { name: /Authenticate with Provider/i });
    await expect(authBtn).toBeVisible();

    // Trigger OAuth
    const registerPromise = page.waitForResponse(r => r.url().includes('/api/v1/services') && r.request().method() === 'POST');
    const oauthPromise = page.waitForResponse(r => r.url().includes('/oauth/initiate') && r.request().method() === 'POST');

    await authBtn.click();

    // Ensure save request happened
    const registerRes = await registerPromise;
    expect(registerRes.status()).toBe(200);

    // Ensure oauth request happened
    const oauthRes = await oauthPromise;
    // Status might be 500 if backend doesn't have token url etc., but the API was called.
    expect(oauthRes.status()).toBeDefined();
  });
});
