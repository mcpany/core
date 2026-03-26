/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Combines multiple class names into a single string, handling conflicts and conditionals.
 *
<<<<<<< HEAD
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
=======
 * Summary: Merges Tailwind CSS classes.
 *
 * @param inputs - A list of class values (strings, arrays, objects) to combine.
 * @returns A merged class name string.
 *
 * Side Effects:
 * - None.
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
