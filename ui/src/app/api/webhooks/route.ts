
import { NextResponse } from 'next/server';

export async function GET() {
  const webhooks = [
    { id: "wh_1", url: "https://example.com/webhook", events: ["service.up", "service.down"], active: true },
    { id: "wh_2", url: "https://slack.com/api/webhook/...", events: ["alert.critical"], active: true },
  ];
  return NextResponse.json(webhooks);
}
