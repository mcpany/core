/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * Summary: Merges Tailwind CSS classes.
 *
 * @param inputs - A list of class values (strings, arrays, objects) to combine.
 * @returns A merged class name string.
 *
 * Side Effects:
 * - None.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Recursively traverses an object and attempts to parse string values that look like JSON.
 *
 * Summary: Deeply parses JSON strings within an object.
 *
 * @param obj - The object to traverse and parse.
 * @returns A new object with JSON strings parsed, or the original object if parsing fails.
 *
 * Side Effects:
 * - None.
 */
export function deepParseJson(obj: unknown): unknown {
  if (typeof obj === 'string') {
    const trimmed = obj.trim();
    if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
      try {
        const parsed = JSON.parse(trimmed);
        return deepParseJson(parsed); // Recursively parse the parsed JSON
      } catch (e) {
        return obj; // Return original string if it's not valid JSON
      }
    }
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(deepParseJson);
  }

  if (typeof obj === 'object' && obj !== null) {
    const newObj: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj)) {
      newObj[key] = deepParseJson(value);
    }
    return newObj;
  }

  return obj;
}
