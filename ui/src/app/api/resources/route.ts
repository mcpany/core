
import { NextResponse } from "next/server";
import { db } from "@/lib/mock-data";

export async function GET() {
  return NextResponse.json({ resources: db.resources });
}

export async function POST(request: Request) {
  try {
      const text = await request.text();
      const body = text ? JSON.parse(text) : {};

      // Toggle status
      if (body.hasOwnProperty('enabled')) {
         const index = db.resources.findIndex(r => r.uri === body.uri);
         if (index >= 0) {
             db.resources[index].enabled = body.enabled;
         }
         return NextResponse.json({ success: true });
      }

      return NextResponse.json({ success: false });
  } catch (e) {
       return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
}
