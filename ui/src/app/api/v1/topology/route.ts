/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import { Graph, Node, NodeType, NodeStatus } from '@/types/topology';
import { MockDB } from '@/lib/server/mock-db';
import { BuiltInTools } from '@/lib/server/tools';
import { UpstreamServiceConfig } from '@/lib/client';

export async function GET() {
    // Construct the topology graph dynamically from MockDB and other sources

    // 1. Create the Core Node
    const coreNode: Node = {
        id: 'core-server',
        label: 'MCP Gateway',
        type: 'NODE_TYPE_CORE',
        status: 'NODE_STATUS_ACTIVE',
        metrics: {
            qps: Math.random() * 100 + 50,
            latencyMs: Math.random() * 20 + 5,
            errorRate: Math.random() * 0.01
        },
        metadata: {
            version: '1.0.0',
            region: 'us-east-1',
            environment: 'production'
        },
        children: []
    };

    // 2. Add Services as children of Core
    const services = MockDB.services || [];

    // Process services
    services.forEach((service) => {
        const serviceId = service.id || `srv-${Math.random().toString(36).substr(2, 9)}`;
        const isActive = !service.disable;

        const serviceNode: Node = {
            id: serviceId,
            label: service.name,
            type: 'NODE_TYPE_SERVICE',
            status: isActive ? 'NODE_STATUS_ACTIVE' : 'NODE_STATUS_INACTIVE',
            metrics: isActive ? {
                qps: Math.random() * 20,
                latencyMs: Math.random() * 50 + 10,
                errorRate: Math.random() < 0.1 ? Math.random() * 0.05 : 0
            } : undefined,
            metadata: {
                version: service.version || 'unknown',
                address: getServiceAddress(service)
            },
            children: []
        };

        // Add Tools to Service
        // If it's a known mock service, we might want to attach specific tools
        // For now, let's distribute built-in tools or generate mock tools
        if (isActive) {
            // Distribute built-in tools
            const toolKeys = Object.keys(BuiltInTools);
            // Deterministic pseudo-random assignment based on name length
            const assignedTools = toolKeys.filter(k => (k.length + service.name.length) % 3 === 0);

            // If no built-in tools match, add some generic ones
            if (assignedTools.length === 0 && service.name.includes("weather")) {
                 assignedTools.push("weather");
            } else if (assignedTools.length === 0) {
                 // Add mock tools
                 serviceNode.children?.push({
                     id: `${serviceId}-tool-1`,
                     label: `${service.name}-tool`,
                     type: 'NODE_TYPE_TOOL',
                     status: 'NODE_STATUS_ACTIVE',
                     metrics: { qps: Math.random() * 5 },
                     children: []
                 });
            }

            assignedTools.forEach(toolName => {
                 serviceNode.children?.push({
                     id: `${serviceId}-${toolName}`,
                     label: toolName,
                     type: 'NODE_TYPE_TOOL',
                     status: 'NODE_STATUS_ACTIVE',
                     metrics: { qps: Math.random() * 5 },
                     children: []
                 });
            });

            // Add Resources if applicable
            if (service.name.includes("files") || service.name.includes("memory")) {
                serviceNode.children?.push({
                    id: `${serviceId}-res-1`,
                    label: `fs://${service.name}/root`,
                    type: 'NODE_TYPE_RESOURCE',
                    status: 'NODE_STATUS_ACTIVE',
                    children: []
                });
            }
        }

        coreNode.children?.push(serviceNode);
    });

    // 3. Add Clients (Mock)
    const clients: Node[] = [
        {
            id: 'client-web',
            label: 'Web Dashboard',
            type: 'NODE_TYPE_CLIENT',
            status: 'NODE_STATUS_ACTIVE',
            metrics: { qps: Math.random() * 10 },
            metadata: { userAgent: 'Chrome/120.0' }
        },
        {
            id: 'client-cli',
            label: 'MCP CLI',
            type: 'NODE_TYPE_CLIENT',
            status: 'NODE_STATUS_ACTIVE',
            metrics: { qps: Math.random() * 2 },
            metadata: { version: '0.5.2' }
        }
    ];

    // 4. Add Middleware (Mock) - attached to Core in a way?
    // In our tree structure, they might be siblings or special children.
    // Let's add them as children of Core for visualization, though logically they sit between.
    coreNode.children?.push({
        id: 'mw-auth',
        label: 'Auth Middleware',
        type: 'NODE_TYPE_MIDDLEWARE',
        status: 'NODE_STATUS_ACTIVE',
        metrics: { latencyMs: 2 },
        children: []
    });

    coreNode.children?.push({
        id: 'mw-logger',
        label: 'Audit Logger',
        type: 'NODE_TYPE_MIDDLEWARE',
        status: 'NODE_STATUS_ACTIVE',
        children: []
    });

    const graph: Graph = {
        clients: clients,
        core: coreNode
    };

    return NextResponse.json(graph);
}

function getServiceAddress(service: UpstreamServiceConfig): string {
    if (service.http_service) return service.http_service.address;
    if (service.grpc_service) return service.grpc_service.address;
    if (service.command_line_service) return `cmd: ${service.command_line_service.command}`;
    if (service.websocket_service) return service.websocket_service.address;
    if (service.mcp_service) {
        if (service.mcp_service.sse_connection) return service.mcp_service.sse_connection.sse_address;
        if (service.mcp_service.http_connection) return service.mcp_service.http_connection.http_address;
    }
    return 'unknown';
}
