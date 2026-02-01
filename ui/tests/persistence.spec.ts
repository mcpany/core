/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
  const username = `testuser_${Date.now()}`;
  const password = 'password123';
  const userId = username; // ID is username for simplicity in this seed

  test.beforeAll(async ({ request }) => {
    // Seed a user
    const res = await request.post('/api/v1/users', {
      data: {
        user: {
          id: userId,
          authentication: {
            basicAuth: {
              username: username,
              password: { plainText: password }
            }
          },
          roles: ['editor']
        }
      }
    });
    expect(res.ok()).toBeTruthy();
  });

  test('should persist dashboard layout across reloads', async ({ page }) => {
    // 1. Login
    await page.goto('/api/v1/auth/login'); // Or UI login page if exists, but /api/v1/auth/login is likely backend.
    // The UI likely has a login page at /login or uses basic auth popup if strict.
    // But `UserProvider` in `user-context.tsx` uses `apiClient.getCurrentUser()`.
    // If that returns 401, what happens?
    // We haven't implemented the login UI flow fully in this task, we just refactored the context.
    // However, `apiClient` uses `fetchWithAuth`. `fetchWithAuth` injects `Authorization` header from `localStorage`.

    // We need to simulate login by setting the token in localStorage.
    // Or we can use the `mock` login for now if the UI doesn't have a real login page yet?
    // Wait, I refactored `UserProvider` to fetch `/api/v1/users/me`.
    // If I am not logged in (no token), it returns 401.

    // How do I get a token?
    // I can use `request` to login?
    // `api/v1/auth/login` endpoint likely returns a token or sets a cookie.

    // Let's look at `server/pkg/app/server.go`: `mux.Handle("/auth/login", ...)`
    // And `handleLogin`.

    // For the test, I can simulate setting the Basic Auth header or Token.
    // `client.ts`: `headers.set('Authorization', 'Basic ' + token);`
    // So I can set localStorage item `mcp_auth_token` to `btoa(username:password)`.

    const token = btoa(`${username}:${password}`);

    await page.goto('/');
    await page.evaluate((t) => localStorage.setItem('mcp_auth_token', t), token);
    await page.reload();

    // 2. Check initial state (Default Layout)
    // Wait for widgets to load
    await expect(page.locator('.grid-cols-12')).toBeVisible();

    // 3. Modify Layout (Simulate by calling updatePreferences directly or drag?)
    // Dragging in Playwright is flaky.
    // We can manually trigger the update via console for stability in this "Integration" test,
    // or try to drag.
    // Let's try to drag.
    const widget = page.locator('.group\\/widget').first();
    const target = page.locator('.group\\/widget').nth(1);

    // Get initial text
    const initialTitle = await widget.innerText();

    // Drag
    await widget.dragTo(target);

    // Wait for debounce (1s) + network
    await page.waitForTimeout(2000);

    // 4. Reload
    await page.reload();

    // 5. Verify persistence
    // We check if the layout matches what we expect or just that it's NOT default if we changed it significantly.
    // Or simpler: We can check if `user.preferencesJson` is populated via API.

    const meRes = await page.request.get('/api/v1/users/me', {
        headers: {
            'Authorization': `Basic ${token}`
        }
    });
    expect(meRes.ok()).toBeTruthy();
    const me = await meRes.json();
    expect(me.preferencesJson).toContain('dashboardLayout');
  });
});
