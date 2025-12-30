/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

let tools = [
    {
        name: "get_weather",
        description: "Get weather for a location",
        serviceName: "weather-service",
        enabled: true,
        schema: {
            type: "object",
            properties: {
                location: { type: "string", description: "City name or coordinates" },
                unit: { type: "string", enum: ["celsius", "fahrenheit"], description: "Temperature unit" }
            },
            required: ["location"]
        }
    },
    {
        name: "list_files",
        description: "List files in directory",
        serviceName: "local-files",
        enabled: true,
        schema: {
            type: "object",
            properties: {
                path: { type: "string", description: "Path to directory" },
                recursive: { type: "boolean", description: "List recursively" }
            },
            required: ["path"]
        }
    },
    {
        name: "read_file",
        description: "Read file content",
        serviceName: "local-files",
        enabled: false,
        schema: {
            type: "object",
            properties: {
                path: { type: "string", description: "Path to file" }
            },
            required: ["path"]
        }
    },
    {
        name: "save_memory",
        description: "Save a key-value pair",
        serviceName: "memory-store",
        enabled: true,
        schema: {
            type: "object",
            properties: {
                key: { type: "string", description: "Memory key" },
                value: { type: "string", description: "Memory value" },
                tags: { type: "array", items: { type: "string" }, description: "Tags for the memory" }
            },
            required: ["key", "value"]
        }
    },
];

export async function GET() {
  return NextResponse.json({ tools });
}

export async function POST(request: Request) {
    let body;
    try {
        body = await request.json();
    } catch (e) {
        return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
    }

    // Execute tool (mock)
    if (body.action === 'execute') {
        const toolName = body.name;
        const args = body.arguments;

        // Simulate processing delay
        await new Promise(resolve => setTimeout(resolve, 800));

        if (toolName === 'get_weather') {
            return NextResponse.json({
                result: {
                    temperature: 22,
                    condition: "Sunny",
                    location: args.location,
                    unit: args.unit || "celsius"
                }
            });
        }
        if (toolName === 'list_files') {
             return NextResponse.json({
                result: {
                    files: ["file1.txt", "file2.jpg", "dir/"]
                }
            });
        }
        if (toolName === 'save_memory') {
            return NextResponse.json({
                result: {
                    status: "stored",
                    id: Math.random().toString(36).substring(7)
                }
            });
        }

        return NextResponse.json({ result: "Tool execution successful (mock)" });
    }

    // Update status
    if (body.name && typeof body.enabled === 'boolean') {
        tools = tools.map(t => t.name === body.name ? { ...t, enabled: body.enabled } : t);
        return NextResponse.json({ message: "Updated" });
    }

    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
}
