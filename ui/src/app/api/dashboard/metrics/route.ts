
import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json({
    totalRequests: 245231,
    requestsChange: 20.1,
    avgLatency: "45ms",
    latencyChange: -10,
    activeServices: 12,
    servicesChange: 2,
    resourcesServed: 573,
    resourcesChange: 201,
    requestVolume: [
        { time: "10:00", reqs: 400 },
        { time: "10:05", reqs: 300 },
        { time: "10:10", reqs: 550 },
        { time: "10:15", reqs: 450 },
        { time: "10:20", reqs: 600 },
        { time: "10:25", reqs: 700 },
        { time: "10:30", reqs: 650 },
    ]
  });
}
