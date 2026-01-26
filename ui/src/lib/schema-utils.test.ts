/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { findFileFieldPath, setDeepValue } from './schema-utils';

describe('findFileFieldPath', () => {
  it('should return null for empty schema', () => {
    expect(findFileFieldPath({})).toBeNull();
    expect(findFileFieldPath(null)).toBeNull();
  });

  it('should find root file field', () => {
    const schema = {
      type: "string",
      contentEncoding: "base64"
    };
    expect(findFileFieldPath(schema)).toBe("");
  });

  it('should find file field in object properties', () => {
    const schema = {
      type: "object",
      properties: {
        name: { type: "string" },
        data: { type: "string", contentEncoding: "base64" }
      }
    };
    expect(findFileFieldPath(schema)).toBe("data");
  });

  it('should find nested file field', () => {
    const schema = {
      type: "object",
      properties: {
        meta: {
          type: "object",
          properties: {
            image: {
                type: "object",
                properties: {
                    raw: { type: "string", contentEncoding: "base64" }
                }
            }
          }
        }
      }
    };
    expect(findFileFieldPath(schema)).toBe("meta.image.raw");
  });

  it('should find file field with format: binary', () => {
    const schema = {
        type: "object",
        properties: {
            file: { type: "string", format: "binary" }
        }
    };
    expect(findFileFieldPath(schema)).toBe("file");
  });

  it('should return the first one found', () => {
    const schema = {
        type: "object",
        properties: {
            first: { type: "string", contentEncoding: "base64" },
            second: { type: "string", contentEncoding: "base64" }
        }
    };
    expect(findFileFieldPath(schema)).toBe("first");
  });
});

describe('setDeepValue', () => {
    it('should set simple property', () => {
        const obj = { a: 1 };
        const result = setDeepValue(obj, "b", 2);
        expect(result).toEqual({ a: 1, b: 2 });
    });

    it('should set nested property', () => {
        const obj = { a: { b: 1 } };
        const result = setDeepValue(obj, "a.c", 2);
        expect(result).toEqual({ a: { b: 1, c: 2 } });
    });

    it('should create intermediate objects', () => {
        const obj = {};
        const result = setDeepValue(obj, "a.b.c", 3);
        expect(result).toEqual({ a: { b: { c: 3 } } });
    });
});
