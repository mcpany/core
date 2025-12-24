
import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json({ profiles: ["default", "dev", "prod"] });
}
