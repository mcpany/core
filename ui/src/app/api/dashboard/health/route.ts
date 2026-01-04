
import { NextResponse } from "next/server";
import { db } from "@/lib/mock-data";

export async function GET() {
  const services = db.services.map(s => {
    // Determine status based on 'disable' flag + some randomness for the demo
    let status = "healthy";
    let uptime = "99.99%";
    let latency = "24ms";

    if (s.disable) {
        status = "unhealthy"; // Or 'disabled' if we had that state, but UI uses healthy/degraded/unhealthy
        uptime = "0.00%";
        latency = "N/A";
    } else if (s.name.includes("Legacy")) {
        status = "degraded";
        latency = "450ms";
        uptime = "98.50%";
    }

    return {
      id: s.id,
      name: s.name,
      status,
      latency,
      uptime
    };
  });

  return NextResponse.json(services);
}
