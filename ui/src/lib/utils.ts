/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string.
 *
 * This function uses `clsx` to handle conditional class names and `tailwind-merge`
 * to resolve conflicting Tailwind CSS classes.
 *
 * @param inputs - A list of class values (strings, arrays, objects) to combine.
 * @returns A merged class name string.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
