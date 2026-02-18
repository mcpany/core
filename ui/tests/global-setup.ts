/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
  const apiKey = process.env.MCPANY_API_KEY || 'test-token';

  console.log(`Resetting database at ${backendUrl}...`);
  try {
    const res = await fetch(`${backendUrl}/debug/reset`, {
      method: 'POST',
      headers: {
        'X-API-Key': apiKey,
      },
    });

    if (!res.ok) {
      console.warn(`Failed to reset database: ${res.status} ${res.statusText}`);
      // Don't fail setup, as backend might not support it yet or be unreachable if starting up
    } else {
      console.log('Database reset successful.');
    }
  } catch (error) {
    console.warn('Error resetting database:', error);
  }
}

export default globalSetup;
