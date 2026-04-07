import { test, expect } from '@playwright/test';

// Define headers for the mock API requests
const HEADERS = {
  'Content-Type': 'application/json',
};

test.describe('Lazy MCP Dashboard', () => {
  const SERVICE_ID = 'test-lazy-mcp-service';

  test.beforeEach(async ({ request }) => {
    // Delete any existing service with this ID
    await request.delete(`/api/v1/services/${SERVICE_ID}`).catch(() => {});

    // Seed a mock service via API
    const response = await request.post('/api/v1/services', {
      data: {
        id: SERVICE_ID,
        name: 'Lazy MCP Test Service',
        version: '1.0.0',
        priority: 10,
        disable: false,
        command_line_service: {
            command: 'go run server/cmd/mock_mcp_server/main.go',
            working_directory: '.',
            env: {}
        }
      },
      headers: HEADERS
    });

    expect(response.status()).toBe(200);

    // Wait for the service to be loaded and its tools to be available
    // For this, we poll the tools endpoint until we see "read_file" or "list_directory" etc.
    let toolsLoaded = false;
    for (let i = 0; i < 20; i++) {
      const toolsRes = await request.get('/api/v1/tools');
      if (toolsRes.status() === 200) {
        const toolsData = await toolsRes.json();
        // Our mock server provides "read_file" tool
        if (toolsData.some((t: any) => t.name === 'read_file')) {
           toolsLoaded = true;
           break;
        }
      }
      await new Promise(resolve => setTimeout(resolve, 500));
    }
    expect(toolsLoaded).toBeTruthy();
  });

  test.afterEach(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${SERVICE_ID}`);
  });

  test('should filter tools by intent correctly using the real backend', async ({ page }) => {
    await page.goto('/universal-agent-bus');

    // Make sure the lazy MCP dashboard is visible
    await expect(page.getByText('Lazy-MCP Tool Search Dashboard')).toBeVisible();

    const searchInput = page.getByPlaceholder('Search tool index by intent...');
    const searchButton = page.getByRole('button', { name: 'Search' });

    // Search for "read"
    await searchInput.fill('read');
    await searchButton.click();

    // Verify loading state appears
    // wait for response or UI update
    await expect(page.getByText('read_file')).toBeVisible();

    // The results should NOT contain other tools like "list_directory"
    await expect(page.getByText('list_directory')).not.toBeVisible();

    // Now search for "directory"
    await searchInput.fill('directory');
    await searchButton.click();

    await expect(page.getByText('list_directory')).toBeVisible();
    await expect(page.getByText('read_file')).not.toBeVisible();

    // Now search for nonsense
    await searchInput.fill('asdfqwerzxcv');
    await searchButton.click();
    await expect(page.getByText('No tools found matching your intent.')).toBeVisible();
  });
});
