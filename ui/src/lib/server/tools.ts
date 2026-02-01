/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Defines the structure of a built-in tool that can be executed by the server.
 */
export interface Tool {
  /** The unique name of the tool. */
  name: string;
  /** A brief description of what the tool does. */
  description: string;
  /** JSON Schema defining the expected arguments for the tool. */
  schema: Record<string, any>;
  /**
   * The function that implements the tool's logic.
   * @param args - The arguments passed to the tool, validated against the schema.
   * @returns A promise resolving to the execution result.
   */
  execute: (args: any) => Promise<any>;
}

/**
 * A registry of built-in tools available in the server.
 * These tools are simulated or provide basic utility without needing an upstream MCP server.
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
      description: "Get current weather for a location (Mock)",
      schema: {
          type: "object",
          properties: {
              location: { type: "string" },
              unit: { type: "string", enum: ["celsius", "fahrenheit"] }
          },
          required: ["location"]
      },
      execute: async ({ location, unit = "celsius" }: { location: string, unit: string }) => {
          // Mock data
          const conditions = ["Sunny", "Cloudy", "Rainy", "Snowy", "Windy"];
          const condition = conditions[Math.floor(Math.random() * conditions.length)];
          const tempBase = Math.floor(Math.random() * 30);
          const temp = unit === "fahrenheit" ? (tempBase * 9/5) + 32 : tempBase;

          return {
              location,
              temperature: temp,
              unit,
              condition,
              humidity: Math.floor(Math.random() * 100) + "%"
          }
      }
  }
};

/**
 * Executes a built-in tool by name.
 *
 * @param toolName - The unique identifier of the tool to execute.
 * @param args - The arguments object to pass to the tool.
 * @returns A promise that resolves to the result of the tool execution.
 * @throws {Error} If the tool with the specified name is not found or execution fails.
 */
export async function executeTool(toolName: string, args: any) {
  const tool = BuiltInTools[toolName];
  if (!tool) {
    throw new Error(`Tool '${toolName}' not found`);
  }
  return tool.execute(args);
}
