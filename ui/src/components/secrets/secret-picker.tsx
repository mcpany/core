/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Check, Key, ChevronsUpDown } from "lucide-react";
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
import { apiClient, SecretDefinition } from "@/lib/client";

interface SecretPickerProps {
  onSelect: (secretKey: string) => void;
  value?: string;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  /**
   * Optional custom trigger element. If provided, it replaces the default button.
   */
  children?: React.ReactNode;
}

/**
 * A component to select a secret from the stored secrets.
 * Can be used as a full combobox or attached to a custom trigger (icon).
 */
export function SecretPicker({
  onSelect,
  value,
  placeholder = "Select secret...",
  disabled,
  className,
  children
}: SecretPickerProps) {
  const [open, setOpen] = React.useState(false);
  const [secrets, setSecrets] = React.useState<SecretDefinition[]>([]);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (open) {
      setLoading(true);
      apiClient.listSecrets()
        .then((data) => {
            setSecrets(data);
        })
        .catch((err) => {
            console.error("Failed to load secrets", err);
        })
        .finally(() => {
            setLoading(false);
        });
    }
  }, [open]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        {children ? (
            children
        ) : (
            <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className={cn("w-full justify-between", className)}
            disabled={disabled}
            >
            {value ? (
                <span className="flex items-center gap-2 truncate">
                    <Key className="h-4 w-4 text-primary" />
                    {value}
                </span>
            ) : (
                <span className="text-muted-foreground flex items-center gap-2">
                    <Key className="h-4 w-4" />
                    {placeholder}
                </span>
            )}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
        )}
      </PopoverTrigger>
      <PopoverContent className="w-[300px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Search secrets..." />
          <CommandList>
            <CommandEmpty>{loading ? "Loading..." : "No secrets found."}</CommandEmpty>
            <CommandGroup>
              {secrets.map((secret) => (
                <CommandItem
                  key={secret.id}
                  value={secret.key || secret.name} // Command uses value for filtering
                  onSelect={() => {
                    const keyName = secret.key || secret.name;
                    onSelect(keyName);
                    setOpen(false);
                  }}
                >
                  <Check
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === secret.key ? "opacity-100" : "opacity-0"
                    )}
                  />
                  <div className="flex flex-col">
                      <span className="font-medium">{secret.name}</span>
                      <span className="text-xs text-muted-foreground font-mono">{secret.key}</span>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
