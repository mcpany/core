
import { NextResponse } from 'next/server';

export async function GET() {
  const settings = {
    mcp_listen_address: ":8080",
    log_level: "LOG_LEVEL_INFO",
    api_key: "****************",
    profiles: ["default", "dev"],
    allowed_ips: ["127.0.0.1", "10.0.0.0/8"]
  };
  return NextResponse.json(settings);
}
