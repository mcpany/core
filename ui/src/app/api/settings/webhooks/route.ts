/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json([{ id: "wh_01", url: "https://example.com/hook", events: ["service.up"] }]);
}
