import { NextResponse } from 'next/server';

let prompts = [
    { name: "summarize_text", description: "Summarizes the given text", enabled: true, serviceName: "gpt-4-proxy" },
    { name: "code_review", description: "Reviews code for bugs", enabled: true, serviceName: "gpt-4-proxy" },
];

export async function GET() {
  return NextResponse.json(prompts);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        prompts = prompts.map(p => p.name === body.name ? { ...p, enabled: body.enabled } : p);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
