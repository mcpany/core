// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

'use client';

import React, { useEffect, useState } from 'react';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Badge } from '@/components/ui/badge';
import { Check, Loader2, Info } from 'lucide-react';
import { apiClient, ToolDefinition } from '@/lib/client';
import { cn } from '@/lib/utils';
import { toast } from 'sonner';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { CaretSortIcon } from '@radix-ui/react-icons';

interface ToolSelectorProps {
  selected: string[];
  onChange: (tools: string[]) => void;
  disabled?: boolean;
}

/**
 * ToolSelector component.
 * Allows searching and selecting multiple tools from the available list.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function ToolSelector({ selected, onChange, disabled }: ToolSelectorProps) {
  const [open, setOpen] = useState(false);
  const [tools, setTools] = useState<ToolDefinition[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Only load if open to save bandwidth, or load once on mount?
    // Let's load once on mount to show count? Or on open.
    // Given the previous "Visual Skill Builder" goal, loading on mount is better UX for "Loading..." state
    // if we were showing it inline. But we are using a Popover (Combobox style).
    // Let's load on mount.
    loadTools();
  }, []);

  const loadTools = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listTools();
      setTools(res.tools || []);
    } catch (err: any) {
      toast.error('Failed to load tools: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSelect = (toolName: string) => {
    const newSelected = selected.includes(toolName)
      ? selected.filter((t) => t !== toolName)
      : [...selected, toolName];
    onChange(newSelected);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between h-auto min-h-[40px]"
          disabled={disabled}
        >
          {selected.length > 0 ? (
            <div className="flex flex-wrap gap-1 py-1 text-left">
              {selected.map((tool) => (
                <Badge key={tool} variant="secondary" className="mr-1 mb-1">
                  {tool}
                </Badge>
              ))}
            </div>
          ) : (
            <span className="text-muted-foreground">Select tools...</span>
          )}
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[400px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search tools..." className="h-9" />
          <CommandList>
            {loading && (
              <div className="flex items-center justify-center p-4 text-sm text-muted-foreground gap-2">
                <Loader2 className="h-4 w-4 animate-spin" /> Loading tools...
              </div>
            )}
            {!loading && tools.length === 0 && (
              <CommandEmpty>No tools found.</CommandEmpty>
            )}
            <CommandGroup className="max-h-[300px] overflow-y-auto">
              {tools.map((tool) => {
                const isSelected = selected.includes(tool.name);
                return (
                  <CommandItem
                    key={tool.name}
                    value={tool.name}
                    onSelect={() => handleSelect(tool.name)}
                    className="flex flex-col items-start gap-1 p-2 cursor-pointer"
                  >
                    <div className="flex items-center justify-between w-full">
                      <div className="flex items-center gap-2 font-medium">
                        <Check
                          className={cn(
                            "h-4 w-4",
                            isSelected ? "opacity-100" : "opacity-0"
                          )}
                        />
                        {tool.name}
                      </div>
                      {tool.serviceId && (
                        <Badge variant="outline" className="text-[10px] h-4 px-1">
                          {tool.serviceId}
                        </Badge>
                      )}
                    </div>
                    {tool.description && (
                      <p className="text-xs text-muted-foreground ml-6 line-clamp-2">
                        {tool.description}
                      </p>
                    )}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
          {selected.length > 0 && (
             <div className="p-2 border-t flex justify-between items-center bg-muted/20">
                 <span className="text-xs text-muted-foreground ml-2">{selected.length} selected</span>
                 <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={() => onChange([])}>
                     Clear Selection
                 </Button>
             </div>
          )}
        </Command>
      </PopoverContent>
    </Popover>
  );
}
