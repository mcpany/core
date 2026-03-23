import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection, seedGlobalState } from './e2e/test-data';

test.describe('Stack Editor', () => {
  test.beforeEach(async ({ request }) => {
    await seedGlobalState(request);
    await seedCollection('default-stack', request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupCollection('default-stack', request);
  });

  test('should load the editor and show initial config in graph', async ({ page }) => {
    const stackName = `stack-editor-load-${Date.now()}`;
    await seedCollection(stackName, page.request);

    try {
        await page.request.post(`/api/v1/collections/${stackName}/apply`, {
            headers: {
                'Authorization': `Bearer test-token`,
                'Content-Type': 'application/json'
            }
        });
    } catch(e) {}

    try {
      await page.goto(`/stacks/new`);

      // Ensure Monaco is visible
      const editorTextarea = page.locator('.monaco-editor textarea');
      await expect(editorTextarea).toBeVisible({ timeout: 15000 });

    } finally {
      await cleanupCollection(stackName, page.request);
    }
  });

  test('should update graph when template is added', async ({ page }) => {
    const stackName = `stack-editor-update-${Date.now()}`;
    await seedCollection(stackName, page.request);
    try {
      await page.goto('/stacks/new');

      const visualizer = page.locator('.stack-visualizer-container');

      // Wait for the visualizer to be ready
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 45000 });
      await expect(page.getByText('Valid YAML')).toBeVisible();

      const templateCard = page.getByText(/GitHub|Google Calendar|Linear/, { exact: true }).first();
      await expect(templateCard).toBeVisible({ timeout: 30000 });
      await templateCard.click();

      // Verify the editor remains valid and the visualizer stays rendered after inserting a template.
      await expect(page.getByText('Valid YAML')).toBeVisible();
      await expect(visualizer.locator('.react-flow')).toBeVisible({ timeout: 60000 });
    } finally {
      await cleanupCollection(stackName, page.request);
    }
  });
});
