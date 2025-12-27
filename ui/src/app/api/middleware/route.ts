import { NextResponse } from 'next/server';

let middlewares = [
    { id: "mw-1", name: "Authentication", type: "auth", enabled: true, order: 1 },
    { id: "mw-2", name: "Rate Limiter", type: "rate_limit", enabled: true, order: 2 },
    { id: "mw-3", name: "Logging", type: "logging", enabled: true, order: 3 },
];

export async function GET() {
  return NextResponse.json(middlewares);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.action === 'toggle' && body.id) {
        middlewares = middlewares.map(m => m.id === body.id ? { ...m, enabled: body.enabled } : m);
        return NextResponse.json({ message: "Updated" });
    }
     if (body.action === 'reorder' && body.order) {
        // Mock reorder
        return NextResponse.json({ message: "Reordered" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
