/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { memo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Wrench, Play, Star, Info } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";

interface ToolTableProps {
  tools: ToolDefinition[];
  isCompact: boolean;
  isPinned: (name: string) => boolean;
  togglePin: (name: string) => void;
  toggleTool: (name: string, currentStatus: boolean) => void;
  openInspector: (tool: ToolDefinition) => void;
}

// ⚡ Bolt Optimization: Extracted ToolTable from page.tsx to prevent unnecessary re-renders
// of the entire table when parent state (like search query) changes.
// Memoization ensures table only updates when props change.

/**
 * ToolTable component.
 * @param props - The component props.
 * @returns The rendered component.
 */
export const ToolTable = memo(function ToolTable({
  tools,
  isCompact,
  isPinned,
  togglePin,
  toggleTool,
  openInspector
}: ToolTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[30px]"></TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Service</TableHead>
          <TableHead title="Estimated context size when tool is defined">Est. Context</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tools.map((tool) => (
          <TableRow key={tool.name} className={cn("group", isCompact ? "h-8" : "")}>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={() => togglePin(tool.name)}
                  aria-label={`Pin ${tool.name}`}
                >
                    <Star className={`h-4 w-4 ${isPinned(tool.name) ? "fill-yellow-400 text-yellow-400" : "text-muted-foreground"}`} />
                </Button>
            </TableCell>
            <TableCell className={cn("font-medium flex items-center", isCompact ? "py-0 px-2 h-8" : "")}>
              <Wrench className={cn("mr-2 text-muted-foreground", isCompact ? "h-3 w-3" : "h-4 w-4")} />
              {tool.name}
            </TableCell>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>{tool.description}</TableCell>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>
                <Badge variant="outline" className={isCompact ? "h-5 text-[10px] px-1" : ""}>{tool.serviceId}</Badge>
            </TableCell>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>
                <div className="flex items-center text-muted-foreground text-xs" title={`${estimateTokens(JSON.stringify(tool))} tokens`}>
                    <Info className="w-3 h-3 mr-1 opacity-50" />
                    {formatTokenCount(estimateTokens(JSON.stringify(tool)))}
                </div>
            </TableCell>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>
              <div className="flex items-center space-x-2">
                  <Switch
                      checked={!tool.disable}
                      onCheckedChange={() => toggleTool(tool.name, !tool.disable)}
                      className={isCompact ? "scale-75" : ""}
                  />
                  <span className={cn("text-muted-foreground", isCompact ? "text-[10px] w-12" : "text-sm w-16")}>
                      {!tool.disable ? "Enabled" : "Disabled"}
                  </span>
              </div>
            </TableCell>
            <TableCell className={cn("text-right", isCompact ? "py-0 px-2" : "")}>
                <Button variant="outline" size={isCompact ? "xs" as any : "sm"} onClick={() => openInspector(tool)} className={isCompact ? "h-6 px-2 text-[10px]" : ""}>
                    <Play className={cn("mr-1", isCompact ? "h-2 w-2" : "h-3 w-3")} /> Inspect
                </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
});
