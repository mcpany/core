/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { memo, useState, useCallback, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import { Wrench, Play, Star, Info, PlayCircle, PauseCircle, StarOff, CheckSquare } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";
import { ToolAnalytics } from "@/lib/client";

interface ToolTableProps {
  tools: ToolDefinition[];
  isCompact: boolean;
  isPinned: (name: string) => boolean;
  togglePin: (name: string) => void;
  toggleTool: (name: string, currentStatus: boolean) => void;
  openInspector: (tool: ToolDefinition) => void;
  usageStats?: Record<string, ToolAnalytics>;
  onBulkToggle?: (names: string[], enabled: boolean) => void;
  onBulkPin?: (names: string[], pinned: boolean) => void;
}

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
  openInspector,
  usageStats,
  onBulkToggle,
  onBulkPin
}: ToolTableProps) {
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set());

  // Reset selection when tools list changes (e.g. filtering) to avoid stale selection
  useEffect(() => {
      setSelectedTools(new Set());
  }, [tools]);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedTools(new Set(tools.map(t => t.name)));
    } else {
      setSelectedTools(new Set());
    }
  }, [tools]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    setSelectedTools(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(name);
        } else {
          newSelected.delete(name);
        }
        return newSelected;
    });
  }, []);

  const isAllSelected = tools.length > 0 && selectedTools.size === tools.length;

  return (
    <div className="space-y-4">
        {/* Bulk Actions Toolbar */}
        {selectedTools.size > 0 && (
            <div className="flex items-center gap-2 animate-in fade-in slide-in-from-top-2 p-2 bg-muted/40 rounded-md border border-muted-foreground/10 mb-4 sticky top-0 z-10 backdrop-blur-md">
                <span className="text-sm font-medium ml-2 mr-2">
                    {selectedTools.size} selected
                </span>
                <div className="h-4 w-px bg-border mx-2" />
                {onBulkToggle && (
                    <>
                        <Button size="sm" variant="outline" className="h-8 text-xs" onClick={() => {
                            onBulkToggle(Array.from(selectedTools), true);
                            setSelectedTools(new Set());
                        }}>
                            <PlayCircle className="mr-2 h-3.5 w-3.5 text-green-600" /> Enable
                        </Button>
                        <Button size="sm" variant="outline" className="h-8 text-xs" onClick={() => {
                            onBulkToggle(Array.from(selectedTools), false);
                            setSelectedTools(new Set());
                        }}>
                            <PauseCircle className="mr-2 h-3.5 w-3.5 text-amber-600" /> Disable
                        </Button>
                    </>
                )}
                {onBulkPin && (
                    <>
                         <Button size="sm" variant="outline" className="h-8 text-xs" onClick={() => {
                            onBulkPin(Array.from(selectedTools), true);
                            setSelectedTools(new Set());
                        }}>
                            <Star className="mr-2 h-3.5 w-3.5 text-yellow-500" /> Pin
                        </Button>
                        <Button size="sm" variant="outline" className="h-8 text-xs" onClick={() => {
                            onBulkPin(Array.from(selectedTools), false);
                            setSelectedTools(new Set());
                        }}>
                            <StarOff className="mr-2 h-3.5 w-3.5" /> Unpin
                        </Button>
                    </>
                )}
                <div className="flex-1" />
                <Button size="sm" variant="ghost" className="h-8 text-xs" onClick={() => setSelectedTools(new Set())}>
                    Cancel
                </Button>
            </div>
        )}

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[40px]">
                  <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={(checked) => handleSelectAll(!!checked)}
                    aria-label="Select all"
                  />
              </TableHead>
              <TableHead className="w-[30px]"></TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead className="text-right">Calls</TableHead>
              <TableHead className="text-right">Success</TableHead>
              <TableHead title="Estimated context size when tool is defined">Est. Context</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tools.map((tool) => (
              <TableRow key={tool.name} className={cn("group", isCompact ? "h-8" : "", selectedTools.has(tool.name) ? "bg-muted/30" : "")}>
                <TableCell className={isCompact ? "py-0 px-2" : ""}>
                    <Checkbox
                        checked={selectedTools.has(tool.name)}
                        onCheckedChange={(checked) => handleSelectOne(tool.name, !!checked)}
                        aria-label={`Select ${tool.name}`}
                    />
                </TableCell>
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
                <TableCell className={cn("text-right font-mono", isCompact ? "py-0 px-2" : "")}>
                    {usageStats?.[`${tool.name}@${tool.serviceId}`]?.totalCalls || "-"}
                </TableCell>
                <TableCell className={cn("text-right", isCompact ? "py-0 px-2" : "")}>
                    {(() => {
                        const stats = usageStats?.[`${tool.name}@${tool.serviceId}`];
                        if (!stats || stats.totalCalls === 0) return "-";
                        const rate = stats.successRate;
                        let color = "text-green-500";
                        if (rate < 50) color = "text-red-500";
                        else if (rate < 90) color = "text-yellow-500";
                        return <span className={cn("font-bold", color)}>{rate.toFixed(1)}%</span>;
                    })()}
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
            {tools.length === 0 && (
                <TableRow>
                    <TableCell colSpan={10} className="text-center py-8 text-muted-foreground">
                        No tools found.
                    </TableCell>
                </TableRow>
            )}
          </TableBody>
        </Table>
    </div>
  );
});
