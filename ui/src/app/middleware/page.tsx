/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { PipelineVisualizer } from "@/components/middleware/pipeline-visualizer";

/** MiddlewarePage renders the middleware pipeline visualization page. */
export default function MiddlewarePage() {
    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Middleware Pipeline</h1>
                    <p className="text-muted-foreground">Manage the request processing pipeline order.</p>
                </div>
            </div>
            <PipelineVisualizer />
        </div>
    );
}
