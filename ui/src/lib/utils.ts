/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * Summary: Merges Tailwind CSS classes and handles conditional class logic.
 *
 * @param inputs - ClassValue[]. A list of class values (strings, arrays, objects) to combine.
 * @returns string. A merged class name string.
 *
 * Side Effects:
 *  - None.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
