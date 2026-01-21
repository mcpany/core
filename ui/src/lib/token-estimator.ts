/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Estimates the number of tokens a given object or string would consume.
 * This is a heuristic based on the common rule of thumb: 1 token ~= 4 characters.
 *
 * @param data - The data to estimate (string or object).
 * @returns The estimated number of tokens.
 */
export function estimateTokens(data: unknown): number {
  if (data === null || data === undefined) {
    return 0;
  }

  let text = "";
  if (typeof data === "string") {
    text = data;
  } else {
    try {
      text = JSON.stringify(data);
    } catch (_e) {
      // Fallback for circular references or other stringify errors
      return 0;
    }
  }

  // Common approximation: 1 token is roughly 4 characters
  return Math.ceil(text.length / 4);
}
