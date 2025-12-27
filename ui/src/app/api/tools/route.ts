import { NextResponse } from 'next/server';

let tools = [
    { name: "get_weather", description: "Get current weather for a location", enabled: true, serviceName: "weather-service" },
    { name: "read_file", description: "Read a file from the filesystem", enabled: true, serviceName: "local-files" },
    { name: "write_file", description: "Write data to a file", enabled: false, serviceName: "local-files" },
];

export async function GET() {
  return NextResponse.json(tools);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.name) {
        tools = tools.map(t => t.name === body.name ? { ...t, enabled: body.enabled } : t);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
