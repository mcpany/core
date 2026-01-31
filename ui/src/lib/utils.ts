/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * It uses `clsx` for conditional logic (objects, arrays, truthy values) and `tailwind-merge`
 * to resolve conflicting Tailwind CSS classes (e.g., `p-4` vs `p-2`).
 *
 * @param inputs - A list of class values (strings, arrays of strings, or objects with boolean values).
 * @returns A merged, deduplicated class name string suitable for the `className` prop.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
