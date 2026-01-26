/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Span } from "@/types/trace";

/**
 * Generates a curl command from a trace span.
 *
 * @param span - The span to generate the curl command from.
 * @returns The generated curl command string.
 */
export function generateCurlCommand(span: Span): string {
  if (span.type === 'service') {
    // Attempt to extract HTTP details from input
    // Assuming input has url, method, headers, body
    const input = span.input || {};
    const url = input.url;
    // If no URL is present, we can't generate a valid curl for a generic service call
    // unless we know the service schema.
    if (!url) {
      return "# Unable to generate curl: Missing URL in service span input";
    }

    const method = (input.method || 'GET').toUpperCase();
    const headers = input.headers || {};
    const body = input.body;

    let cmd = `curl -X ${method} "${url}"`;

    // Add headers
    for (const [key, value] of Object.entries(headers)) {
      // Escape double quotes in header values
      const safeValue = String(value).replace(/"/g, '\\"');
      cmd += ` \\\n  -H "${key}: ${safeValue}"`;
    }

    // Add body
    if (body) {
        let bodyStr = body;
        if (typeof body !== 'string') {
            try {
                bodyStr = JSON.stringify(body);
            } catch {
                bodyStr = String(body);
            }
        }

        // Escape single quotes for shell since we wrap body in single quotes
        const escapedBody = bodyStr.replace(/'/g, "'\\''");
        cmd += ` \\\n  -d '${escapedBody}'`;
    }

    return cmd;
  } else if (span.type === 'tool') {
      // Generate MCP tool call
      const toolName = span.name;
      const args = span.input || {};

      // Default endpoint. Use window.location.origin if available (browser),
      // fallback to localhost:50050 (server/test)
      let endpoint = "http://localhost:50050/mcp";
      if (typeof window !== 'undefined') {
          endpoint = `${window.location.origin}/mcp`;
      }

      const payload = {
          jsonrpc: "2.0",
          method: "tools/call",
          params: {
              name: toolName,
              arguments: args
          },
          id: 1 // Dummy ID
      };

      const bodyStr = JSON.stringify(payload);
      const escapedBody = bodyStr.replace(/'/g, "'\\''");

      return `curl -X POST ${endpoint} \\\n  -H "Content-Type: application/json" \\\n  -d '${escapedBody}'`;
  }

  return "# Copy as Curl is not supported for this span type";
}
