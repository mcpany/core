
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json([
    { id: "r1", name: "System Logs", uri: "file:///var/log/system.log", mime_type: "text/plain", enabled: true },
    { id: "r2", name: "User Manual", uri: "https://docs.example.com/manual", mime_type: "text/html", enabled: true },
    { id: "r3", name: "Config Schema", uri: "internal://schema/config", mime_type: "application/json", enabled: false },
  ]);
}

export async function POST(request: Request) {
    const body = await request.json();
    return NextResponse.json({ success: true, ...body });
}
