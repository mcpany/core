/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from "next/server";

const tools = [
  { name: "stripe_charge", description: "Create a charge on Stripe", service: "Payment Gateway", type: "function" },
  { name: "get_user", description: "Retrieve user details", service: "User Service", type: "function" },
  { name: "search_docs", description: "Search internal documentation", service: "Search Indexer", type: "read" },
];

export async function GET() {
  return NextResponse.json(tools);
}
