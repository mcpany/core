
import { NextResponse } from 'next/server';

// Mock database
let services = [
  {
    id: "svc_01",
    name: "Payment Gateway",
    connection_pool: { max_connections: 100 },
    disable: false,
    version: "v1.2.0",
    service_config: {
        http_service: {
            address: "https://api.stripe.com"
        }
    }
  },
  {
    id: "svc_02",
    name: "User Service",
    connection_pool: { max_connections: 50 },
    disable: false,
    version: "v2.1.0",
    service_config: {
        grpc_service: {
            address: "localhost:50051"
        }
    }
  },
  {
      id: "svc_03",
      name: "Legacy Auth",
      connection_pool: { max_connections: 10 },
      disable: true,
      version: "v0.5.0",
      service_config: {
          http_service: {
              address: "http://legacy-auth:8080"
          }
      }
  }
];

export async function GET() {
  return NextResponse.json(services);
}

export async function POST(request: Request) {
  const body = await request.json();
  // Simulate update
  services = services.map(s => s.id === body.id ? { ...s, ...body } : s);
  return NextResponse.json({ success: true, service: body });
}
