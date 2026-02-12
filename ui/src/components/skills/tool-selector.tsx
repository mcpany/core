/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Check, ChevronsUpDown, Loader2, Wrench } from "lucide-react";
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
 * Allows selecting multiple tools from a searchable list.
 *
 * @param props - The component props.
 * @param props.selectedTools - The list of currently selected tool names.
 * @param props.onChange - Callback when selection changes.
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
        const data = await apiClient.listTools();
        setTools(data.tools || []);
      } catch (error) {
        console.error("Failed to fetch tools", error);
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

  return (
    <div className="flex flex-col gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
          >
            {selectedTools.length > 0
              ? `${selectedTools.length} tool${selectedTools.length > 1 ? "s" : ""} selected`
              : "Select tools..."}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command>
            <CommandInput placeholder="Search tools..." />
            <CommandList>
              <CommandEmpty>
                  {loading ? (
                      <div className="flex items-center justify-center p-4 text-muted-foreground">
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Loading tools...
                      </div>
                  ) : "No tools found."}
              </CommandEmpty>
              <CommandGroup>
                {tools.map((tool) => (
                  <CommandItem
                    key={tool.name}
                    value={tool.name}
                    onSelect={() => handleSelect(tool.name)}
                  >
                    <div className="flex items-center gap-2 w-full overflow-hidden">
                        <Check
                        className={cn(
                            "mr-2 h-4 w-4",
                            selectedTools.includes(tool.name) ? "opacity-100" : "opacity-0"
                        )}
                        />
                        <Wrench className="h-4 w-4 text-muted-foreground" />
                        <div className="flex flex-col flex-1 min-w-0">
                            <span className="truncate font-medium">{tool.name}</span>
                            {tool.description && (
                                <span className="text-xs text-muted-foreground truncate">
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

      {selectedTools.length > 0 && (
        <div className="flex flex-wrap gap-2 mt-2">
          {selectedTools.map((tool) => (
            <Badge key={tool} variant="secondary" className="flex items-center gap-1">
              {tool}
              <button
                className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    handleSelect(tool);
                  }
                }}
                onMouseDown={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                }}
                onClick={() => handleSelect(tool)}
              >
                <span className="sr-only">Remove {tool}</span>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="h-3 w-3 hover:text-destructive"
                >
                    <path d="M18 6 6 18" />
                    <path d="m6 6 12 12" />
                </svg>
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
