/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
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

export interface Workflow {
    id: string;
    name: string;
    description?: string;
    steps: WorkflowStep[];
    createdAt: string;
    updatedAt: string;
}
