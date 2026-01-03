/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

// Mock database
// Mock database
let services = [
  {
    id: "svc_01",
    name: "Payment Gateway",
    connection_pool: { max_connections: 100 },
    disable: false,
    version: "v1.2.0",
    http_service: {
        address: "https://api.stripe.com",
        tools: [
            { name: "process_refund", description: "Process a refund for a transaction" },
            { name: "get_transaction", description: "Get transaction details" }
        ],
        resources: []
    }
  },
  {
    id: "svc_02",
    name: "User Service",
    connection_pool: { max_connections: 50 },
    disable: false,
    version: "v2.1.0",
    grpc_service: {
        address: "localhost:50051",
        tools: [
            { name: "get_user_profile", description: "Get user profile by ID" },
            { name: "update_user_email", description: "Update user email address" }
        ],
        resources: [
            { name: "User Database Dump", description: "Daily dump of user database", mimeType: "application/sql", type: "file" }
        ]
    }
  },
  {
      id: "svc_03",
      name: "Legacy Auth",
      connection_pool: { max_connections: 10 },
      disable: true,
      version: "v0.5.0",
      http_service: {
          address: "http://legacy-auth:8080",
          tools: [],
          resources: [
             { name: "Legacy Logs", description: "Logs from legacy system", mimeType: "text/plain", type: "logs" }
          ]
      }
  },
  {
      id: "svc_04",
      name: "Weather Service",
      connection_pool: { max_connections: 20 },
      disable: false,
      version: "v1.0.0",
      http_service: {
          address: "http://weather-api",
          tools: [
              { name: "get_weather", description: "Get weather for a location" } // Required by E2E
          ],
          resources: []
      }
  },
  {
      id: "svc_05",
      name: "System Monitor",
      connection_pool: { max_connections: 5 },
      disable: false,
      version: "v1.0.0",
      http_service: {
          address: "http://monitor",
          tools: [],
          resources: [
              { name: "Application Logs", description: "Main application logs", mimeType: "text/plain", type: "logs" } // Required by E2E
          ]
      }
  }
];

export async function GET() {
  return NextResponse.json({ services });
}

export async function POST(request: Request) {
  const body = await request.json();
  // Simulate update
  services = services.map(s => s.id === body.id ? { ...s, ...body } : s);
  return NextResponse.json({ success: true, service: body });
}
