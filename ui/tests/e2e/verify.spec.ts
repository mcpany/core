import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedCollection, cleanupCollection } from './test-data';

test('Take screenshot of Recent Activity', async ({ page, request }) => {
    await seedCollection('mcpany-system', request);
    await seedUser(request, "e2e-admin-screenshot");

    // Login
    await page.goto('http://localhost:9002/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', "e2e-admin-screenshot");
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('http://localhost:9002/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);

    // Seed data
    const execRes = await request.post('http://localhost:9002/api/v1/execute', {
        data: {
            name: 'echo',
            arguments: { message: 'Hello from verification script' }
        },
        headers: {
            'Content-Type': 'application/json'
        }
    });
    expect(execRes.ok()).toBeTruthy();

    await page.goto('http://localhost:9002/');

    const recentActivityCard = page.getByText('Recent Activity').first();
    await expect(recentActivityCard).toBeVisible({ timeout: 15000 });

    const echoActivity = page.getByText('echo').first();
    await expect(echoActivity).toBeVisible({ timeout: 15000 });

    await page.screenshot({ path: '/home/jules/verification/recent-activity.png' });

    await cleanupCollection('mcpany-system', request);
});