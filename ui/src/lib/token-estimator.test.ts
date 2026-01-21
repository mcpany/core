/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { estimateTokens } from './token-estimator';

describe('estimateTokens', () => {
  it('should return 0 for null or undefined', () => {
    expect(estimateTokens(null)).toBe(0);
    expect(estimateTokens(undefined)).toBe(0);
  });

  it('should estimate tokens for a string correctly', () => {
    const text = "12345678"; // 8 chars -> 2 tokens
    expect(estimateTokens(text)).toBe(2);
  });

  it('should round up for partial tokens', () => {
    const text = "12345"; // 5 chars -> 2 tokens
    expect(estimateTokens(text)).toBe(2);
  });

  it('should estimate tokens for an object', () => {
    const obj = { key: "value" };
    // JSON.stringify(obj) -> '{"key":"value"}' (15 chars) -> 4 tokens
    expect(estimateTokens(obj)).toBe(4);
  });

  it('should handle empty string', () => {
    expect(estimateTokens("")).toBe(0);
  });

  it('should handle empty object', () => {
    expect(estimateTokens({})).toBe(1); // "{}" is 2 chars -> 1 token
  });
});
