/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { memo, useState, useCallback, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import { Wrench, Play, Star, Info, PlayCircle, PauseCircle, StarOff } from "lucide-react";
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
  openInspector,
  usageStats,
  onBulkToggle,
  onBulkPin
}: ToolTableProps) {
  const [selected, setSelected] = useState<Set<string>>(new Set());

  // Reset selection when tools list changes (e.g. filtering)
  useEffect(() => {
    setSelected(new Set());
  }, [tools]);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelected(new Set(tools.map(t => t.name)));
    } else {
      setSelected(new Set());
    }
  }, [tools]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    setSelected(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(name);
        } else {
          newSelected.delete(name);
        }
        return newSelected;
    });
  }, []);

  const isAllSelected = tools.length > 0 && selected.size === tools.length;

  return (
    <div className="space-y-2">
      {selected.size > 0 && (
          <div className="flex items-center gap-2 p-2 bg-muted/40 rounded-md animate-in fade-in slide-in-from-top-1 duration-200 sticky top-0 z-10 backdrop-blur-md border">
              <span className="text-sm text-muted-foreground mr-2 font-medium px-2">{selected.size} selected</span>
              <div className="h-4 w-px bg-border mx-1" />
              {onBulkToggle && (
                  <>
                    <Button size="sm" variant="ghost" onClick={() => {
                        onBulkToggle(Array.from(selected), true);
                        setSelected(new Set());
                    }} className="h-8 text-green-600 hover:text-green-700 hover:bg-green-100 dark:hover:bg-green-900/20">
                        <PlayCircle className="mr-2 h-4 w-4" /> Enable
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => {
                        onBulkToggle(Array.from(selected), false);
                        setSelected(new Set());
                    }} className="h-8 text-amber-600 hover:text-amber-700 hover:bg-amber-100 dark:hover:bg-amber-900/20">
                        <PauseCircle className="mr-2 h-4 w-4" /> Disable
                    </Button>
                  </>
              )}
              {onBulkPin && (
                  <>
                    <div className="h-4 w-px bg-border mx-1" />
                    <Button size="sm" variant="ghost" onClick={() => {
                        onBulkPin(Array.from(selected), true);
                        setSelected(new Set());
                    }} className="h-8 text-yellow-600 hover:text-yellow-700 hover:bg-yellow-100 dark:hover:bg-yellow-900/20">
                        <Star className="mr-2 h-4 w-4 fill-current" /> Pin
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => {
                        onBulkPin(Array.from(selected), false);
                        setSelected(new Set());
                    }} className="h-8">
                        <StarOff className="mr-2 h-4 w-4" /> Unpin
                    </Button>
                  </>
              )}
          </div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[40px] px-2">
                <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={handleSelectAll}
                    aria-label="Select all"
                />
            </TableHead>
            <TableHead className="w-[40px] px-2"></TableHead>
            <TableHead>Tool</TableHead>
            {!isCompact && <TableHead>Description</TableHead>}
            {!isCompact && <TableHead className="hidden md:table-cell">Provider</TableHead>}
            {!isCompact && <TableHead className="text-right">Usage (24h)</TableHead>}
            <TableHead className="w-[100px] text-center">Status</TableHead>
            <TableHead className="w-[60px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {tools.map((tool) => (
            <TableRow
              key={tool.name}
              className={cn(tool.disable && "opacity-60", "group cursor-pointer")}
              onClick={(e) => {
                  // Don't trigger if clicking buttons/switches
                  if ((e.target as HTMLElement).closest('button') || (e.target as HTMLElement).closest('[role="switch"]')) {
                      return;
                  }
                  openInspector(tool);
              }}
            >
              <TableCell className="px-2">
                  <Checkbox
                      checked={selected.has(tool.name)}
                      onCheckedChange={(c) => handleSelectOne(tool.name, !!c)}
                      aria-label={`Select ${tool.name}`}
                  />
              </TableCell>
              <TableCell className="px-2">
                <Button
                  variant="ghost"
                  size="icon"
                  className={cn(
                    "h-8 w-8",
                    isPinned(tool.name) ? "text-yellow-500 hover:text-yellow-600" : "text-muted-foreground opacity-0 group-hover:opacity-100"
                  )}
                  onClick={(e) => { e.stopPropagation(); togglePin(tool.name); }}
                >
                  <Star className={cn("h-4 w-4", isPinned(tool.name) && "fill-current")} />
                </Button>
              </TableCell>
              <TableCell className="font-medium">
                <div className="flex items-center gap-2">
                  <Wrench className="h-4 w-4 text-muted-foreground" />
                  <span className="truncate max-w-[200px]" title={tool.name}>{tool.name}</span>
                  {tool.disable && (
                    <Badge variant="outline" className="ml-2">Disabled</Badge>
                  )}
                </div>
              </TableCell>
              {!isCompact && (
                <TableCell className="text-muted-foreground text-sm max-w-[300px] truncate" title={tool.description}>
                  {tool.description || "No description provided."}
                </TableCell>
              )}
              {!isCompact && (
                <TableCell className="hidden md:table-cell">
                  <div className="flex items-center text-muted-foreground text-xs" title={`${estimateTokens(JSON.stringify(tool))} tokens`}>
                      {formatTokenCount(estimateTokens(JSON.stringify(tool)))}
                  </div>
                </TableCell>
              )}
              {!isCompact && (
                  <TableCell className="text-right text-muted-foreground">
                      {usageStats?.[tool.name]?.totalCalls || 0} calls
                  </TableCell>
              )}
              <TableCell className="text-center">
                <Switch
                  checked={!tool.disable}
                  onCheckedChange={(checked) => { toggleTool(tool.name, !checked); }}
                  onClick={(e) => e.stopPropagation()}
                />
              </TableCell>
              <TableCell>
                <Button variant="ghost" size="icon" onClick={(e) => { e.stopPropagation(); openInspector(tool); }}>
                  <Info className="h-4 w-4" />
                </Button>
              </TableCell>
            </TableRow>
          ))}
          {tools.length === 0 && (
            <TableRow>
              <TableCell colSpan={isCompact ? 5 : 8} className="h-24 text-center text-muted-foreground">
                No tools found matching your criteria.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
});
