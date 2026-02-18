/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Generates a cURL command to execute a tool.
 * @param toolName The name of the tool.
 * @param args The arguments for the tool.
 * @returns The cURL command string.
 */
export function generateCurlCommand(toolName: string, args: Record<string, unknown>): string {
  const url = typeof window !== 'undefined' ? `${window.location.origin}/api/v1/execute` : 'http://localhost:50050/api/v1/execute';
  const body = JSON.stringify({
    name: toolName,
    arguments: args
  }, null, 2);

  // Escape single quotes for shell safety
  const escapedBody = body.replace(/'/g, "'\\''");

  return `curl -X POST "${url}" \\
  -H "Content-Type: application/json" \\
  -d '${escapedBody}'`;
}

/**
 * Generates a Python script to execute a tool using the requests library.
 * @param toolName The name of the tool.
 * @param args The arguments for the tool.
 * @returns The Python script string.
 */
export function generatePythonCommand(toolName: string, args: Record<string, unknown>): string {
  const url = typeof window !== 'undefined' ? `${window.location.origin}/api/v1/execute` : 'http://localhost:50050/api/v1/execute';
  const argsJson = JSON.stringify(args, null, 4);

  return `import requests
import json

url = "${url}"
payload = {
    "name": "${toolName}",
    "arguments": ${argsJson}
}
headers = {
    "Content-Type": "application/json"
}

response = requests.post(url, json=payload, headers=headers)
print(response.json())`;
}
