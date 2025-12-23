
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json([
    { id: "wh_01", url: "https://ops.example.com/alerts", events: ["service.down", "error.critical"], active: true },
    { id: "wh_02", url: "https://audit.example.com/logs", events: ["*"], active: false },
  ]);
}

export async function POST(request: Request) {
    const body = await request.json();
    return NextResponse.json({ success: true, ...body });
}
