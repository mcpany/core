/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Dialog, DialogContent, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { ToolDefinition } from "@/lib/client";
import { ToolRunner } from "@/components/playground/tool-runner";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ToolInspector.
 * Wrapper around ToolRunner for dialog presentation.
 *
 * @param onOpenChange - The onOpenChange.
 */
export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  if (!tool) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[900px] h-[85vh] flex flex-col p-0 gap-0 overflow-hidden bg-background border-none shadow-2xl">
         {/* Accessibility compliance */}
         <div className="sr-only">
            <DialogTitle>{tool.name}</DialogTitle>
            <DialogDescription>{tool.description}</DialogDescription>
         </div>
        <ToolRunner tool={tool} onClose={() => onOpenChange(false)} />
      </DialogContent>
    </Dialog>
  );
}
