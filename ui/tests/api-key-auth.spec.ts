/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { cleanupUser } from './e2e/test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
// This Master Key is used to seed the user (admin access)
const MASTER_KEY = process.env.MCPANY_API_KEY || 'test-token';
const ADMIN_HEADERS = { 'X-API-Key': MASTER_KEY };

const TEST_USER = "service-account-test";
const TEST_API_KEY = "mcp_sk_test_1234567890";

test.describe('API Key Authentication', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeAll(async ({ request }) => {
      // 1. Clean up potential leftovers
      await cleanupUser(request, TEST_USER);

      // 2. Seed a user with API Key Authentication
      const user = {
          id: TEST_USER,
          authentication: {
              api_key: {
                  param_name: "X-API-Key",
                  in: 0, // HEADER
                  verification_value: TEST_API_KEY
              }
          },
          roles: ["viewer"]
      };

      const res = await request.post(`${BASE_URL}/api/v1/users`, {
          data: user,
          headers: ADMIN_HEADERS
      });
      expect(res.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
      await cleanupUser(request, TEST_USER);
  });

  test('Should authenticate with User API Key', async ({ request }) => {
      // 3. Attempt to access a protected endpoint (e.g. list services) using the User API Key
      const res = await request.get(`${BASE_URL}/api/v1/services`, {
          headers: {
              'X-API-Key': TEST_API_KEY
          }
      });

      expect(res.status()).toBe(200);
      const data = await res.json();
      expect(Array.isArray(data.service_states)).toBeTruthy();
  });

  test('Should fail with invalid API Key', async ({ request }) => {
      const res = await request.get(`${BASE_URL}/api/v1/services`, {
          headers: {
              'X-API-Key': 'invalid-key'
          }
      });

      expect(res.status()).toBe(401); // Unauthorized or 500 depending on implementation details, but usually 401/500
      // The auth manager returns error, middleware might map it.
      // Based on AuthMiddleware, if err != nil, it returns error.
      // The HTTP transport likely converts error to 500 or 401.
      // Let's assume non-200 for now.
      expect(res.ok()).toBeFalsy();
  });
});
