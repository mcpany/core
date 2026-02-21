/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

interface CodeGeneratorOptions {
  toolName: string;
  args: Record<string, unknown>;
  baseUrl?: string;
  token?: string;
}

/**
 * Generates a Curl command for executing a tool.
 * @param params - The parameters for generation.
 * @param params.toolName - The name of the tool.
 * @param params.args - The arguments for the tool.
 * @param params.baseUrl - The base URL of the API.
 * @param params.token - The authentication token.
 * @returns The generated Curl command string.
 */
export function generateCurlCommand({ toolName, args, baseUrl = "http://localhost:8080", token }: CodeGeneratorOptions): string {
  const payload = {
    name: toolName,
    arguments: args,
  };

  let command = `curl -X POST ${baseUrl}/api/v1/execute \\\n`;
  command += `  -H "Content-Type: application/json" \\\n`;

  if (token) {
    command += `  -H "Authorization: Basic ${token}" \\\n`;
  }

  command += `  -d '${JSON.stringify(payload, null, 2)}'`;

  return command;
}

/**
 * Generates Python code for executing a tool using the requests library.
 * @param params - The parameters for generation.
 * @param params.toolName - The name of the tool.
 * @param params.args - The arguments for the tool.
 * @param params.baseUrl - The base URL of the API.
 * @param params.token - The authentication token.
 * @returns The generated Python code string.
 */
export function generatePythonCode({ toolName, args, baseUrl = "http://localhost:8080", token }: CodeGeneratorOptions): string {
  const payload = {
    name: toolName,
    arguments: args,
  };

  // Python requests code generation
  // Retry CI
  let code = `import requests\nimport json\n\n`;
  code += `url = "${baseUrl}/api/v1/execute"\n`;
  code += `payload = ${JSON.stringify(payload, null, 4)}\n`;
  code += `headers = {\n    "Content-Type": "application/json"\n`;

  if (token) {
    code += `    "Authorization": "Basic ${token}"\n`;
  }

  code += `}\n\n`;
  code += `response = requests.post(url, json=payload, headers=headers)\n`;
  code += `print(json.dumps(response.json(), indent=2))`;

  return code;
}
