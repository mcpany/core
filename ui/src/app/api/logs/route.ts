/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest } from "next/server";

export const runtime = "nodejs";

// Mock data sources
const SOURCES = ["gateway", "auth-service", "db-worker", "api-server", "scheduler", "payment-processor", "llm-client"];
const MESSAGES = {
  INFO: [
    "Server started on port 8080",
    "Request received: GET /api/v1/tools",
    "Health check passed",
    "Configuration reloaded",
    "User authenticated successfully",
    "Backup completed",
    "Job scheduled: daily-cleanup",
    "Cache refreshed",
  ],
  WARN: [
    "Response time > 500ms",
    "Rate limit approaching for user",
    "Deprecated API usage detected",
    "Memory usage at 75%",
    "Retrying connection to database...",
    "Disk space low on volume /mnt/data",
  ],
  ERROR: [
    "Database connection timeout",
    "Failed to parse JSON body",
    "Upstream service unavailable",
    "Permission denied: /etc/hosts",
    "Payment gateway rejected transaction",
    "Critical system failure in module X",
  ],
  DEBUG: [
    "Payload size: 1024 bytes",
    "Executing query: SELECT * FROM users",
    "Cache miss for key: user_123",
    "Context switch",
    "Parsing configuration file...",
    "Variable x = 42",
    "Thread-12 acquired lock",
  ],
};

export async function GET(req: NextRequest) {
  const encoder = new TextEncoder();
  const readable = new ReadableStream({
    async start(controller) {
      const sendEvent = (data: any) => {
        controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`));
      };

      // Send initial connection message
      sendEvent({
        id: "sys-1",
        timestamp: new Date().toISOString(),
        level: "INFO",
        source: "system",
        message: "Log stream connected successfully.",
      });

      const intervalId = setInterval(() => {
        const levels = Object.keys(MESSAGES) as Array<keyof typeof MESSAGES>;
        // Bias towards INFO and DEBUG
        const weightedLevels = [...levels, "INFO", "INFO", "DEBUG", "DEBUG"];
        const level = weightedLevels[Math.floor(Math.random() * weightedLevels.length)];
        const messages = MESSAGES[level];
        const message = messages[Math.floor(Math.random() * messages.length)];
        const source = SOURCES[Math.floor(Math.random() * SOURCES.length)];

        const logEntry = {
          id: Math.random().toString(36).substring(7),
          timestamp: new Date().toISOString(),
          level,
          source,
          message,
        };

        sendEvent(logEntry);
      }, 800); // Send a log every 800ms

      // Close the stream when the client disconnects
      req.signal.addEventListener("abort", () => {
        clearInterval(intervalId);
        controller.close();
      });
    },
  });

  return new Response(readable, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    },
  });
}
