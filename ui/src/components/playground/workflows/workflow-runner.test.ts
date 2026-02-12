/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { runWorkflow } from './workflow-runner';
import { Workflow } from '@/types/workflow';
import { vi, describe, it, expect } from 'vitest';

describe('WorkflowRunner', () => {
    it('should execute steps sequentially', async () => {
        const mockExecutor = vi.fn();
        mockExecutor.mockResolvedValueOnce({ output: 'step1' });
        mockExecutor.mockResolvedValueOnce({ output: 'step2' });

        const workflow: Workflow = {
            id: '1',
            name: 'Test Workflow',
            steps: [
                { id: 's1', name: 'Step 1', toolName: 'tool1', arguments: {} },
                { id: 's2', name: 'Step 2', toolName: 'tool2', arguments: {} }
            ],
            createdAt: '',
            updatedAt: ''
        };

        const onStepUpdate = vi.fn();

        await runWorkflow(workflow, onStepUpdate, mockExecutor);

        expect(mockExecutor).toHaveBeenCalledTimes(2);
        expect(mockExecutor).toHaveBeenNthCalledWith(1, { name: 'tool1', arguments: {} });
        expect(mockExecutor).toHaveBeenNthCalledWith(2, { name: 'tool2', arguments: {} });

        // Check status updates
        // Step 1 running
        expect(onStepUpdate).toHaveBeenCalledWith('s1', expect.objectContaining({ status: 'running' }));
        // Step 1 success
        expect(onStepUpdate).toHaveBeenCalledWith('s1', expect.objectContaining({ status: 'success', result: { output: 'step1' } }));
        // Step 2 running
        expect(onStepUpdate).toHaveBeenCalledWith('s2', expect.objectContaining({ status: 'running' }));
        // Step 2 success
        expect(onStepUpdate).toHaveBeenCalledWith('s2', expect.objectContaining({ status: 'success', result: { output: 'step2' } }));
    });

    it('should stop on failure', async () => {
        const mockExecutor = vi.fn();
        mockExecutor.mockRejectedValueOnce(new Error('Failed step 1'));

        const workflow: Workflow = {
            id: '1',
            name: 'Fail Workflow',
            steps: [
                { id: 's1', name: 'Step 1', toolName: 'tool1', arguments: {} },
                { id: 's2', name: 'Step 2', toolName: 'tool2', arguments: {} }
            ],
            createdAt: '',
            updatedAt: ''
        };

        const onStepUpdate = vi.fn();

        await runWorkflow(workflow, onStepUpdate, mockExecutor);

        expect(mockExecutor).toHaveBeenCalledTimes(1);
        expect(onStepUpdate).toHaveBeenCalledWith('s1', expect.objectContaining({ status: 'error', error: 'Failed step 1' }));
        expect(mockExecutor).not.toHaveBeenCalledWith(expect.objectContaining({ name: 'tool2' }));
    });
});
