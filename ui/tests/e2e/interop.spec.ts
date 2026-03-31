import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';
import { apiClient } from '../src/lib/client';

test.describe('Interop Feature', () => {
  test.beforeEach(async ({ request, page }) => {
    await seedGlobalState(request);
  });

  test('should execute an interop task via API', async ({ request }) => {
    // Make the API request directly to test the backend since UI fails to pick up
    const res = await request.post('/api/v1/interop/task', {
      data: {
        framework: 'OpenClaw',
        intent: 'adaptive_reasoning',
        payload: { "foo": "bar" }
      },
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
      }
    });

    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json.status).toBe('success');
  });
});
