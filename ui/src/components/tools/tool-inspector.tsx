/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Dialog, DialogContent } from "@/components/ui/dialog";
import { ToolDefinition } from "@/lib/client";
import { ToolRunner } from "@/components/playground/tool-runner";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ToolInspector.
 * Wraps the ToolRunner in a Dialog for use in list views.
 *
 * @param onOpenChange - The onOpenChange.
 */
export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  // If no tool is selected but open is true, we still render Dialog to allow animation out?
  // Radix Dialog handles null children fine usually, but let's be safe.
  if (!tool && open) {
      // If open but no tool, maybe close?
      // onOpenChange(false);
      // return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[800px] h-[80vh] p-0 overflow-hidden flex flex-col gap-0 border-none bg-background">
          <ToolRunner
            tool={tool}
            onClose={() => onOpenChange(false)}
            className="border-none shadow-none"
          />
      </DialogContent>
    </Dialog>
  );
}
