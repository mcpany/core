
import { NextResponse } from "next/server";
import { db } from "@/lib/mock-data";

export async function GET() {
  return NextResponse.json(db.services);
}

export async function POST(request: Request) {
  try {
    const text = await request.text();
    if (!text) {
        return NextResponse.json({ error: "Empty body" }, { status: 400 });
    }
    const body = JSON.parse(text);

    if (body.action === 'toggle') {
        const index = db.services.findIndex(s => s.name === body.name || s.id === body.name);
        if (index >= 0) {
             db.services[index].disable = body.disable;
             // Also update the dashboard/health by implication since they share the same db object
        }
        return NextResponse.json({ success: true });
    }

    // Create or Update
    const newService = {
        id: body.id || `svc-${db.services.length + 1}`,
        ...body,
    };

    const index = db.services.findIndex(s => s.id === newService.id);
    if (index >= 0) {
        db.services[index] = newService;
    } else {
        db.services.push(newService);
    }

    return NextResponse.json(newService);
  } catch (e) {
      console.error("API Error", e);
      return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
}
