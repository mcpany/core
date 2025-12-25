
import { NextResponse } from 'next/server';

export async function GET() {
  const prompts = [
    { name: "summarize_text", description: "Summarizes the given text", service: "LLMService" },
    { name: "code_review", description: "Reviews code for bugs", service: "LLMService" },
  ];
  return NextResponse.json(prompts);
}
