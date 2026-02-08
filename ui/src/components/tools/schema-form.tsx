/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useId } from "react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

/**
 * Schema definition compatible with JSON Schema.
 */
export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: any[];
  default?: any;
  format?: string;
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

/**
 * A recursive form component that renders inputs based on a JSON Schema.
 */
export function SchemaForm({
  schema,
  value,
  onChange,
  name,
  required,
  disabled,
  depth = 0
}: SchemaFormProps) {
  // Determine type
  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Handle default values initialization
  useEffect(() => {
    if (value === undefined && schema.default !== undefined) {
      onChange(schema.default);
    }
  }, [schema.default, value, onChange]);

  const uniqueId = useId();
  const inputId = name ? `${uniqueId}-${name}` : uniqueId;

  // Helper to get field label with tooltip
  const LabelWithTooltip = () => (
    <div className="flex items-center gap-2 mb-1.5">
      <Label htmlFor={inputId} className={cn(required && "text-foreground font-semibold")}>
        {name || "Root"}
        {required && <span className="text-red-500 ml-1">*</span>}
      </Label>
      {schema.description && (
        <TooltipProvider>
            <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
                <Info className="h-3.5 w-3.5 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
            </TooltipTrigger>
            <TooltipContent className="max-w-[300px] text-xs">
                <p>{schema.description}</p>
            </TooltipContent>
            </Tooltip>
        </TooltipProvider>
      )}
    </div>
  );

  // 1. Enum (Select)
  if (schema.enum) {
    return (
      <div className="w-full">
        {name && <LabelWithTooltip />}
        <Select
          disabled={disabled}
          value={value !== undefined ? String(value) : ""}
          onValueChange={(v) => {
              // Try to cast back to number if original enum had numbers
              const isNumber = schema.enum?.some(e => typeof e === 'number');
              onChange(isNumber ? Number(v) : v);
          }}
        >
          <SelectTrigger id={inputId}>
            <SelectValue placeholder="Select an option" />
          </SelectTrigger>
          <SelectContent>
            {schema.enum.map((opt: any) => (
              <SelectItem key={String(opt)} value={String(opt)}>
                {String(opt)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );
  }

  // 2. Boolean (Switch)
  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between space-x-2 py-2">
        <LabelWithTooltip />
        <Switch
          id={inputId}
          disabled={disabled}
          checked={!!value}
          onCheckedChange={onChange}
        />
      </div>
    );
  }

  // 3. Object (Nested)
  if (type === "object" || schema.properties) {
    const properties = schema.properties || {};
    const currentValue = value || {};

    const handlePropChange = (prop: string, newVal: any) => {
        const next = { ...currentValue, [prop]: newVal };
        // If undefined, maybe remove the key? For now keep it.
        onChange(next);
    };

    // If it's the root or a nested object, we render differently
    const Content = (
      <div className="space-y-4">
        {Object.entries(properties).map(([propName, propSchema]) => (
          <SchemaForm
            key={propName}
            name={propName}
            schema={propSchema}
            value={currentValue[propName]}
            onChange={(v) => handlePropChange(propName, v)}
            required={schema.required?.includes(propName)}
            disabled={disabled}
            depth={depth + 1}
          />
        ))}
      </div>
    );

    if (depth === 0 && !name) {
        // Root object, just render content
        return Content;
    }

    return (
      <Card className="border-dashed shadow-sm bg-muted/20">
        <CardHeader className="py-3 px-4">
             <div className="flex items-center gap-2">
                <span className="text-sm font-semibold">{name}</span>
                {schema.description && (
                     <span className="text-xs text-muted-foreground truncate max-w-[300px]" title={schema.description}>
                        - {schema.description}
                     </span>
                )}
             </div>
        </CardHeader>
        <CardContent className="p-4 pt-0">
             {Content}
        </CardContent>
      </Card>
    );
  }

  // 4. Array (List)
  if (type === "array") {
    const items = value || [];
    const itemSchema = schema.items || {};

    const addItem = () => {
      onChange([...items, itemSchema.default !== undefined ? itemSchema.default : (itemSchema.type === "object" ? {} : "")]);
    };

    const removeItem = (index: number) => {
      const next = [...items];
      next.splice(index, 1);
      onChange(next);
    };

    const updateItem = (index: number, val: any) => {
      const next = [...items];
      next[index] = val;
      onChange(next);
    };

    return (
      <div className="space-y-2">
        <div className="flex items-center justify-between">
            <LabelWithTooltip />
            <Button type="button" variant="outline" size="sm" onClick={addItem} disabled={disabled} className="h-6 text-xs">
                <Plus className="mr-1 h-3 w-3" /> Add Item
            </Button>
        </div>
        <div className="space-y-2 pl-2 border-l-2 border-muted">
            {items.map((item: any, idx: number) => (
                <div key={idx} className="flex gap-2 items-start group">
                    <div className="flex-1">
                        <SchemaForm
                            schema={itemSchema}
                            value={item}
                            onChange={(v) => updateItem(idx, v)}
                            disabled={disabled}
                            depth={depth + 1}
                        />
                    </div>
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => removeItem(idx)}
                        disabled={disabled}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            {items.length === 0 && (
                <div className="text-xs text-muted-foreground italic py-2">No items.</div>
            )}
        </div>
      </div>
    );
  }

  // 5. String / Number / Integer
  return (
    <div className="w-full">
      {name && <LabelWithTooltip />}
      <Input
        id={inputId}
        type={type === "integer" || type === "number" ? "number" : "text"}
        value={value !== undefined ? value : ""}
        onChange={(e) => {
          const val = e.target.value;
          if (type === "integer" || type === "number") {
             // Handle empty string as undefined or 0?
             // If empty, let's pass undefined to allow backspace clearing
             if (val === "") {
                 onChange(undefined);
             } else {
                 // Pass string to allow typing decimals (e.g. "1.")
                 onChange(val);
             }
          } else {
            onChange(val);
          }
        }}
        disabled={disabled}
        placeholder={schema.default ? `Default: ${schema.default}` : ""}
      />
    </div>
  );
}
