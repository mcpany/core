import { request } from '@playwright/test';

export async function seedDatabase() {
  const context = await request.newContext();
  const baseURL = process.env.BACKEND_URL || 'http://localhost:50050';
  const apiKey = process.env.MCPANY_API_KEY || ''; // If empty, server might block if auth enabled, or allow if disabled.

  // 1. Seed Traffic Data
  const trafficData = [
    { timestamp: new Date(Date.now() - 3600000).toISOString(), count: 10, error_count: 0 },
    { timestamp: new Date(Date.now() - 1800000).toISOString(), count: 25, error_count: 1 },
    { timestamp: new Date().toISOString(), count: 50, error_count: 2 },
  ];

  const trafficRes = await context.post(`${baseURL}/api/v1/debug/seed_traffic`, {
    data: trafficData,
    headers: {
        'X-API-Key': apiKey
    }
  });

  if (!trafficRes.ok()) {
      console.warn(`Failed to seed traffic: ${trafficRes.status()} ${await trafficRes.text()}`);
  }

  // 2. Register Service (if not exists)
  const serviceConfig = {
      name: "e2e-test-service",
      id: "e2e-test-service",
      http_service: {
          address: "http://localhost:8081"
      },
      tools: [
          { name: "test-tool", call_id: "test-call", description: "A test tool" }
      ],
      calls: {
          "test-call": {
              id: "test-call",
              endpoint_path: "/test",
              method: "POST"
          }
      }
  };

  const serviceRes = await context.post(`${baseURL}/api/v1/services`, {
      data: serviceConfig,
      headers: { 'X-API-Key': apiKey }
  });

  if (!serviceRes.ok()) {
      const text = await serviceRes.text();
      // Ignore if it fails (e.g. already exists or validation error), tests might still pass if pre-seeded
      console.log(`Service registration response: ${serviceRes.status()} ${text}`);
  } else {
      console.log("Seeded e2e-test-service");
  }
}
