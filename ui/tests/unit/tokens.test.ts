/*
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { estimateTokens } from '../../src/lib/tokens';

describe('estimateTokens', () => {
  it('returns 0 for empty string', () => {
    expect(estimateTokens('')).toBe(0);
  });

  it('returns appropriate token count for whitespace', () => {
    // Current implementation returns 2 for "   " because:
    // charCount = 3 -> h1 = 1
    // wordCount = "".split() -> [""] -> 1 -> h2 = 2
    // max(1, 2) = 2
    expect(estimateTokens('   ')).toBe(2);
  });

  it('estimates based on characters and words', () => {
    const text = 'hello world';
    // chars: 11. h1 = ceil(11/4) = 3.
    // words: 2. h2 = ceil(2 * 1.3) = 3.
    // result: 3.
    expect(estimateTokens(text)).toBe(3);
  });

  it('handles multiple spaces', () => {
    const text = 'hello   world';
    // chars: 13. h1 = ceil(13/4) = 4.
    // words: 2. h2 = 3.
    // result: 4.
    expect(estimateTokens(text)).toBe(4);
  });

  it('handles newlines', () => {
    const text = 'hello\nworld';
    // chars: 11. h1 = 3.
    // words: 2. h2 = 3.
    // result: 3.
    expect(estimateTokens(text)).toBe(3);
  });

  it('handles mixed content', () => {
    const text = 'a b c';
    // chars: 5. h1 = 2.
    // words: 3. h2 = ceil(3*1.3) = 4.
    // result: 4.
    expect(estimateTokens(text)).toBe(4);
  });
});
