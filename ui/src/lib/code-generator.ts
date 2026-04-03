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
 * generateCurlCommand serves as a public interface for interacting with generateCurlCommand.
 *
 * Summary: Generate the curl command appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
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
 * generatePythonCode serves as a public interface for interacting with generatePythonCode.
 *
 * Summary: Generate the python code appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
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
