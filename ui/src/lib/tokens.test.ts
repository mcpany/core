/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { estimateTokens, estimateMessageTokens } from './tokens';

describe('tokens', () => {
  describe('estimateTokens', () => {
    it('returns 0 for empty string', () => {
      expect(estimateTokens('')).toBe(0);
    });

    it('estimates based on characters (short words)', () => {
      // "abc de f" -> 8 chars, 3 words.
      // h1 = 8/4 = 2
      // h2 = 3 * 1.3 = 3.9 -> 4
      // max(2, 4) = 4
      expect(estimateTokens('abc de f')).toBe(4);
    });

    it('estimates based on words (long text)', () => {
        const text = "word ".repeat(10); // 50 chars, 10 words
        // h1 = 50/4 = 12.5 -> 13
        // h2 = 10 * 1.3 = 13
        // max(13, 13) = 13
        expect(estimateTokens(text)).toBe(13);
    });

    it('handles whitespace correctly', () => {
      // The original failing test was "  hello   world  " (17 chars).
      // h1 = ceil(17/4) = 5
      // h2 = ceil(2 * 1.3) = 3
      // result 5.
      expect(estimateTokens('  hello   world  ')).toBe(5);

      // Verify word counting takes precedence when applicable
      // "one two three four" -> 18 chars -> ceil(4.5) -> 5
      // 4 words -> ceil(5.2) -> 6
      expect(estimateTokens('one two three four')).toBe(6);
    });

    it('handles special characters', () => {
      // "hello\tworld\nfoo" -> 15 chars -> ceil(3.75) -> 4
      // 3 words -> ceil(3.9) -> 4
      expect(estimateTokens('hello\tworld\nfoo')).toBe(4);
    });
  });

  describe('estimateMessageTokens', () => {
      it('calculates tokens for message array', () => {
          const messages = [
              { content: 'hello world' }, // 2 words -> 3 tokens
              { content: 'foo bar' }     // 2 words -> 3 tokens
          ];
          expect(estimateMessageTokens(messages)).toBe(6);
      });
  });
});
