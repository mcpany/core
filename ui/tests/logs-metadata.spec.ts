import { test, expect } from '@playwright/test';

test.describe('Logs Metadata', () => {
  test('should display metadata badges', async ({ page }) => {
    // Mock the WebSocket connection
    await page.routeWebSocket(/\/api\/v1\/ws\/logs/, ws => {
      const logEntry = {
        id: "meta-test-1",
        timestamp: new Date().toISOString(),
        level: "INFO",
        message: "User login",
        source: "auth-service",
        metadata: {
            method: "POST",
            path: "/api/login",
            user_id: 12345,
            is_admin: true
        }
      };
      ws.send(JSON.stringify(logEntry));
    });

    await page.goto('/logs');

    // Wait for the log message to appear
    await expect(page.getByText("User login")).toBeVisible();

    // Now assert metadata visibility (This is what we expect to FAIL currently)
    // We expect the UI to render metadata keys and values
    await expect(page.getByText("method")).toBeVisible();
    await expect(page.getByText("POST")).toBeVisible();
    await expect(page.getByText("path")).toBeVisible();
    await expect(page.getByText("/api/login")).toBeVisible();
    await expect(page.getByText("user_id")).toBeVisible();
    await expect(page.getByText("12345")).toBeVisible();
  });
});
