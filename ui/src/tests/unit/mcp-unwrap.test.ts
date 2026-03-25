/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { deepParseJson } from '@/lib/mcp-unwrap';
import { describe, it, expect } from 'vitest';

describe('deepParseJson', () => {
    it('parses a simple JSON string into an object', () => {
        const input = '{"value": "Version 1"}';
        const expected = { value: 'Version 1' };
        expect(deepParseJson(input)).toEqual(expected);
    });

    it('returns original string if not valid JSON', () => {
        const input = 'Not valid JSON';
        expect(deepParseJson(input)).toEqual(input);
    });

    it('recursively parses nested JSON strings inside an object', () => {
        const input = {
            content: [
                {
                    type: 'text',
                    text: '{"value": "Version 1"}'
                }
            ],
            isError: false,
            nested: '{"inner": {"data": 42}}'
        };
        const expected = {
            content: [
                {
                    type: 'text',
                    text: { value: 'Version 1' }
                }
            ],
            isError: false,
            nested: { inner: { data: 42 } }
        };
        expect(deepParseJson(input)).toEqual(expected);
    });

    it('recursively parses deeply nested JSON strings', () => {
        // Stringified twice
        const deeplyNestedString = JSON.stringify({
            level1: JSON.stringify({ level2: 'value' })
        });
        const input = { data: deeplyNestedString };
        const expected = { data: { level1: { level2: 'value' } } };
        expect(deepParseJson(input)).toEqual(expected);
    });

    it('handles primitive JSON strings that parse to primitives', () => {
        // We only want to convert if the parsed result is an object/array.
        // If someone sends '"just a string"' as a JSON string, it parses to 'just a string'.
        // Our deepParseJson should just return the original string if parsing doesn't yield an object.
        const input = '"just a string"';
        // Note: deepParseJson currently checks `typeof parsed === 'object' && parsed !== null`.
        // So a string primitive will fall through to `return obj`.
        expect(deepParseJson(input)).toEqual(input);
    });

    it('handles arrays properly', () => {
        const input = ['{"a":1}', '{"b":2}'];
        const expected = [{ a: 1 }, { b: 2 }];
        expect(deepParseJson(input)).toEqual(expected);
    });

    it('handles null and undefined', () => {
        expect(deepParseJson(null)).toBeNull();
        expect(deepParseJson(undefined)).toBeUndefined();
    });
});
