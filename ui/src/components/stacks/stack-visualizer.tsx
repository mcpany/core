/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React from "react";
import { Layers } from "lucide-react";
import { StackGraph } from "@/components/stacks/stack-graph";

interface StackVisualizerProps {
    yamlContent: string;
}

/**
 * StackVisualizer executes the StackVisualizer logic.
 *
 * Summary: Executes the StackVisualizer logic.
 *
 * @param { yamlContent } - The { yamlContent } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function StackVisualizer({ yamlContent }: StackVisualizerProps) {
    return (
        <div className="flex flex-col h-full bg-muted/5 w-[280px] border-l stack-visualizer-container">
            <div className="p-4 border-b flex items-center justify-between shrink-0 bg-background/50 backdrop-blur-sm z-10">
                <div className="flex items-center gap-2 font-semibold">
                    <Layers className="h-5 w-5 text-primary" />
                    <span>Live Preview</span>
                </div>
            </div>
            <div className="flex-1 relative overflow-hidden">
                <StackGraph yamlContent={yamlContent} />
            </div>
        </div>
    );
}
