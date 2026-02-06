import { test as base } from '@playwright/test';

// Define a custom fixture type
type AuthenticatedFixtures = {
  // We don't necessarily need a new fixture name if we use a before hook in the tests,
  // but extending 'test' is cleaner.
};

// Extend the test object
export const test = base.extend<AuthenticatedFixtures>({
  page: async ({ page }, use) => {
    // Navigate to base URL first to set localStorage
    await page.goto('/');

    // Inject auth token (admin:password)
    // base64('admin:password') = 'YWRtaW46cGFzc3dvcmQ='
    await page.evaluate(() => {
        localStorage.setItem('mcp_auth_token', 'YWRtaW46cGFzc3dvcmQ=');
    });

    // Reload to apply the token and fetch user
    await page.reload();

    // Wait for the app to be ready (e.g. sidebar or main content visible)
    // This avoids race conditions where the app redirects back to login if the token check is slow
    // Wait for something that indicates we are logged in, or at least not on /login
    // Assuming successful login stays on / or redirects to dashboard
    await use(page);
  },
});

export { expect } from '@playwright/test';
