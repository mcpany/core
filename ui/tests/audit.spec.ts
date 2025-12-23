
import { test, expect } from '@playwright/test';
import path from 'path';

const date = new Date().toISOString().split('T')[0];
const auditDir = path.join(__dirname, `../../.audit/ui/${date}`);

test.describe('Audit Screenshots', () => {
    test.use({ viewport: { width: 1920, height: 1080 } });

    test('capture screenshots', async ({ page }) => {
        // Dashboard
        await page.goto('/');
        await page.waitForTimeout(1000); // Allow animations/charts to settle
        await page.screenshot({ path: path.join(auditDir, 'dashboard.png'), fullPage: true });

        // Services
        await page.goto('/services');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'services.png'), fullPage: true });

        // Tools
        await page.goto('/tools');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'tools.png'), fullPage: true });

        // Resources
        await page.goto('/resources');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'resources.png'), fullPage: true });

        // Prompts
        await page.goto('/prompts');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'prompts.png'), fullPage: true });

        // Middleware
        await page.goto('/settings/middleware');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'middleware.png'), fullPage: true });

        // Webhooks
        await page.goto('/settings/webhooks');
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(auditDir, 'webhooks.png'), fullPage: true });
    });
});
