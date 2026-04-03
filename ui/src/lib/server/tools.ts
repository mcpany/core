/**
 * Summary: Document Tool
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Tool represents the public Tool entity.
 *
 * Summary: Provides tool execution capabilities for .
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export interface Tool {
  name: string;
  description: string;
  schema: Record<string, any>;
  execute: (args: any) => Promise<any>;
}

/**
 * BuiltInTools serves as a public interface for interacting with BuiltInTools.
 *
 * Summary: Defines the structured data model representing a in tools.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
 * executeTool serves as a public interface for interacting with executeTool.
 *
 * Summary: Execute the tool appropriately based on current system conditions.
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
export async function executeTool(toolName: string, args: any) {
  const tool = BuiltInTools[toolName];
  if (!tool) {
    throw new Error(`Tool '${toolName}' not found`);
  }
  return tool.execute(args);
}
