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
import { FileText, Play, Info, PlayCircle, PauseCircle, Trash2 } from "lucide-react";
import { PromptDefinition } from "@proto/config/v1/prompt";

interface PromptTableProps {
  prompts: PromptDefinition[];
  isCompact: boolean;
  togglePrompt: (prompt: PromptDefinition) => void;
  openInspector: (prompt: PromptDefinition) => void;
  onBulkToggle?: (names: string[], enabled: boolean) => void;
  onBulkDelete?: (names: string[]) => void;
}

/**
 * PromptTable component.
 * @param props - The component props.
 * @returns The rendered component.
 */
export const PromptTable = memo(function PromptTable({
  prompts,
  isCompact,
  togglePrompt,
  openInspector,
  onBulkToggle,
  onBulkDelete
}: PromptTableProps) {
  const [selected, setSelected] = useState<Set<string>>(new Set());

  // Reset selection when prompts list changes (e.g. filtering)
  useEffect(() => {
    setSelected(new Set());
  }, [prompts]);

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelected(new Set(prompts.map(p => p.name)));
    } else {
      setSelected(new Set());
    }
  }, [prompts]);

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

  const isAllSelected = prompts.length > 0 && selected.size === prompts.length;

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
              {onBulkDelete && (
                  <>
                    <div className="h-4 w-px bg-border mx-1" />
                    <Button size="sm" variant="ghost" onClick={() => {
                        onBulkDelete(Array.from(selected));
                        setSelected(new Set());
                    }} className="h-8 text-red-600 hover:text-red-700 hover:bg-red-100 dark:hover:bg-red-900/20">
                        <Trash2 className="mr-2 h-4 w-4" /> Delete
                    </Button>
                  </>
              )}
          </div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[30px] pr-0">
               <Checkbox
                  checked={isAllSelected}
                  onCheckedChange={(checked) => handleSelectAll(!!checked)}
                  aria-label="Select all"
                  className="translate-y-[2px]"
                />
            </TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Description</TableHead>
            <TableHead>Service</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {prompts.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="h-24 text-center">
                <div className="flex flex-col items-center justify-center text-muted-foreground">
                  <FileText className="h-8 w-8 mb-2 opacity-50" />
                  <p className="text-base font-medium">No prompts found</p>
                  <p className="text-sm opacity-70">No prompts found for this service or matching your search.</p>
                </div>
              </TableCell>
            </TableRow>
          ) : (
          prompts.map((prompt) => (
            <TableRow key={prompt.name} className={cn("group", isCompact ? "h-8" : "", selected.has(prompt.name) ? "bg-muted/50" : "")}>
               <TableCell className={cn("pr-0", isCompact ? "py-0 px-2" : "")}>
                 <Checkbox
                    checked={selected.has(prompt.name)}
                    onCheckedChange={(checked) => handleSelectOne(prompt.name, !!checked)}
                    aria-label={`Select ${prompt.name}`}
                    className="translate-y-[2px]"
                 />
              </TableCell>
              <TableCell className={cn("font-medium flex items-center", isCompact ? "py-0 px-2 h-8" : "")}>
                <FileText className={cn("mr-2 text-muted-foreground", isCompact ? "h-3 w-3" : "h-4 w-4")} />
                {prompt.name}
              </TableCell>
              <TableCell className={cn("max-w-[300px] truncate", isCompact ? "py-0 px-2" : "")} title={prompt.description}>{prompt.description}</TableCell>
              <TableCell className={isCompact ? "py-0 px-2" : ""}>
                  <Badge variant="outline" className={isCompact ? "h-5 text-[10px] px-1" : ""}>{(prompt as any).serviceId || "System"}</Badge>
              </TableCell>
              <TableCell className={isCompact ? "py-0 px-2" : ""}>
                <div className="flex items-center space-x-2">
                    <Switch
                        checked={!prompt.disable}
                        onCheckedChange={() => togglePrompt(prompt)}
                        className={isCompact ? "scale-75" : ""}
                    />
                    <span className={cn("text-muted-foreground", isCompact ? "text-[10px] w-12" : "text-sm w-16")}>
                        {!prompt.disable ? "Enabled" : "Disabled"}
                    </span>
                </div>
              </TableCell>
              <TableCell className={cn("text-right", isCompact ? "py-0 px-2" : "")}>
                  <Button variant="outline" size={isCompact ? "xs" as any : "sm"} onClick={() => openInspector(prompt)} className={isCompact ? "h-6 px-2 text-[10px]" : ""}>
                      <Play className={cn("mr-1", isCompact ? "h-2 w-2" : "h-3 w-3")} /> Inspect
                  </Button>
              </TableCell>
            </TableRow>
          )))}
        </TableBody>
      </Table>
    </div>
  );
});
