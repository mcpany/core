import { FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  const seedingUrl = process.env.BACKEND_URL || 'http://localhost:50050';

  console.log('Seeding backend state...');
  try {
    const res = await fetch(`${seedingUrl}/api/v1/debug/seed_state`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': process.env.MCPANY_API_KEY || 'test-token'
        }
    });

    if (res.ok) {
        console.log('Backend seeded successfully.');
    } else {
        console.error('Failed to seed backend:', res.status, res.statusText);
    }
  } catch (e) {
      console.error('Error connecting to backend for seeding:', e);
  }
}

export default globalSetup;
