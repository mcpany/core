
import { NextResponse } from "next/server";
import { db } from "@/lib/mock-data";

export async function GET() {
  return NextResponse.json(db.profiles);
}

export async function POST(request: Request) {
  try {
      const text = await request.text();
      const body = text ? JSON.parse(text) : {};

      if (body.action === 'delete') {
          db.profiles = db.profiles.filter(p => p.id !== body.id);
          return NextResponse.json({ success: true });
      }

      // Create or Update
      const newProfile = {
          id: body.id || `prof-${Date.now()}`,
          ...body
      };

      const index = db.profiles.findIndex(p => p.id === newProfile.id);
      if (index >= 0) {
          db.profiles[index] = newProfile;
      } else {
          db.profiles.push(newProfile);
      }

      return NextResponse.json(newProfile);
  } catch (e) {
      return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
}
