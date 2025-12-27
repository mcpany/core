import { NextResponse } from 'next/server';

let webhooks = [
    { id: "wh-1", url: "https://example.com/webhook", events: ["service.started", "service.stopped"], active: true },
];

export async function GET() {
  return NextResponse.json(webhooks);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.url) {
        webhooks.push({ id: `wh-${Date.now()}`, url: body.url, events: body.events || [], active: true });
        return NextResponse.json({ message: "Created" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
