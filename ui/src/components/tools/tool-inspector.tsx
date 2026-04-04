/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { Dialog, DialogContent, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { ToolDefinition } from "@/lib/client";
import { ToolRunner } from "@/components/playground/tool-runner";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ToolInspector executes the ToolInspector logic.
 *
 * Summary: Executes the ToolInspector logic.
 *
 * @param { tool - The { tool parameter.
 * @param open - The open parameter.
 * @param onOpenChange } - The onOpenChange } parameter.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
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
