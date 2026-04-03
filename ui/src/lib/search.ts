/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Calculates a simple similarity score between a query and a target string.
 * Returns a score between 0.0 (no match) and 1.0 (exact match).
 *
 * @param query The search query
 * @param target The string to search against
 * @returns A score from 0.0 to 1.0
 */
export function scoreMatch(query: string, target: string): number {
    if (!query || !target) return 0;

    const q = query.toLowerCase().trim();
    const t = target.toLowerCase().trim();

    if (q === t) return 1.0;

    if (t.includes(q)) {
        // If it's a substring, score is based on how much of the target it covers
        // Add a slight boost (0.1) so pure substrings are scored higher
        const coverage = q.length / t.length;
        return Math.min(0.8 + (coverage * 0.19), 0.99);
    }

    // Very basic fuzzy match: check if all characters in query exist in target in order
    let qIdx = 0;
    let matchCount = 0;
    for (let i = 0; i < t.length && qIdx < q.length; i++) {
        if (t[i] === q[qIdx]) {
            qIdx++;
            matchCount++;
        }
    }

    if (qIdx === q.length) {
        // All chars found in order
        return Math.min(0.5 + (matchCount / t.length * 0.3), 0.79);
    }

    return 0;
}
