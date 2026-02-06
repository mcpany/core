/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

/**
 * Handles the GET request.
 *
 * @summary Handles the GET request.
 * @returns {any} The result.
 */
export async function GET() {
  const webhooks = [
    { id: "wh_1", url: "https://example.com/webhook", events: ["service.up", "service.down"], active: true },
    { id: "wh_2", url: "https://slack.com/api/webhook/...", events: ["alert.critical"], active: true },
  ];
  return NextResponse.json(webhooks);
}
