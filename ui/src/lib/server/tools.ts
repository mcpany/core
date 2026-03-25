/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Tool {
  name: string;
  description: string;
  schema: Record<string, any>;
  execute: (args: any) => Promise<any>;
}

/**
 * BuiltInTools contains the definitions and implementations of standard tools
 * provided by the server, such as calculator, echo, and system info.
 */
export const BuiltInTools: Record<string, Tool> = {
  calculator: {
    name: "calculator",
    description: "Performs basic arithmetic operations (add, subtract, multiply, divide)",
    schema: {
      type: "object",
      properties: {
        operation: { type: "string", enum: ["add", "subtract", "multiply", "divide"] },
        a: { type: "number" },
        b: { type: "number" },
      },
      required: ["operation", "a", "b"],
    },
    execute: async ({ operation, a, b }: { operation: string; a: number; b: number }) => {
      switch (operation) {
        case "add": return a + b;
        case "subtract": return a - b;
        case "multiply": return a * b;
        case "divide":
          if (b === 0) throw new Error("Division by zero");
          return a / b;
        default: throw new Error(`Unknown operation: ${operation}`);
      }
    },
  },
  echo: {
    name: "echo",
    description: "Echoes back the input message",
    schema: {
      type: "object",
      properties: {
        message: { type: "string" },
      },
      required: ["message"],
    },
    execute: async ({ message }: { message: string }) => {
      return { message: `Echo: ${message}`, receivedAt: new Date().toISOString() };
    },
  },
  system_info: {
    name: "system_info",
    description: "Returns basic system information (simulated)",
    schema: {
      type: "object",
      properties: {},
    },
    execute: async () => {
      return {
        platform: process.platform,
        nodeVersion: process.version,
        uptime: process.uptime(),
        memoryUsage: process.memoryUsage(),
      };
    },
  },
  weather: {
      name: "weather",
      description: "Get current weather for a location",
      schema: {
          type: "object",
          properties: {
              location: { type: "string" },
              unit: { type: "string", enum: ["celsius", "fahrenheit"] }
          },
          required: ["location"]
      },
      execute: async ({ location, unit = "celsius" }: { location: string, unit: string }) => {
          try {
              const res = await fetch(`https://wttr.in/${encodeURIComponent(location)}?format=j1`);
              if (!res.ok) {
                  throw new Error(`Failed to fetch weather data: ${res.statusText}`);
              }
              const data = await res.json();
              const current = data.current_condition[0];
              const tempBaseC = parseInt(current.temp_C, 10);
              const tempBaseF = parseInt(current.temp_F, 10);
              const temp = unit === "fahrenheit" ? tempBaseF : tempBaseC;

              return {
                  location,
                  temperature: temp,
                  unit,
                  condition: current.weatherDesc[0].value,
                  humidity: current.humidity + "%"
              };
          } catch (error) {
              return {
                  location,
                  error: "Could not fetch real weather data. Please try again later."
              };
          }
      }
  }
};

/**
 * executeTool executes a built-in tool by name with the provided arguments.
 *
 * @param toolName - The name of the tool to execute.
 * @param args - The arguments for the tool execution.
 * @returns The result of the tool execution.
 * @throws Error if the tool is not found or execution fails.
 */
export async function executeTool(toolName: string, args: any) {
  const tool = BuiltInTools[toolName];
  if (!tool) {
    throw new Error(`Tool '${toolName}' not found`);
  }
  return tool.execute(args);
}
