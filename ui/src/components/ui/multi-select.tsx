"use client";

import * as React from "react";
import { X, ChevronsUpDown, Check } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Command,
  CommandGroup,
  CommandItem,
  CommandList,
  CommandInput,
  CommandEmpty,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export interface Option {
  label: string;
  value: string;
  description?: string;
}

interface MultiSelectProps {
  options: Option[];
  selected: string[];
  onChange: (selected: string[]) => void;
  placeholder?: string;
  className?: string;
}

export function MultiSelect({
  options,
  selected,
  onChange,
  placeholder = "Select items...",
  className,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false);

  const handleUnselect = (item: string) => {
    onChange(selected.filter((i) => i !== item));
  };

  const handleSelect = (item: string) => {
    if (selected.includes(item)) {
        onChange(selected.filter((i) => i !== item));
    } else {
        onChange([...selected, item]);
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            "w-full justify-between h-auto min-h-10 px-3 py-2 hover:bg-background",
            className
          )}
        >
          <div className="flex flex-wrap gap-1">
            {selected.length > 0 ? (
                selected.map((item) => (
                    <Badge
                    variant="secondary"
                    key={item}
                    className="mr-1 mb-1 font-normal"
                    onClick={(e) => {
                        e.stopPropagation();
                        handleUnselect(item);
                    }}
                    >
                    {options.find((opt) => opt.value === item)?.label || item}
                    <button
                        className="ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                        onKeyDown={(e) => {
                        if (e.key === "Enter") {
                            handleUnselect(item);
                        }
                        }}
                        onMouseDown={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        }}
                        onClick={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        handleUnselect(item);
                        }}
                    >
                        <X className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                    </button>
                    </Badge>
                ))
            ) : (
                <span className="text-muted-foreground text-sm font-normal">
                    {placeholder}
                </span>
            )}
          </div>
          <ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50 ml-2" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-full p-0">
        <Command className="w-full">
          <CommandInput placeholder="Search..." />
          <CommandList className="max-h-64">
             <CommandEmpty>No item found.</CommandEmpty>
             <CommandGroup>
                {options.map((option) => (
                    <CommandItem
                        key={option.value}
                        onSelect={() => handleSelect(option.value)}
                        className="cursor-pointer"
                    >
                        <div
                            className={cn(
                                "mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                                selected.includes(option.value)
                                ? "bg-primary text-primary-foreground"
                                : "opacity-50 [&_svg]:invisible"
                            )}
                        >
                            <Check className={cn("h-4 w-4")} />
                        </div>
                        <div className="flex flex-col">
                            <span>{option.label}</span>
                            {option.description && (
                                <span className="text-xs text-muted-foreground">{option.description}</span>
                            )}
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
