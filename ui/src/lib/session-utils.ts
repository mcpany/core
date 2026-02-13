/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface SessionMessage {
  id: string;
  type: "user" | "assistant" | "tool-call" | "tool-result" | "error";
  content?: string;
  toolName?: string;
  toolArgs?: Record<string, unknown>;
  toolResult?: unknown;
  previousResult?: unknown;
  duration?: number;
  timestamp: Date;
}

/**
 * Serializes session messages to a JSON string.
 */
export function serializeSessionMessages(messages: SessionMessage[]): string {
    return JSON.stringify(messages);
}

/**
 * Deserializes session messages from a JSON string, restoring Date objects.
 */
export function deserializeSessionMessages(json: string): SessionMessage[] {
    try {
        return JSON.parse(json, (key, value) => {
            if (key === 'timestamp' && typeof value === 'string') {
                const date = new Date(value);
                // Check if valid date
                if (!isNaN(date.getTime())) {
                    return date;
                }
            }
            return value;
        });
    } catch (e) {
        console.error("Failed to deserialize session messages", e);
        return [];
    }
}
