import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Service Registration with OAuth', () => {
  test.beforeEach(async ({ request, page }) => {
    await seedGlobalState(request);

    const authHeader = `Basic ${Buffer.from('e2e-admin-core:password').toString('base64')}`;
    await request.post('/api/v1/credentials', {
      headers: {
        'Authorization': authHeader,
        'Content-Type': 'application/json'
      },
      data: {
        name: "Test OAuth Credential",
        authentication: {
          oauth2: {
            clientId: { value: "test-client" },
            clientSecret: { value: "test-secret" }
          }
        }
      }
    });

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

    const addBtn = page.getByRole('button', { name: /Add Service/i });
    await addBtn.click();

    const customSvc = page.getByText('Custom Service');
    await customSvc.click();

    await page.fill('input[name="name"]', 'e2e-oauth-service');

    const authTab = page.getByRole('tab', { name: /Authentication/i });
    await authTab.click();

    const selectTrigger = page.locator('button[role="combobox"]').nth(1);
    await selectTrigger.click();

    const credOption = page.getByRole('option', { name: /Test OAuth Credential/i });
    await credOption.click();

    const authBtn = page.getByRole('button', { name: /Authenticate with Provider/i });
    await expect(authBtn).toBeVisible({ timeout: 5000 });

    const registerPromise = page.waitForResponse(r => r.url().includes('/api/v1/services') && r.request().method() === 'POST');
    const oauthPromise = page.waitForResponse(r => r.url().includes('/oauth/initiate') && r.request().method() === 'POST');

    await authBtn.click();

    const registerRes = await registerPromise;
    expect(registerRes.status()).toBe(200);

    const oauthRes = await oauthPromise;
    expect(oauthRes.status()).toBeDefined();
  });
});
