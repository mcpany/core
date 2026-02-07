/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useId } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  anyOf?: Schema[];
  oneOf?: Schema[];
  allOf?: Schema[];
  enum?: any[];
  default?: any;
  format?: string;
  title?: string;
  [key: string]: any;
}

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  depth?: number;
  disabled?: boolean;
}

export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0, disabled = false }: SchemaFormProps) {
  const uniqueId = useId();
  const fieldId = `${uniqueId}-${name || 'field'}`;

  // Determine type (handle array of types by taking first non-null)
  let type = Array.isArray(schema.type) ? schema.type.find(t => t !== "null") : schema.type;
  if (!type && schema.properties) type = "object";
  if (!type && schema.items) type = "array";
  if (!type && schema.enum) type = "string"; // fallback

  // Handle default value on mount if value is undefined
  useEffect(() => {
    if (value === undefined && schema.default !== undefined) {
      onChange(schema.default);
    }
  }, []);

  const handleChange = (newValue: any) => {
    onChange(newValue);
  };

  const renderLabel = (targetId?: string) => (
    <div className="flex items-center gap-2 mb-1.5">
      <Label htmlFor={targetId} className={cn("text-sm font-medium", depth > 0 ? "text-foreground" : "text-base")}>
        {schema.title || name || "Value"}
        {required && <span className="text-destructive ml-1">*</span>}
      </Label>
      {schema.description && (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
          </TooltipTrigger>
          <TooltipContent className="max-w-[300px] text-xs">
            <p>{schema.description}</p>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  // Enum (Select)
  if (schema.enum) {
    return (
      <div className="space-y-1">
        {renderLabel(fieldId)}
        <Select
            value={value !== undefined ? String(value) : ""}
            onValueChange={(v) => {
                const original = schema.enum?.find((e: any) => String(e) === v);
                onChange(original !== undefined ? original : v);
            }}
            disabled={disabled}
        >
          <SelectTrigger id={fieldId}>
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

  // Boolean
  if (type === "boolean") {
    return (
      <div className="flex items-center space-x-2 py-2">
        <Switch
            checked={!!value}
            onCheckedChange={handleChange}
            disabled={disabled}
            id={fieldId}
        />
        <Label htmlFor={fieldId} className="cursor-pointer font-normal">
            {schema.title || name}
            {required && <span className="text-destructive ml-1">*</span>}
        </Label>
        {schema.description && (
            <Tooltip delayDuration={300}>
                <TooltipTrigger asChild>
                <Info className="h-3.5 w-3.5 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
                </TooltipTrigger>
                <TooltipContent className="max-w-[300px] text-xs">
                <p>{schema.description}</p>
                </TooltipContent>
            </Tooltip>
        )}
      </div>
    );
  }

  // String / Number / Integer
  if (type === "string" || type === "number" || type === "integer") {
    const inputType = type === "string" ? "text" : "number";
    return (
      <div className="space-y-1">
        {renderLabel(fieldId)}
        <Input
          id={fieldId}
          type={inputType}
          value={value !== undefined ? value : ""}
          onChange={(e) => {
            const val = e.target.value;
            if ((type === "number" || type === "integer") && val !== "") {
                const num = Number(val);
                // If it parses cleanly and matches string representation (no trailing dot), send number
                if (!isNaN(num) && String(num) === val) {
                    onChange(num);
                } else {
                    // Otherwise send string (allow typing "1.")
                    onChange(val);
                }
            } else {
                if (val === "" && (type === "number" || type === "integer")) {
                    onChange(undefined);
                } else {
                    onChange(val);
                }
            }
          }}
          disabled={disabled}
          placeholder={schema.default ? String(schema.default) : ""}
        />
      </div>
    );
  }

  // Object
  if (type === "object" && schema.properties) {
    const content = (
        <div className={cn("space-y-4", depth > 0 && "p-4")}>
            {Object.entries(schema.properties).map(([key, propSchema]) => (
            <SchemaForm
                key={key}
                name={key}
                schema={propSchema}
                value={value ? value[key] : undefined}
                onChange={(newVal) => {
                    const newObj = { ...(value || {}) };
                    if (newVal === undefined) {
                        delete newObj[key];
                    } else {
                        newObj[key] = newVal;
                    }
                    onChange(newObj);
                }}
                required={schema.required?.includes(key)}
                depth={depth + 1}
                disabled={disabled}
            />
            ))}
        </div>
    );

    if (depth === 0) return content;

    return (
        <div className="space-y-1 mt-2">
            {renderLabel()}
            <Card className="border-dashed bg-muted/10 shadow-none">
                {content}
            </Card>
        </div>
    );
  }

  // Array
  if (type === "array" && schema.items) {
    const items = Array.isArray(value) ? value : [];

    return (
        <div className="space-y-2 mt-2">
            {renderLabel()}
            <div className="space-y-2 pl-2 border-l-2 border-muted">
                {items.map((item: any, index: number) => (
                    <div key={index} className="flex gap-2 items-start group">
                        <div className="flex-1">
                            <SchemaForm
                                schema={schema.items as Schema}
                                value={item}
                                onChange={(newItem) => {
                                    const newArray = [...items];
                                    newArray[index] = newItem;
                                    onChange(newArray);
                                }}
                                depth={depth + 1}
                                disabled={disabled}
                            />
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => {
                                const newArray = items.filter((_: any, i: number) => i !== index);
                                onChange(newArray);
                            }}
                            className="mt-6 h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                        const newArray = [...items, undefined];
                        onChange(newArray);
                    }}
                    className="w-full border-dashed"
                    disabled={disabled}
                >
                    <Plus className="mr-2 h-4 w-4" /> Add Item
                </Button>
            </div>
        </div>
    );
  }

  return (
    <div className="text-destructive text-xs p-2 border border-destructive/20 bg-destructive/10 rounded">
        Unsupported type: {String(type)} for {name}
    </div>
  );
}
