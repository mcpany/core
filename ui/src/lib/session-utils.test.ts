/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { serializeSessionMessages, deserializeSessionMessages, SessionMessage } from "./session-utils";
import { describe, it, expect } from "vitest";

describe("session-utils", () => {
    it("should serialize and deserialize messages including Date objects", () => {
        const timestamp = new Date("2023-01-01T12:00:00Z");
        const messages: SessionMessage[] = [
            {
                id: "1",
                type: "user",
                content: "test",
                timestamp: timestamp
            }
        ];

        const serialized = serializeSessionMessages(messages);
        const deserialized = deserializeSessionMessages(serialized);

        expect(deserialized).toHaveLength(1);
        expect(deserialized[0].content).toBe("test");
        expect(deserialized[0].timestamp).toBeInstanceOf(Date);
        expect(deserialized[0].timestamp.getTime()).toBe(timestamp.getTime());
    });

    it("should handle invalid JSON gracefully", () => {
        // We expect an empty array or handled error, not a crash
        const consoleSpy = { error: () => {} }; // Mock console if needed, but vitest captures it
        expect(deserializeSessionMessages("invalid-json")).toEqual([]);
    });
});
