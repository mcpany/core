/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { estimateTokens } from './tokens';

describe('estimateTokens', () => {
    it('returns 0 for empty string', () => {
        expect(estimateTokens('')).toBe(0);
        expect(estimateTokens(null as any)).toBe(0);
        expect(estimateTokens(undefined as any)).toBe(0);
    });

    it('estimates based on char count (heuristic 1)', () => {
        // 4 chars per token. "abcdefgh" is 8 chars -> 2 tokens.
        expect(estimateTokens('abcdefgh')).toBe(2);
    });

    it('estimates based on word count (heuristic 2)', () => {
        // 1.3 words per token. "one two three" is 3 words -> ceil(3 * 1.3) = 4.
        // Chars: 13 chars / 4 = 3.25 -> 4.
        // Let's try something where word count dominates.
        // "a b c d e f" -> 6 words * 1.3 = 7.8 -> 8.
        // Chars: 11 chars / 4 = 2.75 -> 3.
        // Should be 8.
        expect(estimateTokens('a b c d e f')).toBe(8);
    });

    it('handles multiple spaces correctly', () => {
        // "a   b" is 2 words.
        // 2 * 1.3 = 2.6 -> 3.
        // Chars: 5 / 4 = 1.25 -> 2.
        // Result 3.
        expect(estimateTokens('a   b')).toBe(3);
    });

    it('handles newlines and tabs', () => {
        // "a\nb\tc" is 3 words.
        // 3 * 1.3 = 3.9 -> 4.
        expect(estimateTokens('a\nb\tc')).toBe(4);
    });

    it('handles leading/trailing whitespace', () => {
        // "  a b  " is 2 words.
        // 2 * 1.3 = 2.6 -> 3.
        expect(estimateTokens('  a b  ')).toBe(3);
    });
});
