/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Conditionally merges Tailwind CSS class names.
 *
 * This utility wraps `clsx` for conditional logic and `tailwind-merge` for resolving style conflicts
 * (e.g., `p-4` vs `p-2`). It is the standard way to construct `className` strings in the application.
 *
 * @param {...ClassValue[]} inputs - A variable list of class values, which can be strings, objects, or arrays.
 * @returns {string} The final, deduped, and merged class name string.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
