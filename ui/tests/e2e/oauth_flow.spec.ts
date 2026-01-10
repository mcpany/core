/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('OAuth Flow E2E', () => {
  const serviceID = 'github-test';


  test('should initiate oauth and handle callback', async ({ page }) => {
    // 1. Mock Service Details to include OAuth config
    await page.route(`**/api/v1/registration.RegistrationService/GetService`, async () => {
         // gRPC-Web response or REST?
         // logic in client.ts uses `registrationClient.GetService` which is gRPC-Web.
         // Mocking gRPC-Web in Playwright can be tricky if using binary format.
         // BUT `client.ts` uses `GrpcWebImpl` which usually POSTs to `/.../GetService`.
         // If we use standard gRPC-Web text/proto, mocking might be hard.
         // However, `apiClient.listServices` is implemented via REST `/api/v1/services`.
         // `ServiceDetail` uses `getService` which uses `apiClient.getService`.
         // `apiClient.getService` calls `registrationClient.GetService`.
         // Can we fallback/force REST or just mock the network response if it was REST?
         // The mock for `ServiceDetail` page might need careful handling.
         // Let's assume for this test, we navigate to the service detail page.

         // Wait, `ServiceDetail` component calls `apiClient.getService(serviceId)`.
         // `client.ts` says: `const registrationClient = new RegistrationServiceClientImpl(rpc);`
         // It sends a POST to `http://localhost:9002/mcpany.config.v1.registration.RegistrationService/GetService` (or similar).
         // Protocol is gRPC-Web. Content-Type: application/grpc-web-text or application/grpc-web+proto.
         // Mocking that is painful without helpers.

         // ALTERNATIVE:
         // Navigate to `/services` (List) which uses REST `/api/v1/services`.
         // If we click a service there, does it go to Detail? Yes.
         // Can we test "List" page showing "Connect" button?
         // No, the button is in `ServiceDetail`.

         // Maybe we can skip `ServiceDetail` component test if mocking gRPC is too hard,
         // AND only test the `oauth/callback` page directly?
         // The user wants "walk through the oauth process via UI".
         // This implies clicking the button.

         // If I look at `ui/src/lib/client.ts`, `getService` is gRPC.
         // Is there any other way?
         // Maybe I can patch `apiClient.getService` in the browser context?
         // Playwright allows `page.addInitScript`.
         // But `apiClient` is imported by the component.
         // It's hard to mock imports.

         // What if I blindly mock the POST request with a valid gRPC-Web response?
         // It's base64 encoded protobuf.
         // That's too fragile.

         // Let's look at `auth.spec.ts` again. It mocks `api/v1/users`.
         // `listServices` is REST.
         // Maybe I can just use `listServices`?
         // No, `ServiceDetail` specifically calls `getService`.

         // WORKAROUND:
         // I'll skip the Service Detail page for now if I can't mock it easily
         // OR I update `client.ts` to fallback to REST for `getService` too?
         // The comment in `client.ts` says: `// Services (Migrated to gRPC)`.
         // But `listServices` has fallback.
         // `getService` does NOT.
         // I will ADD fallback to `getService` in `client.ts` to make it testable/robust.
         // `api/v1/services/{name}` REST endpoint should exist if `listServices` uses REST.
         // `server/pkg/api/api.go` usually registers them.

         // Let's assume I added the fallback (I will do it next).
    });

    // Mock REST GetService (assuming I added fallback)
    await page.route(`**/api/v1/services/${serviceID}`, async route => {
        await route.fulfill({
            json: {
                service: {
                    id: serviceID,
                    name: serviceID,
                    upstream_auth: {
                        oauth2: {
                            provider: 'github'
                        }
                    }
                }
            }
        });
    });

    // Mock Initiate OAuth
    await page.route('**/auth/oauth/initiate', async route => {
        // Expect POST
        const req = route.request().postDataJSON();
        expect(req.service_id).toBe(serviceID);
        expect(req.redirect_url).toContain('/auth/callback');

        await route.fulfill({
            json: {
                authorization_url: 'http://localhost:9002/mock-provider-auth?state=xyz',
                state: 'xyz'
            }
        });
    });

    // Mock Callback Endpoint (Backend)
    await page.route('**/auth/oauth/callback', async route => {
        const req = route.request().postDataJSON();
        expect(req.code).toBe('mock-code');
        expect(req.service_id).toBe(serviceID);

        await route.fulfill({ json: { status: 'success' } });
    });


    // Start Test
    // 1. Go to Service Detail
    await page.goto(`/service/${serviceID}`);
    // Wait for "Connect Account" button
    await expect(page.getByRole('button', { name: 'Connect Account' })).toBeVisible();

    // 2. Click Connect
    // This will redirect to authorization_url.
    // We need to intercept the navigation or mock the provider page.
    // The provider page is 'http://localhost:9002/mock-provider-auth?state=xyz'
    // Let's route that location.
    await page.route('**/mock-provider-auth*', async route => {
         // Simulate user approving and redirecting back
         const url = new URL(route.request().url());
         const state = url.searchParams.get('state');

         // Redirect back to callback
         // Playwright route.fulfill with status 302?
         // Or just navigate manually?
         // Browsers follow 302.
         await route.fulfill({
             status: 302,
             headers: {
                 'Location': `http://localhost:9002/auth/callback?code=mock-code&state=${state}`
             }
         });
    });

    await page.getByRole('button', { name: 'Connect Account' }).click();

    // 3. Should end up at /auth/callback
    await page.waitForURL('**/auth/callback*');

    // 4. Verification
    await expect(page.getByText('Authentication Successful')).toBeVisible({ timeout: 10000 });
  });
});
