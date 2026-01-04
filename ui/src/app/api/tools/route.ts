
import { NextResponse } from "next/server";
import { db } from "@/lib/mock-data";

export async function GET() {
  return NextResponse.json({ tools: db.tools });
}

export async function POST(request: Request) {
  try {
      const text = await request.text();
      if (!text) {
         return NextResponse.json({ error: "Empty body" }, { status: 400 });
      }
      const body = JSON.parse(text);

      // Toggle status
      if (body.hasOwnProperty('enabled')) {
         const index = db.tools.findIndex(t => t.name === body.name);
         if (index >= 0) {
             db.tools[index].enabled = body.enabled;
         }
         return NextResponse.json({ success: true });
      }

      return NextResponse.json({ success: false });
  } catch (e) {
      return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
}
