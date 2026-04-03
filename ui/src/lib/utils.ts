/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * cn serves as a public interface for interacting with cn.
 *
 * Summary: Cn the  appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
