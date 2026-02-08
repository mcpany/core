/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useRef, useEffect } from "react";
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandSeparator,
} from "@/components/ui/command";
import { ToolDefinition } from "@proto/config/v1/tool";
import { useRecentTools } from "@/hooks/use-recent-tools";
import { Wrench, History } from "lucide-react";

interface SmartToolSearchProps {
  tools: ToolDefinition[];
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  onToolSelect: (tool: ToolDefinition) => void;
}

/**
 * A search component that provides autocomplete for finding and selecting tools.
 * It also displays recently used tools for quick access.
 *
 * @param props - The component props.
 * @param props.tools - The list of available tools to search.
 * @param props.searchQuery - The current search query string.
 * @param props.setSearchQuery - Callback to update the search query.
 * @param props.onToolSelect - Callback invoked when a tool is selected.
 * @returns The rendered search component.
 */
export function SmartToolSearch({
  tools,
  searchQuery,
  setSearchQuery,
  onToolSelect,
}: SmartToolSearchProps) {
  const [open, setOpen] = useState(false);
  const { recentTools, addRecent } = useRecentTools();
  const wrapperRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const handleSelect = (toolName: string) => {
    const tool = tools.find((t) => t.name === toolName);
    if (tool) {
      addRecent(toolName);
      onToolSelect(tool);
      setOpen(false);
    }
  };

  const recentToolDefs = recentTools
    .map((name) => tools.find((t) => t.name === name))
    .filter((t): t is ToolDefinition => !!t);

  return (
    <div ref={wrapperRef} className="relative w-[250px] group">
      <Command className="rounded-lg border shadow-sm overflow-visible bg-background/50 backdrop-blur-sm" shouldFilter={true}>
        <CommandInput
            placeholder="Search tools..."
            value={searchQuery}
            onValueChange={(val) => {
                setSearchQuery(val);
                if (!open) setOpen(true);
            }}
            onFocus={() => setOpen(true)}
            className="h-9"
        />

        {open && (
          <div className="absolute top-full left-0 w-full z-50 mt-1 rounded-md border bg-popover text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
            <CommandList>
              <CommandEmpty>No results found.</CommandEmpty>

              {recentToolDefs.length > 0 && (
                <CommandGroup heading="Recent">
                  {recentToolDefs.map((tool) => (
                    <CommandItem
                      key={`recent-${tool.name}`}
                      value={`${tool.name} ${tool.description}`}
                      onSelect={() => handleSelect(tool.name)}
                    >
                      <History className="mr-2 h-4 w-4 opacity-70" />
                      {tool.name}
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}

              {recentToolDefs.length > 0 && <CommandSeparator />}

              <CommandGroup heading="All Tools">
                {tools.map((tool) => (
                  <CommandItem
                    key={tool.name}
                    value={`${tool.name} ${tool.description} ${tool.serviceId}`}
                    onSelect={() => handleSelect(tool.name)}
                    className="flex justify-between"
                  >
                    <div className="flex items-center">
                        <Wrench className="mr-2 h-4 w-4 opacity-70" />
                        <span>{tool.name}</span>
                    </div>
                    {/* Show service ID as subtext */}
                    <span className="ml-2 text-xs text-muted-foreground truncate max-w-[80px]">
                        {tool.serviceId}
                    </span>
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </div>
        )}
      </Command>
    </div>
  );
}
