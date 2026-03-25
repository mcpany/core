/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { executeTool } from '@/lib/server/tools';

describe('Tool Executor', () => {
    it('should execute calculator tool correctly', async () => {
        const result = await executeTool('calculator', { operation: 'add', a: 5, b: 3 });
        expect(result).toBe(8);
    });

    it('should throw error for unknown tool', async () => {
        await expect(executeTool('unknown', {})).rejects.toThrow("Tool 'unknown' not found");
    });

    it('should handle division by zero', async () => {
        await expect(executeTool('calculator', { operation: 'divide', a: 10, b: 0 })).rejects.toThrow("Division by zero");
    });

    it('should execute echo tool', async () => {
        const result = await executeTool('echo', { message: 'hello' });
        expect(result.message).toBe('Echo: hello');
    });
});
