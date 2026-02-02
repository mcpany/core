/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Check, ChevronsUpDown, Loader2, X } from "lucide-react";
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
import { toast } from "sonner";

interface ToolSelectorProps {
  value: string[];
  onChange: (value: string[]) => void;
  disabled?: boolean;
  id?: string;
}

/**
 * ToolSelector component.
 * @param props - The component props.
 * @param props.value - The value.
 * @param props.onChange - The onChange.
 * @param props.disabled - The disabled.
 * @param props.id - The id.
 * @returns The rendered component.
 */
export function ToolSelector({ value = [], onChange, disabled, id }: ToolSelectorProps) {
  const [open, setOpen] = React.useState(false);
  const [tools, setTools] = React.useState<ToolDefinition[]>([]);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    const fetchTools = async () => {
      setLoading(true);
      try {
        const res = await apiClient.listTools();
        setTools(res.tools);
      } catch (e) {
        console.error("Failed to fetch tools", e);
        toast.error("Failed to load available tools.");
      } finally {
        setLoading(false);
      }
    };
    fetchTools();
  }, []);

  // Group tools by serviceId
  const groupedTools = React.useMemo(() => {
    const groups: Record<string, ToolDefinition[]> = {};
    tools.forEach((tool) => {
      const serviceId = tool.serviceId || "Unknown Service";
      if (!groups[serviceId]) {
        groups[serviceId] = [];
      }
      groups[serviceId].push(tool);
    });
    return groups;
  }, [tools]);

  const handleSelect = (toolName: string) => {
    const newValue = value.includes(toolName)
      ? value.filter((t) => t !== toolName)
      : [...value, toolName];
    onChange(newValue);
  };

  const handleRemove = (toolName: string) => {
    onChange(value.filter((t) => t !== toolName));
  };

  return (
    <div className="flex flex-col gap-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            id={id}
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
            disabled={disabled}
          >
            {value.length > 0
              ? `${value.length} tool${value.length === 1 ? "" : "s"} selected`
              : "Select tools..."}
            {loading ? (
              <Loader2 className="ml-2 h-4 w-4 animate-spin opacity-50" />
            ) : (
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command>
            <CommandInput placeholder="Search tools..." />
            <CommandList>
              <CommandEmpty>No tools found.</CommandEmpty>
              {Object.entries(groupedTools).map(([serviceId, serviceTools]) => (
                <CommandGroup key={serviceId} heading={serviceId}>
                  {serviceTools.map((tool) => (
                    <CommandItem
                      key={`${serviceId}-${tool.name}`}
                      value={`${tool.name} ${tool.description || ''}`}
                      onSelect={() => handleSelect(tool.name)}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          value.includes(tool.name) ? "opacity-100" : "opacity-0"
                        )}
                      />
                      <div className="flex flex-col">
                        <span>{tool.name}</span>
                        {tool.description && (
                          <span className="text-[10px] text-muted-foreground truncate max-w-[280px]">
                            {tool.description}
                          </span>
                        )}
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              ))}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {/* Selected Items Badges */}
      {value.length > 0 && (
        <div className="flex flex-wrap gap-2 mt-2">
          {value.map((toolName) => (
            <Badge key={toolName} variant="secondary" className="pl-2 pr-1 py-1 flex items-center gap-1">
              {toolName}
              <button
                type="button"
                onClick={() => handleRemove(toolName)}
                className="ml-1 hover:bg-muted-foreground/20 rounded-full p-0.5 focus:outline-none"
                disabled={disabled}
                aria-label={`Remove ${toolName}`}
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
