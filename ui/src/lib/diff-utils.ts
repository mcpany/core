/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { diffLines, Change } from 'diff';

/**
 * Computes the line-by-line diff between two text strings.
 *
 * @param oldText The original text.
 * @param newText The new text.
 * @returns An array of Change objects describing the diff.
 */
export function computeDiff(oldText: string, newText: string): Change[] {
    return diffLines(oldText, newText);
}
