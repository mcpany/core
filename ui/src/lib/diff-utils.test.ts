/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { computeDiff } from './diff-utils';

describe('computeDiff', () => {
    it('should detect added lines', () => {
        const oldText = 'line 1\nline 2';
        const newText = 'line 1\nline 2\nline 3';
        const changes = computeDiff(oldText, newText);

        expect(changes).toBeDefined();
        // Expect at least one change to be "added"
        const added = changes.find(c => c.added);
        expect(added).toBeDefined();
        expect(added?.value).toContain('line 3');
    });

    it('should detect removed lines', () => {
        const oldText = 'line 1\nline 2';
        const newText = 'line 1';
        const changes = computeDiff(oldText, newText);

        const removed = changes.find(c => c.removed);
        expect(removed).toBeDefined();
        expect(removed?.value).toContain('line 2');
    });

    it('should handle JSON strings', () => {
        const oldJson = JSON.stringify({ a: 1, b: 2 }, null, 2);
        const newJson = JSON.stringify({ a: 1, b: 3 }, null, 2);
        const changes = computeDiff(oldJson, newJson);

        const removed = changes.find(c => c.removed);
        const added = changes.find(c => c.added);

        expect(removed?.value).toContain('"b": 2');
        expect(added?.value).toContain('"b": 3');
    });

    it('should return one change with no added/removed for identical strings', () => {
         const text = 'identical';
         const changes = computeDiff(text, text);
         expect(changes.length).toBe(1);
         expect(changes[0].added).toBeFalsy();
         expect(changes[0].removed).toBeFalsy();
    });
});
