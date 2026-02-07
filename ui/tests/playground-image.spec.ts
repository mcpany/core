import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
  const serviceName = 'image-test-service';
  const toolName = 'get_image';
  // 1x1 red pixel
  const base64Image = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=";
  const jsonOutput = JSON.stringify([
    {
      type: "image",
      data: base64Image,
      mimeType: "image/png"
    }
  ]);

  test.beforeAll(async ({ request }) => {
    // Clean up if exists
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Register service
    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
          command: 'echo',
          tools: [
            {
              name: toolName,
              description: 'Returns an image',
              call_id: '1',
              input_schema: {
                type: "object",
                properties: {}
              }
            }
          ],
          calls: {
            '1': {
                args: [jsonOutput]
            }
          }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
     await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should render image in tool result', async ({ page }) => {
    // Go to playground
    await page.goto('/playground');

    // Wait for services/tools to load
    // We can wait for the tool to appear in the sidebar
    // Sidebar might filter tools, so let's use the filter input
    await page.waitForSelector('input[placeholder="Enter command or select a tool..."]');

    // Sidebar filter is separate.
    const filterInput = page.getByPlaceholder('Search tools...');
    await filterInput.fill(toolName);

    // Click the tool in the list
    // The list items show tool name.
    await page.click(`text=${toolName}`);

    // Config dialog opens. Click "Build Command"
    await page.click('button:has-text("Build Command")');

    // Dialog closes, input is populated. Click "Send"
    await page.click('button[aria-label="Send"]');

    // Wait for result
    // Look for image with specific src
    const srcPrefix = `data:image/png;base64,${base64Image}`;

    // We increase timeout because tool execution + rendering might take a bit
    await expect(page.locator(`img[src="${srcPrefix}"]`)).toBeVisible({ timeout: 10000 });
  });
});
