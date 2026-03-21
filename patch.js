const fs = require('fs');
let code = fs.readFileSync('ui/tests/onboarding.spec.ts', 'utf-8');

code = code.replace(
    /test\('shows dashboard when services exist', async \(\{ page, request \}\) => \{[\s\S]*?\}\);/m,
    `test('shows dashboard when services exist', async ({ page, request }) => {
    // Seed a service
    await seedCollection('mcpany-system', request);

    // Explicitly apply the collection to trigger service registration on the backend
    try {
        await request.post('/api/v1/collections/mcpany-system/apply', {
            headers: {
                'Authorization': \`Bearer test-token\`,
                'Content-Type': 'application/json'
            }
        });
    } catch(e) {}

    await page.waitForTimeout(2000); // Give backend a moment to process the service
    await page.goto('/');

    // Wait a bit and check content to debug if necessary
    try {
        await expect(page.getByRole('heading', { name: /Dashboard/i })).toBeVisible({ timeout: 15000 });
    } catch (e) {
        console.error("Dashboard not visible! HTML content:");
        console.error(await page.content());
        throw e;
    }

    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();

    // Cleanup
    await cleanupCollection('mcpany-system', request);
    try {
        await request.delete('/api/v1/services/weather-service', {
            headers: {
                'Authorization': \`Bearer test-token\`
            }
        });
    } catch(e) {}
  });`
);

fs.writeFileSync('ui/tests/onboarding.spec.ts', code);
