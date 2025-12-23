
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json([
    { name: "Payment Gateway", status: "healthy", latency: "120ms" },
    { name: "User Service", status: "healthy", latency: "45ms" },
    { name: "Legacy Auth", status: "degraded", latency: "850ms" },
    { name: "Notification Service", status: "healthy", latency: "80ms" },
    { name: "Search Index", status: "down", latency: "-" },
    { name: "Data Pipeline", status: "healthy", latency: "200ms" },
  ]);
}
