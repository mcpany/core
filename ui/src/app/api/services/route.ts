
import { NextResponse } from 'next/server';

let services = [
    {
      id: "svc-1",
      name: "Payment Service",
      version: "1.0.0",
      disable: false,
      service_config: { http_service: { address: "http://localhost:8080" } }
    },
    {
      id: "svc-2",
      name: "Auth Service",
      version: "1.2.0",
      disable: false,
      service_config: { grpc_service: { address: "localhost:9090" } }
    },
    {
      id: "svc-3",
      name: "Legacy Service",
      version: "0.9.0",
      disable: true,
      service_config: { command_line_service: { command: "./run.sh" } }
    }
];

export async function GET() {
  return NextResponse.json(services);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.id && typeof body.disable !== 'undefined') {
        services = services.map(s => s.id === body.id ? { ...s, disable: body.disable } : s);
    }
    return NextResponse.json({ success: true });
}
