/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect } from 'vitest';
import { estimateTokens, estimateMessageTokens } from './tokens';

describe('tokens', () => {
  describe('estimateTokens', () => {
    it('should return 0 for empty string', () => {
      expect(estimateTokens('')).toBe(0);
    });

    it('should estimate based on characters (short words)', () => {
      // "a b c d" -> 7 chars / 4 = 1.75 -> 2
      // words: 4 * 1.3 = 5.2 -> 6
      // max(2, 6) = 6
      expect(estimateTokens('a b c d')).toBe(6);
    });

    it('should estimate based on words (standard sentence)', () => {
      const text = "The quick brown fox jumps over the lazy dog";
      // chars: 43 / 4 = 10.75 -> 11
      // words: 9 * 1.3 = 11.7 -> 12
      // max(11, 12) = 12
      expect(estimateTokens(text)).toBe(12);
    });

    it('should handle multiple spaces and newlines', () => {
      const text = "Word1   Word2\nWord3";
      // chars: 19 / 4 = 4.75 -> 5
      // words: 3 * 1.3 = 3.9 -> 4
      // max(5, 4) = 5
      expect(estimateTokens(text)).toBe(5);
    });

    it('should handle large text roughly correctly', () => {
        const text = "a".repeat(100);
        // chars: 100/4 = 25
        // words: 1 * 1.3 = 2
        // max = 25
        expect(estimateTokens(text)).toBe(25);
    });
  });

  describe('estimateMessageTokens', () => {
    it('should sum up tokens from messages', () => {
      const messages = [
        { content: "hello world" }, // chars: 11/4=3, words: 2*1.3=3 -> 3
        { content: "foo bar" }      // chars: 7/4=2, words: 2*1.3=3 -> 3
      ];
      // total = 6
      expect(estimateMessageTokens(messages)).toBe(6);
    });

    it('should handle complex message objects', () => {
        const msg = {
            content: "test",
            toolName: "my_tool",
            toolArgs: { arg: "val" }
        };
        // content string: "test my_tool {\"arg\":\"val\"}"
        // length: 4 + 1 + 7 + 1 + 13 = 26
        // chars: 26/4 = 7
        // words: "test", "my_tool", "{"arg":"val"}" -> 3 words?
        // split on space: "test", "my_tool", "{\"arg\":\"val\"}" -> 3 words
        // 3 * 1.3 = 4
        // max(7, 4) = 7
        const tokens = estimateMessageTokens([msg]);
        expect(tokens).toBeGreaterThan(0);
    });
  });
});
