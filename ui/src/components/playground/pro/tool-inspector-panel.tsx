/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import { ToolForm } from "@/components/playground/tool-form";
import { Button } from "@/components/ui/button";
import { X, Zap } from "lucide-react";

interface ToolInspectorPanelProps {
  tool: ToolDefinition | null;
  onClose: () => void;
  onInsert: (data: Record<string, unknown>) => void;
  onRun: (data: Record<string, unknown>) => void;
}

/**
 * ToolInspectorPanel component.
 * Displays the configuration form for a selected tool in a side panel.
 *
 * @param props - The component props.
 * @param props.tool - The tool to configure.
 * @param props.onClose - Callback when the panel is closed.
 * @param props.onInsert - Callback when "Insert" is clicked.
 * @param props.onRun - Callback when "Run" is clicked.
 * @returns The rendered component.
 */
export function ToolInspectorPanel({ tool, onClose, onInsert, onRun }: ToolInspectorPanelProps) {
  if (!tool) {
      return (
          <div className="h-full flex flex-col items-center justify-center text-muted-foreground p-8 text-center border-l bg-muted/5">
              <Zap className="h-12 w-12 opacity-10 mb-4" />
              <p className="text-sm">Select a tool from the library to configure it.</p>
          </div>
      );
  }

  return (
    <div className="h-full flex flex-col bg-background border-l shadow-xl z-20 relative">
      <div className="flex items-center justify-between p-4 border-b bg-muted/10">
        <div className="flex items-center gap-2 overflow-hidden">
            <div className="bg-primary/10 p-1.5 rounded-md shrink-0">
                <Zap className="w-4 h-4 text-primary" />
            </div>
            <div className="flex flex-col min-w-0">
                <h3 className="font-semibold text-sm truncate" title={tool.name}>
                {tool.name}
                </h3>
                <span className="text-[10px] text-muted-foreground truncate">{tool.serviceId || "core"}</span>
            </div>
        </div>
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onClose} title="Close Inspector">
          <X className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex-1 overflow-hidden p-4">
        {/* @ts-expect-error: ToolForm will be updated to accept onRun in the next step */}
        <ToolForm
          tool={tool}
          onSubmit={onInsert}
          onRun={onRun}
          onCancel={onClose}
        />
      </div>
    </div>
  );
}
