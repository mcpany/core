/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";

interface UnifiedDiffViewerProps {
  diff: string;
  className?: string;
}

/**
 * UnifiedDiffViewer renders a unified diff string with syntax highlighting.
 * Lines starting with '+' are green, '-' are red, and headers are blue.
 *
 * @param props - The component props.
 * @param props.diff - The raw unified diff string to render.
 * @param props.className - Optional CSS classes for the container.
 * @returns The rendered diff viewer component.
 */
export function UnifiedDiffViewer({ diff, className }: UnifiedDiffViewerProps) {
  if (!diff) return null;

  const lines = diff.split("\n");

  return (
    <div className={cn("rounded-md border bg-muted/30 font-mono text-xs overflow-hidden", className)}>
      <ScrollArea className="h-full max-h-[400px]">
        <div className="p-2">
          {lines.map((line, index) => {
            const isAdd = line.startsWith("+");
            const isDel = line.startsWith("-");
            // Standard diff headers
            const isHeader = line.startsWith("@@") || line.startsWith("diff") || line.startsWith("index") || line.startsWith("---") || line.startsWith("+++");

            return (
              <div
                key={index}
                className={cn(
                  "flex whitespace-pre-wrap break-all px-2 py-0.5 rounded-sm",
                  isAdd && "bg-green-500/10 text-green-700 dark:text-green-400",
                  isDel && "bg-red-500/10 text-red-700 dark:text-red-400",
                  isHeader && "text-blue-600 dark:text-blue-400 opacity-70 font-semibold",
                  !isAdd && !isDel && !isHeader && "text-muted-foreground"
                )}
              >
                {/* Optional line number or just symbol */}
                <span className="w-4 select-none opacity-50 shrink-0 text-center mr-2">
                    {isAdd ? '+' : isDel ? '-' : ''}
                </span>
                <span>{line}</span>
              </div>
            );
          })}
        </div>
      </ScrollArea>
    </div>
  );
}
