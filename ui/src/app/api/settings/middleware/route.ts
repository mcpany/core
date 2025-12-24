
import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json([{ name: "auth", priority: 1, disabled: false }, { name: "logging", priority: 2, disabled: false }]);
}
