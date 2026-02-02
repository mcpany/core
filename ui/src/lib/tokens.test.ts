import { describe, it, expect } from 'vitest';
import { estimateTokens } from './tokens';

describe('estimateTokens', () => {
    it('should return 0 for empty or null input', () => {
        expect(estimateTokens("")).toBe(0);
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        expect(estimateTokens(null as any)).toBe(0);
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        expect(estimateTokens(undefined as any)).toBe(0);
    });

    it('should estimate based on characters (h1)', () => {
        // "abcde" -> 5 chars. ceil(5/4) = 2.
        // "abcde" -> 1 word. ceil(1*1.3) = 2.
        // max(2, 2) = 2.
        expect(estimateTokens("abcde")).toBe(2);
    });

    it('should estimate based on words (h2)', () => {
        // "a b c d e" -> 9 chars. ceil(9/4) = 3.
        // 5 words. ceil(5*1.3) = 7.
        // max(3, 7) = 7.
        expect(estimateTokens("a b c d e")).toBe(7);
    });

    it('should handle multiple spaces', () => {
        // "a   b" -> 5 chars. ceil(5/4) = 2.
        // 2 words. ceil(2*1.3) = 3.
        // max(2, 3) = 3.
        expect(estimateTokens("a   b")).toBe(3);
    });

    it('should handle newlines and tabs', () => {
        // "a\nb\tc" -> 5 chars. ceil(5/4) = 2.
        // 3 words. ceil(3*1.3) = 4.
        // max(2, 4) = 4.
        expect(estimateTokens("a\nb\tc")).toBe(4);
    });

    it('should trim whitespace', () => {
        // "  a b  " -> 7 chars. ceil(7/4) = 2.
        // 2 words. ceil(2*1.3) = 3.
        // max(2, 3) = 3.
        expect(estimateTokens("  a b  ")).toBe(3);
    });
});
