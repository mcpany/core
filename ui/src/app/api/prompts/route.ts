
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json([
    { id: "p1", name: "summarize_text", description: "Summarizes the input text", enabled: true },
    { id: "p2", name: "generate_code", description: "Generates code based on spec", enabled: true },
    { id: "p3", name: "explain_error", description: "Explains a stack trace", enabled: true },
  ]);
}

export async function POST(request: Request) {
    const body = await request.json();
    return NextResponse.json({ success: true, ...body });
}
