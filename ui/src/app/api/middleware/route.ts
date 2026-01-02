/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

/**
 * Handles GET requests to retrieve the list of configured middlewares.
 * @returns A JSON response containing an array of middleware configurations.
 */
export async function GET() {
  const middlewares = [
    { name: "auth", priority: 1, disabled: false, description: "Authenticates requests" },
    { name: "logging", priority: 2, disabled: false, description: "Logs all requests" },
    { name: "rate_limit", priority: 3, disabled: true, description: "Limits request rate" },
  ];
  return NextResponse.json(middlewares);
}
