import { test, expect } from '@playwright/test';

test.describe('Upstream Auth Display', () => {
  const credentialName = 'e2e-auth-display-cred';

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test credential using the backend API directly
    const response = await request.post('/api/v1/credentials', {
      data: {
        name: credentialName,
        authentication: {
          oauth2: {
            client_id: { plainText: 'test-client-id-e2e' },
            scopes: 'read,write,admin',
            token_url: 'https://auth.example.com/oauth/token',
            grant_type: 'authorization_code'
          }
        }
      }
    });

    if (!response.ok()) {
       console.log("Failed to seed credential:", await response.text());
    }
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Cleanup the seeded credential
    const listRes = await request.get('/api/v1/credentials');
    if (listRes.ok()) {
      const credentials = await listRes.json();
      const target = credentials.find((c: any) => c.name === credentialName);
      if (target && target.id) {
        await request.delete(`/api/v1/credentials/${target.id}`);
      }
    }
  });

  test('should display AuthConfigurationCard instead of raw JSON dump when a credential is selected', async ({ page }) => {
    // Go to the Upstream Services page
    await page.goto('/upstream-services');

    // Wait for the table to load
    await expect(page.getByRole('button', { name: 'Add Service' })).toBeVisible();

    // Click "Add Service" button to open the Register Service dialog
    await page.getByRole('button', { name: 'Add Service' }).click();

    // Verify dialog opened
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('New Service')).toBeVisible(); // Using the correct header

    // Navigate to Advanced tab or Authentication step (using 'Advanced' tab structure per component)
    await page.getByRole('tab', { name: 'Advanced' }).click();

    // There should be a "Select From Credentials" dropdown or combo box. Let's find it.
    // The dialog fetches credentials. We need to wait for it.
    await page.getByRole('button', { name: /Select Credential/i }).click();

    // Select the credential we seeded
    await page.getByRole('option', { name: credentialName }).click();

    // Now, the AuthConfigurationCard should render instead of a raw JSON block

    // 1. Verify "Authentication Details" title is visible (part of the new Card)
    await expect(page.getByText('Authentication Details')).toBeVisible();

    // 2. Verify Auth Type Badge "OAuth2"
    await expect(page.getByText('OAuth2')).toBeVisible();

    // 3. Verify specific fields from the credential we seeded
    await expect(page.getByText('test-client-id-e2e')).toBeVisible();
    await expect(page.getByText('read,write,admin')).toBeVisible();
    await expect(page.getByText('https://auth.example.com/oauth/token')).toBeVisible();

    // 4. Verify raw JSON dump structure is ABSENT.
    // The previous implementation had: {JSON.stringify(form.watch("upstreamAuth"), null, 2)}
    // which would display things like "oauth2": {
    await expect(page.locator('pre')).not.toContainText('"oauth2": {');
    await expect(page.locator('pre')).not.toContainText('"client_id":');
    await expect(page.locator('pre')).not.toContainText('"clientId":');
  });
});
