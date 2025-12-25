
import { NextResponse } from 'next/server';

export async function GET() {
  const resources = [
    { name: "users_db", type: "postgresql", service: "DBService" },
    { name: "logs_bucket", type: "s3", service: "LogService" },
  ];
  return NextResponse.json(resources);
}
