import { test, expect } from '@playwright/test';

test.describe('Tool Usage Graph', () => {
    test.beforeEach(async ({ page }) => {
        // Abort gRPC calls to force REST fallback
        await page.route(/RegistrationService\/GetService/, async (route) => {
             await route.abort();
        });

        // Mock listServices (sidebar)
        await page.route(/\/api\/v1\/services$/, async (route) => {
             await route.fulfill({ json: { services: [] } });
        });

        // Mock service detail (REST)
        await page.route(/\/api\/v1\/services\/weather-service$/, async (route) => {
             await route.fulfill({
                json: {
                    service: {
                        id: 'weather-service',
                        name: 'weather-service',
                        http_service: {
                            address: 'http://localhost',
                            tools: [
                                {
                                    name: 'weather-tool',
                                    description: 'Get weather',
                                    input_schema: {}
                                }
                            ]
                        },
                    }
                }
            });
        });

        // Mock status
        await page.route(/\/api\/v1\/services\/weather-service\/status$/, async (route) => {
            await route.fulfill({
                json: {
                    metrics: {
                        'tool_usage:weather-tool': 42
                    }
                }
            });
        });

        // Mock audit logs
        await page.route(/\/api\/v1\/audit\/logs/, async (route) => {
             await route.fulfill({
                json: {
                    entries: [
                        {
                            timestamp: new Date().toISOString(),
                            tool_name: 'weather-tool',
                            duration_ms: 120,
                        },
                        {
                            timestamp: new Date(Date.now() - 3600000).toISOString(),
                            tool_name: 'weather-tool',
                            duration_ms: 80,
                             error: "Timeout"
                        }
                    ]
                }
            });
        });
    });

    test('should display usage chart on tool detail page', async ({ page }) => {
        // Navigate to tool detail page
        await page.goto('/service/weather-service/tool/weather-tool');

        // Check static metrics
        await expect(page.getByText('Total Calls')).toBeVisible();
        await expect(page.getByText('42')).toBeVisible();

        // Check Chart Components
        await expect(page.getByText('Usage Over Time')).toBeVisible();

        // Check Dropdown
        const combobox = page.getByRole('combobox');
        await expect(combobox).toBeVisible();
        await expect(combobox).toContainText('Last 24 Hours');

        // Check if Recharts rendered something (use first to avoid strict mode error)
        await expect(page.locator('.recharts-surface').first()).toBeVisible();

        // Check for legend items
        await expect(page.getByText('Executions')).toBeVisible();
        await expect(page.getByText('Avg Latency')).toBeVisible();
    });
});
