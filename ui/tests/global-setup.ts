/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { spawn } from 'child_process';
import path from 'path';
import seedData from './fixtures/seed-data';

export default async function globalSetup() {
  console.log('Global Setup: Starting...');

  // 1. Start Fake OAuth Server
  console.log('Global Setup: Starting Fake OAuth Server...');
  const oauthServerPath = path.resolve(__dirname, 'fixtures/oauth-server.js');
  const oauthProcess = spawn('node', [oauthServerPath], {
    detached: true,
    stdio: 'inherit' // Pipe output to parent
  });

  // Store the process reference in global (hacky but works for teardown if needed)
  // Playwright doesn't have a clean global teardown for this, but process exit handles it usually.
  // Or we can rely on `webServer` in config if we moved it there.
  // But plan says globalSetup.

  // Give it a second to start
  await new Promise(resolve => setTimeout(resolve, 1000));

  // 2. Wait for Backend
  // We assume backend is started by webServer or docker-compose
  console.log('Global Setup: Waiting for backend...');
  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
  let retries = 30;
  while (retries > 0) {
    try {
      const res = await fetch(`${backendUrl}/healthz`);
      if (res.ok) {
        console.log('Global Setup: Backend is ready.');
        break;
      }
    } catch (e) {
      // ignore
    }
    await new Promise(resolve => setTimeout(resolve, 1000));
    retries--;
  }

  if (retries === 0) {
      console.error('Global Setup: Backend failed to start. Seeding might fail.');
  }

  // 3. Seed Data
  console.log('Global Setup: Seeding data...');
  try {
      await seedData();
      console.log('Global Setup: Data seeded successfully.');
  } catch (e) {
      console.error('Global Setup: Failed to seed data', e);
  }
}
