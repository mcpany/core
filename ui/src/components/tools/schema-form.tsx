/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Info, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: any[];
  default?: any;
  title?: string;
  minItems?: number;
  maxItems?: number;
  [key: string]: any;
}

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  disabled?: boolean;
  depth?: number;
}

const NumberInput = ({ value, onChange, placeholder, disabled, className }: {
    value: any,
    onChange: (val: number | undefined) => void,
    placeholder?: string,
    disabled?: boolean,
    className?: string
}) => {
    const [localValue, setLocalValue] = useState<string>(value === undefined ? "" : String(value));

    // Sync local value with prop value if they differ significantly (e.g. external update)
    // We avoid syncing if localValue matches prop value numerically to preserve "1." vs "1"
    useEffect(() => {
        const numVal = Number(localValue);
        // If prop value changed and is different from our local numeric interpretation, update local
        if (value !== undefined && value !== numVal) {
             setLocalValue(String(value));
        } else if (value === undefined && localValue !== "") {
             setLocalValue("");
        }
    }, [value]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const val = e.target.value;
        setLocalValue(val);

        if (val === "") {
            onChange(undefined);
            return;
        }

        const num = Number(val);
        // Only propagate if it's a valid number and doesn't end with a dot or is just a minus sign
        // This allows typing "1." or "-" without triggering an update that would strip characters
        if (!isNaN(num) && !val.endsWith('.') && val !== "-" && val !== "-.") {
            onChange(num);
        }
    };

    return (
        <Input
            type="text" // Use text to allow "1." intermediate state
            inputMode="decimal" // Hint for mobile keyboards
            value={localValue}
            onChange={handleChange}
            placeholder={placeholder}
            disabled={disabled}
            className={className}
        />
    );
};

/**
 * A dynamic form component generated from a JSON Schema.
 * It recursively renders inputs based on the schema definition.
 */
