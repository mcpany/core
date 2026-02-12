/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Workflow, WorkflowStep } from "@/types/workflow";
import { apiClient } from "@/lib/client";

/**
 * Callback function type for updating workflow step status.
 * @param stepId - The ID of the step being updated.
 * @param updates - The partial updates to apply to the step.
 */
export type StepUpdateCallback = (stepId: string, updates: Partial<WorkflowStep>) => void;

/**
 * Executes a workflow sequentially.
 * @param workflow The workflow to run.
 * @param onStepUpdate Callback to update the UI with step progress.
 * @param executor Optional custom executor for testing.
 */
export async function runWorkflow(
    workflow: Workflow,
    onStepUpdate: StepUpdateCallback,
    executor = apiClient.executeTool
) {
    for (const step of workflow.steps) {
        onStepUpdate(step.id, { status: 'running', error: undefined, result: undefined });

        try {
            const result = await executor({
                name: step.toolName,
                arguments: step.arguments
            });

            onStepUpdate(step.id, { status: 'success', result });
        } catch (error: any) {
            const errorMsg = error instanceof Error ? error.message : String(error);
            onStepUpdate(step.id, { status: 'error', error: errorMsg });
            // Stop on first error? Or continue?
            // "Stop on failure" is standard for test suites.
            break;
        }
    }
}
