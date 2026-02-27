import { test, expect } from '@playwright/test';

const MOCK_SERVICE_NAME = 'security-test-service';

test.describe('Service Security Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // 1. Register a service with provenance data
    await page.goto('/upstream-services');

    // Seed service via API directly to ensure specific provenance data
    await page.evaluate(async (name) => {
      const response = await fetch('/api/v1/services', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: name,
          name: name,
          version: '1.0.0',
          http_service: {
            address: 'https://example.com'
          },
          provenance: {
            verified: true,
            signer_identity: 'admin@corp.com',
            attestation_time: new Date().toISOString(),
            signature_algorithm: 'ECDSA-SHA256'
          },
          tool_export_policy: {
            default_action: 2, // UNEXPORT
            rules: [
              { name_regex: '^public_.*', action: 1 } // EXPORT
            ]
          },
          call_policies: [
            {
              default_action: 0, // ALLOW
              rules: [
                { name_regex: 'delete_user', action: 1 } // DENY
              ]
            }
          ]
        })
      });
      if (!response.ok) throw new Error('Failed to seed service');
    }, MOCK_SERVICE_NAME);

    // Refresh page to see the service
    await page.reload();
  });

  test.afterEach(async ({ page }) => {
    // Cleanup
    await page.evaluate(async (name) => {
      await fetch(`/api/v1/services/${name}`, { method: 'DELETE' });
    }, MOCK_SERVICE_NAME);
  });

  test('should display provenance and policy data', async ({ page }) => {
    // 1. Navigate to service detail
    await page.click(`text=${MOCK_SERVICE_NAME}`);
    await expect(page).toHaveURL(new RegExp(`/upstream-services/${MOCK_SERVICE_NAME}`));

    // 2. Click "Security" tab
    await page.click('button[role="tab"]:has-text("Security")');

    // 3. Verify Provenance Card
    await expect(page.locator('text=Supply Chain Attestation')).toBeVisible();
    await expect(page.locator('text=Verified').first()).toBeVisible();
    await expect(page.locator('text=admin@corp.com')).toBeVisible();
    await expect(page.locator('text=ECDSA-SHA256')).toBeVisible();

    // 4. Verify Export Policies (DLP)
    await expect(page.locator('text=Data Loss Prevention (DLP)')).toBeVisible();
    // Check for the rule we added
    await expect(page.locator('text=^public_.*')).toBeVisible();
    await expect(page.locator('text=Export').first()).toBeVisible();

    // 5. Verify Network Rules
    await expect(page.locator('text=Network & Call Policies')).toBeVisible();
    await expect(page.locator('text=delete_user')).toBeVisible();
    await expect(page.locator('text=Deny').first()).toBeVisible();
  });
});
