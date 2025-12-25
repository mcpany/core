/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json([{ name: "auth", priority: 1, disabled: false }, { name: "logging", priority: 2, disabled: false }]);
}
