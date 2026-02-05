/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * @remarks
 * This function serves as a utility for conditional class merging, typically used with Tailwind CSS.
 * It uses `clsx` to construct the class string and `tailwind-merge` to resolve conflicting classes.
 *
 * @param inputs - A list of class values (strings, arrays, objects) to combine.
 * @returns A merged class name string.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
