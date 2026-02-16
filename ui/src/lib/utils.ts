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
 * Utilities `clsx` for conditional classes and `tailwind-merge` to handle Tailwind CSS conflicts.
 * This is the standard utility for Shadcn UI components.
 *
 * @param inputs - ClassValue[]. A list of class values (strings, arrays, objects) to combine.
 *
 * @returns string - A merged, deduplicated class name string.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
