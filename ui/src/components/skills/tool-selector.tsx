/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Check, ChevronsUpDown, Search, Wrench, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { apiClient, ToolDefinition } from "@/lib/client";

interface ToolSelectorProps {
  selectedTools: string[];
  onChange: (tools: string[]) => void;
}

/**
 * ToolSelector component.
 * Allows searching and selecting multiple tools from the available list.
 *
 * @param props - The component props.
 * @param props.selectedTools - The list of currently selected tool names.
 * @param props.onChange - Callback when the selection changes.
 * @returns The rendered component.
 */
export function ToolSelector({ selectedTools, onChange }: ToolSelectorProps) {
  const [open, setOpen] = React.useState(false);
  const [tools, setTools] = React.useState<ToolDefinition[]>([]);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    const fetchTools = async () => {
      setLoading(true);
      try {
        const res = await apiClient.listTools();
        setTools(res.tools || []);
      } catch (e) {
        console.error("Failed to fetch tools", e);
      } finally {
        setLoading(false);
      }
    };
    fetchTools();
  }, []);

  const handleSelect = (toolName: string) => {
    if (selectedTools.includes(toolName)) {
      onChange(selectedTools.filter((t) => t !== toolName));
    } else {
      onChange([...selectedTools, toolName]);
    }
  };

  const handleRemove = (toolName: string, e: React.MouseEvent) => {
    e.stopPropagation();
    onChange(selectedTools.filter((t) => t !== toolName));
  };

  return (
    <div className="flex flex-col gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between h-auto min-h-[40px] px-3 py-2"
          >
            {selectedTools.length > 0 ? (
              <div className="flex flex-wrap gap-1">
                {selectedTools.map((tool) => (
                  <Badge key={tool} variant="secondary" className="mr-1 mb-1">
                    {tool}
                    <div
                      role="button"
                      className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 cursor-pointer"
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          handleRemove(tool, e as any);
                        }
                      }}
                      onMouseDown={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                      }}
                      onClick={(e) => handleRemove(tool, e)}
                    >
                      <X className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                    </div>
                  </Badge>
                ))}
              </div>
            ) : (
              <span className="text-muted-foreground">Select tools...</span>
            )}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command>
            <div className="flex items-center border-b px-3" cmdk-input-wrapper="">
                <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
                <CommandInput placeholder="Search tools..." className="flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50" />
            </div>
            <CommandList>
              <CommandEmpty>{loading ? "Loading..." : "No tools found."}</CommandEmpty>
              <CommandGroup className="max-h-[300px] overflow-y-auto">
                {tools.map((tool) => (
                  <CommandItem
                    key={tool.name}
                    value={tool.name}
                    onSelect={() => handleSelect(tool.name)}
                  >
                    <div className="flex items-center w-full gap-2">
                        <Check
                        className={cn(
                            "mr-2 h-4 w-4",
                            selectedTools.includes(tool.name) ? "opacity-100" : "opacity-0"
                        )}
                        />
                        <Wrench className="h-4 w-4 text-muted-foreground" />
                        <div className="flex flex-col">
                            <span className="font-medium">{tool.name}</span>
                            {tool.description && (
                                <span className="text-xs text-muted-foreground truncate max-w-[250px]">
                                    {tool.description}
                                </span>
                            )}
                        </div>
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      {selectedTools.length === 0 && (
          <p className="text-[10px] text-muted-foreground">
              Select tools available to the skill.
          </p>
      )}
    </div>
  );
}
