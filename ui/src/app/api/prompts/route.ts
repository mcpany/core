
import { NextResponse } from "next/server";

const prompts = [
  { name: "code_review", description: "Review code for best practices", service: "Code Assistant" },
  { name: "summarize_email", description: "Summarize a long email thread", service: "Email Service" },
];

export async function GET() {
  return NextResponse.json(prompts);
}
