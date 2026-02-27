// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import { request, expect } from '@playwright/test';

async function globalSetup(config) {
  const { baseURL, storageState } = config.projects[0].use;

  // We need to bypass auth or use a known key.
  // Ideally, the server is started with MCPANY_API_KEY=test-token via playwright.config.ts env
  const apiKey = process.env.MCPANY_API_KEY || 'test-token';

  // Construct a comprehensive seed payload
  // This matches what we want in the database for E2E tests
  const seedData = {
    upstream_services: [
        {
            id: "echo-service",
            name: "Echo Service",
            version: "1.0.0",
            disable: false,
            command_line_service: {
                command: "echo",
                working_directory: ".",
                env: {}
            }
        }
    ],
    // Add default user if needed, though server usually has one
    users: [
        {
            id: "admin",
            username: "admin",
            roles: ["admin"]
        }
    ]
  };

  const requestContext = await request.newContext({
    baseURL,
    extraHTTPHeaders: {
      'X-API-Key': apiKey,
      'Content-Type': 'application/json'
    },
  });

  // Wait for server to be ready (health check)
  // This is technically redundant if webServer.reuseExistingServer is false and it waits for url,
  // but good practice.
  let retries = 10;
  while (retries > 0) {
      try {
          const health = await requestContext.get('/health');
          if (health.ok()) break;
      } catch (e) {
          // ignore
      }
      await new Promise(r => setTimeout(r, 1000));
      retries--;
  }

  console.log('Seeding database via /api/v1/debug/seed...');
  const response = await requestContext.post('/api/v1/debug/seed', {
    data: seedData,
  });

  if (!response.ok()) {
      console.error('Seeding failed:', await response.text());
      throw new Error('Failed to seed database in global setup');
  }

  console.log('Database seeded successfully.');
  await requestContext.dispose();
}

export default globalSetup;
