/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Summary: Combines multiple class names into a single string, handling conflicts and conditionals.
 *
 * @param inputs - ClassValue[]. A list of class values (strings, arrays, objects) to combine.
 * @returns string - A merged class name string.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
