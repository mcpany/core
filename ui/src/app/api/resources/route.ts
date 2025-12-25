/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  const resources = [
    { name: "users_db", type: "postgresql", service: "DBService" },
    { name: "logs_bucket", type: "s3", service: "LogService" },
  ];
  return NextResponse.json(resources);
}
