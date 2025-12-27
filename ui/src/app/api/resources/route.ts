import { NextResponse } from 'next/server';

let resources = [
    { uri: "file:///users/me/docs/notes.txt", name: "My Notes", description: "Personal notes", enabled: true, serviceName: "local-files" },
    { uri: "postgres://db/users", name: "User Table", description: "User database schema", enabled: true, serviceName: "sql-connector" },
];

export async function GET() {
  return NextResponse.json(resources);
}

export async function POST(request: Request) {
    const body = await request.json();
    if (body.uri) {
        resources = resources.map(r => r.uri === body.uri ? { ...r, enabled: body.enabled } : r);
        return NextResponse.json({ message: "Updated" });
    }
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