export function SchemaForm({
  schema,
  value,
  onChange,
  name,
  required = false,
  disabled = false,
  depth = 0
}: SchemaFormProps) {

  // Helper to determine the type
  const getType = (s: Schema) => {
    if (Array.isArray(s.type)) return s.type[0];
    if (s.type) return s.type;
    if (s.properties) return "object";
    if (s.items) return "array";
    // Enum is handled separately but acts like a type
    return "string";
  };

  const type = getType(schema);
  const label = schema.title || name;
  const description = schema.description;

  // Handle Enums (Select)
  if (schema.enum) {
    return (
      <div className="grid gap-2">
        {label && (
           <div className="flex items-center gap-2">
            <Label className={cn(required && "text-foreground font-semibold")}>
              {label}
              {required && <span className="text-destructive ml-1">*</span>}
            </Label>
            {description && (
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <Info className="h-3 w-3 text-muted-foreground/70 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-[300px] text-xs">
                    <p>{description}</p>
                  </TooltipContent>
                </Tooltip>
            )}
           </div>
        )}
        <Select
          value={value === undefined ? "" : String(value)}
          onValueChange={(val) => {
            // Try to restore type if enum has mixed types, though usually strings
            const original = schema.enum?.find(e => String(e) === val);
            onChange(original !== undefined ? original : val);
          }}
          disabled={disabled}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            {schema.enum.map((opt, i) => (
              <SelectItem key={i} value={String(opt)}>
                {String(opt)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );
  }

  // Handle Booleans (Switch)
  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between rounded-lg border p-3 shadow-sm bg-muted/20">
        <div className="space-y-0.5">
          <Label className={cn(required && "text-foreground font-semibold")}>
            {label}
            {required && <span className="text-destructive ml-1">*</span>}
          </Label>
          {description && (
            <p className="text-[10px] text-muted-foreground">{description}</p>
          )}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
          disabled={disabled}
        />
      </div>
    );
  }

  // Handle Strings and Numbers (Input)
  if (type === "string" || type === "number" || type === "integer") {
    const isNumber = type === "number" || type === "integer";
    return (
      <div className="grid gap-2">
        {label && (
           <div className="flex items-center gap-2">
            <Label className={cn(required && "text-foreground font-semibold")}>
              {label}
              {required && <span className="text-destructive ml-1">*</span>}
            </Label>
            {description && (
                <Tooltip delayDuration={300}>
                  <TooltipTrigger asChild>
                    <Info className="h-3 w-3 text-muted-foreground/70 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-[300px] text-xs">
                    <p>{description}</p>
                  </TooltipContent>
                </Tooltip>
            )}
           </div>
        )}
        {isNumber ? (
            <NumberInput
                value={value}
                onChange={onChange}
                placeholder={schema.default ? String(schema.default) : undefined}
                disabled={disabled}
                className="bg-background"
            />
        ) : (
            <Input
            type="text"
            value={value === undefined ? "" : value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={schema.default ? String(schema.default) : undefined}
            disabled={disabled}
            className="bg-background"
            />
        )}
      </div>
    );
  }

  // Handle Objects (Nested Form)
  if (type === "object" && schema.properties) {
    return (
      <div className={cn("space-y-4", depth > 0 && "border rounded-lg p-4 bg-muted/10")}>
        {label && (
          <div className="flex flex-col gap-1 mb-2">
            <h4 className="font-medium text-sm flex items-center gap-2">
                {label}
                {required && <span className="text-destructive text-xs">*</span>}
            </h4>
            {description && (
                <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
        )}

        <div className="grid gap-4">
            {Object.entries(schema.properties).map(([propName, propSchema]) => (
            <SchemaForm
                key={propName}
                name={propName}
                schema={propSchema}
                value={value ? value[propName] : undefined}
                onChange={(newVal) => {
                    const newValue = { ...value, [propName]: newVal };
                    // Clean up undefined values
                    if (newVal === undefined) {
                        delete newValue[propName];
                    }
                    onChange(newValue);
                }}
                required={schema.required?.includes(propName)}
                disabled={disabled}
                depth={depth + 1}
            />
            ))}
        </div>
      </div>
    );
  }

  // Handle Arrays (Dynamic List)
  if (type === "array" && schema.items) {
    const items = Array.isArray(value) ? value : [];
    return (
      <div className={cn("space-y-2", depth > 0 && "border rounded-lg p-4 bg-muted/10")}>
         <div className="flex items-center justify-between">
            <div className="flex flex-col gap-1">
                <Label className={cn(required && "text-foreground font-semibold")}>
                {label}
                {required && <span className="text-destructive ml-1">*</span>}
                </Label>
                {description && (
                    <p className="text-xs text-muted-foreground">{description}</p>
                )}
            </div>
            <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                    onChange([...items, schema.items?.default || ""]);
                }}
                disabled={disabled}
            >
                <Plus className="h-3 w-3 mr-1" /> Add Item
            </Button>
         </div>

         <div className="space-y-2 mt-2">
            {items.map((item: any, index: number) => (
                <div key={index} className="flex gap-2 items-start relative group">
                    <div className="flex-1">
                         <SchemaForm
                            schema={schema.items as Schema}
                            value={item}
                            onChange={(newVal) => {
                                const newItems = [...items];
                                newItems[index] = newVal;
                                onChange(newItems);
                            }}
                            disabled={disabled}
                            depth={depth + 1}
                        />
                    </div>
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive absolute -right-2 top-0 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => {
                            const newItems = items.filter((_, i) => i !== index);
                            onChange(newItems);
                        }}
                        disabled={disabled}
                    >
                        <X className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            {items.length === 0 && (
                <div className="text-xs text-muted-foreground italic text-center p-2 border border-dashed rounded">
                    No items
                </div>
            )}
         </div>
      </div>
    );
  }

  // Fallback for unknown types
  return (
    <div className="grid gap-2">
        <Label className="text-muted-foreground">{label} (Unknown Type: {type})</Label>
        <div className="text-xs bg-destructive/10 text-destructive p-2 rounded">
            Unsupported schema type: {JSON.stringify(type)}
        </div>
    </div>
  );
}
