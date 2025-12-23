
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json([
    { id: "t1", name: "search_users", description: "Search for users by criteria", enabled: true, service_id: "svc_02" },
    { id: "t2", name: "create_payment", description: "Initiate a payment transaction", enabled: true, service_id: "svc_01" },
    { id: "t3", name: "refund_payment", description: "Refund a transaction", enabled: false, service_id: "svc_01" },
  ]);
}

export async function POST(request: Request) {
    const body = await request.json();
    return NextResponse.json({ success: true, ...body });
}
