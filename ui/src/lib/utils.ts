/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * Summary: Merges Tailwind CSS classes logically, resolving conflicts.
 *
 * Parameters:
 *   - inputs (...ClassValue[]): A list of class values (strings, arrays, objects) to combine.
 *
 * Returns:
 *   - string: A merged class name string.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
