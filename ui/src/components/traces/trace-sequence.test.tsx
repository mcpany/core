/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { calculateLayout } from "./trace-sequence";
import { Trace } from "@/types/trace";

describe("TraceSequence Layout", () => {
    it("should correctly identify participants and generate events for a simple trace", () => {
        const trace: Trace = {
            id: "test-trace",
            timestamp: new Date().toISOString(),
            totalDuration: 100,
            status: "success",
            trigger: "user",
            rootSpan: {
                id: "span-core",
                name: "Core Operation",
                type: "core",
                startTime: 0,
                endTime: 100,
                status: "success",
                children: [
                    {
                        id: "span-service",
                        name: "Service Call",
                        type: "service",
                        serviceName: "weather-service",
                        startTime: 10,
                        endTime: 90,
                        status: "success",
                        children: [
                            {
                                id: "span-tool",
                                name: "get_weather",
                                type: "tool",
                                startTime: 20,
                                endTime: 80,
                                status: "success",
                                children: []
                            }
                        ]
                    }
                ]
            }
        };

        const layout = calculateLayout(trace);

        // Check Participants
        // Expected: Client, Core, Service, Tool
        expect(layout.participants).toHaveLength(4);
        expect(layout.participants.map(p => p.id)).toEqual([
            "client",
            "core",
            "service:weather-service",
            "tool:get_weather"
        ]);

        // Check Events
        // Call Client->Core
        // Call Core->Service
        // Call Service->Tool
        // Return Tool->Service
        // Return Service->Core
        // Return Core->Client
        expect(layout.events).toHaveLength(6);

        const calls = layout.events.filter(e => e.type === "call");
        const returns = layout.events.filter(e => e.type === "return");

        expect(calls).toHaveLength(3);
        expect(returns).toHaveLength(3);

        // Verify ordering (Calls should be before returns in this simple stack)
        // But layout events are chronologically sorted by generation order (DFS),
        // which maps to:
        // Call Core (by Client)
        // Call Service (by Core)
        // Call Tool (by Service)
        // Return Tool (to Service)
        // Return Service (to Core)
        // Return Core (to Client)

        expect(layout.events[0].id).toBe("call-span-core");
        expect(layout.events[1].id).toBe("call-span-service");
        expect(layout.events[2].id).toBe("call-span-tool");
        expect(layout.events[3].id).toBe("return-span-tool");
        expect(layout.events[4].id).toBe("return-span-service");
        expect(layout.events[5].id).toBe("return-span-core");
    });

    it("should handle sibling spans correctly", () => {
        const trace: Trace = {
            id: "test-trace-siblings",
            timestamp: new Date().toISOString(),
            totalDuration: 100,
            status: "success",
            trigger: "user",
            rootSpan: {
                id: "span-core",
                name: "Core",
                type: "core",
                startTime: 0,
                endTime: 100,
                status: "success",
                children: [
                    {
                        id: "span-1",
                        name: "First",
                        type: "tool",
                        startTime: 10,
                        endTime: 20,
                        status: "success"
                    },
                    {
                        id: "span-2",
                        name: "Second",
                        type: "tool",
                        startTime: 30,
                        endTime: 40,
                        status: "success"
                    }
                ]
            }
        };

        const layout = calculateLayout(trace);

        // Participants: Client, Core, Tool:First, Tool:Second
        expect(layout.participants).toHaveLength(4);

        // Events:
        // Call Core
        // Call First
        // Return First
        // Call Second
        // Return Second
        // Return Core
        expect(layout.events).toHaveLength(6);

        expect(layout.events[1].label).toBe("First");
        expect(layout.events[3].label).toBe("Second");
    });
});
