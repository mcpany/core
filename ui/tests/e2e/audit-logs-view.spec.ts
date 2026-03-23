import { test, expect } from '@playwright/test';

test.describe('Audit Logs Viewer', () => {
    test.beforeAll(async ({ request }) => {
        // Seed an audit log by registering a service and executing a tool
        await request.post('/api/v1/services', {
            data: {
                name: 'audit-test-service',
                command_line_service: {
                    command: 'echo',
                    tools: [
                        { name: 'echo_tool', call_id: 'echo_call', description: 'Echoes data' }
                    ],
                    calls: {
                        'echo_call': {
                            args: ['{"foo": "bar"}']
                        }
                    }
                }
            }
        });

        // Execute the tool to generate an audit log
        await request.post('/api/v1/execute', {
            data: {
                tool: 'audit-test-service.echo_tool',
                arguments: { "test": "data" }
            }
        });
    });

    test.afterAll(async ({ request }) => {
        await request.delete('/api/v1/services/audit-test-service');
    });

    test('should render JSON objects nicely in the View dialog', async ({ page }) => {
        await page.goto('/audit');

        // Wait for the audit table to load the log
        const row = page.getByRole('row').filter({ hasText: 'audit-test-service.echo_tool' }).first();
        await row.waitFor({ state: 'visible', timeout: 30000 });

        // Click "View"
        await row.getByRole('button', { name: 'View' }).click();

        // Ensure dialog opens
        const dialog = page.getByRole('dialog');
        await expect(dialog).toBeVisible();
        await expect(dialog.getByRole('heading', { name: 'Log Details' })).toBeVisible();

        // Verify the Arguments area renders with JsonView (it has Raw/Tree buttons or just renders keys)
        // Check for 'test' or 'data' in Arguments
        await expect(dialog.getByText('"test"')).toBeVisible();

        // Check for 'foo' or 'bar' in Result (from echo output)
        await expect(dialog.getByText('"foo"')).toBeVisible();

        // Check for JsonView controls if they are visible (e.g. Raw)
        // Only if they are big enough, but just asserting the text is a good start.
    });
});
