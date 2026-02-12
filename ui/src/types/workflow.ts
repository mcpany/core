/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents a single step in a workflow.
 */
export interface WorkflowStep {
    id: string;
    name: string;
    toolName: string;
    arguments: Record<string, unknown>;
    status?: 'pending' | 'running' | 'success' | 'error';
    result?: unknown;
    error?: string;
}

/**
 * Represents a workflow containing multiple steps.
 */
export interface Workflow {
    id: string;
    name: string;
    description?: string;
    steps: WorkflowStep[];
    createdAt: string;
    updatedAt: string;
}
