/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Check, ChevronsUpDown, X } from "lucide-react";
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
import { ToolDefinition } from "@/lib/client";

interface ToolMultiSelectProps {
  availableTools: ToolDefinition[];
  selected: string[];
  onChange: (selected: string[]) => void;
  disabled?: boolean;
}

/**
 * ToolMultiSelect component.
 * @param props - The component props.
 * @param props.availableTools - The availableTools property.
 * @param props.selected - The selected property.
 * @param props.onChange - The onChange property.
 * @param props.disabled - The disabled property.
 * @returns The rendered component.
 */
export function ToolMultiSelect({ availableTools, selected, onChange, disabled }: ToolMultiSelectProps) {
  const [open, setOpen] = React.useState(false);

  const handleSelect = (toolName: string) => {
    if (selected.includes(toolName)) {
      onChange(selected.filter((item) => item !== toolName));
    } else {
      onChange([...selected, toolName]);
    }
  };

  const handleRemove = (toolName: string, e: React.MouseEvent) => {
    e.stopPropagation();
    onChange(selected.filter((item) => item !== toolName));
  };

  return (
    <div className="flex flex-col gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between h-auto min-h-10 py-2 px-3 text-left font-normal"
            disabled={disabled}
          >
            <div className="flex flex-wrap gap-1">
              {selected.length === 0 && <span className="text-muted-foreground">Select tools...</span>}
              {selected.map((toolName) => (
                <Badge key={toolName} variant="secondary" className="mr-1 mb-1">
                  {toolName}
                  <div
                    className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 cursor-pointer"
                    onMouseDown={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                    }}
                    onClick={(e) => handleRemove(toolName, e)}
                  >
                    <X className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                  </div>
                </Badge>
              ))}
            </div>
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0">
          <Command>
            <CommandInput placeholder="Search tools..." />
            <CommandList>
                <CommandEmpty>No tools found.</CommandEmpty>
                <CommandGroup className="max-h-64 overflow-y-auto">
                {availableTools.map((tool) => (
                    <CommandItem
                    key={tool.name}
                    value={tool.name}
                    onSelect={() => handleSelect(tool.name)}
                    >
                    <Check
                        className={cn(
                        "mr-2 h-4 w-4",
                        selected.includes(tool.name) ? "opacity-100" : "opacity-0"
                        )}
                    />
                    <div className="flex flex-col">
                        <span>{tool.name}</span>
                        {tool.description && (
                            <span className="text-xs text-muted-foreground line-clamp-1">{tool.description}</span>
                        )}
                    </div>
                    </CommandItem>
                ))}
                </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      <p className="text-[10px] text-muted-foreground">
        Select the tools that this skill is allowed to use.
      </p>
    </div>
  );
}
