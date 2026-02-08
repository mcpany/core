/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

/**
 * Schema interface matching the one in schema-viewer.
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
  depth?: number;
  isRoot?: boolean;
}

/**
 * A recursive form component that generates UI controls based on a JSON Schema.
 */
export function SchemaForm({
  schema,
  value,
  onChange,
  name,
  required = false,
  depth = 0,
  isRoot = false,
}: SchemaFormProps) {
  // Determine type
  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Handle undefined value by setting default if available, or empty appropriate type
  React.useEffect(() => {
    if (value === undefined) {
      if (schema.default !== undefined) {
        onChange(schema.default);
      } else if (type === "object") {
        onChange({});
      } else if (type === "array") {
        onChange([]);
      }
    }
  }, [value, schema.default, type, onChange]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (type === "number" || type === "integer") {
      const val = parseFloat(e.target.value);
      onChange(isNaN(val) ? undefined : val);
    } else {
      onChange(e.target.value);
    }
  };

  const label = (
    <div className="flex items-center gap-2 mb-1.5">
      {name && (
        <Label className={cn("text-sm font-medium", required && "text-foreground")}>
          {name}
          {required && <span className="text-destructive ml-0.5">*</span>}
        </Label>
      )}
      {schema.description && (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>
            <Info className="h-3 w-3 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
          </TooltipTrigger>
          <TooltipContent className="max-w-[300px] text-xs">
            <p>{schema.description}</p>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  // Render based on type
  if (schema.enum) {
    return (
      <div className="w-full">
        {label}
        <Select
          value={value === undefined ? "" : String(value)}
          onValueChange={(val) => {
             // Try to restore type if enum values are numbers/booleans
             const original = schema.enum?.find(e => String(e) === val);
             onChange(original !== undefined ? original : val);
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select..." />
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

  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between space-x-2 py-1">
        <div className="flex flex-col gap-1">
           {name && (
            <Label className={cn("text-sm font-medium", required && "text-foreground")}>
              {name}
              {required && <span className="text-destructive ml-0.5">*</span>}
            </Label>
          )}
           {schema.description && (
            <span className="text-xs text-muted-foreground">{schema.description}</span>
          )}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
        />
      </div>
    );
  }

  if (type === "object" || schema.properties) {
    const properties = schema.properties || {};
    return (
      <div className={cn("space-y-4", !isRoot && "p-4 border rounded-md bg-card/50")}>
        {!isRoot && name && (
            <div className="font-semibold text-sm flex items-center gap-2 border-b pb-2 mb-4">
                {name}
                {schema.description && <span className="text-xs font-normal text-muted-foreground">- {schema.description}</span>}
            </div>
        )}
        {Object.entries(properties).map(([key, propSchema]) => (
          <SchemaForm
            key={key}
            name={key}
            schema={propSchema}
            value={value?.[key]}
            onChange={(newVal) => {
              onChange({
                ...value,
                [key]: newVal,
              });
            }}
            required={schema.required?.includes(key)}
            depth={depth + 1}
          />
        ))}
      </div>
    );
  }

  if (type === "array") {
    const items = value || [];
    return (
      <div className="space-y-2">
        {label}
        <div className="space-y-2">
          {items.map((item: any, index: number) => (
            <div key={index} className="flex gap-2 items-start">
              <div className="flex-1">
                <SchemaForm
                  schema={schema.items || {}}
                  value={item}
                  onChange={(newItem) => {
                    const newItems = [...items];
                    newItems[index] = newItem;
                    onChange(newItems);
                  }}
                  depth={depth + 1}
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => {
                  const newItems = items.filter((_: any, i: number) => i !== index);
                  onChange(newItems);
                }}
              >
                <Trash2 className="h-4 w-4 text-destructive" />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              // Add new item (try to use default or null)
              const newItem = schema.items?.default !== undefined ? schema.items.default :
                              schema.items?.type === "object" ? {} :
                              schema.items?.type === "array" ? [] : "";
              onChange([...items, newItem]);
            }}
          >
            <Plus className="mr-2 h-4 w-4" /> Add Item
          </Button>
        </div>
      </div>
    );
  }

  // Default string/number input
  return (
    <div className="w-full">
      {label}
      <Input
        type={type === "number" || type === "integer" ? "number" : "text"}
        value={value === undefined ? "" : value}
        onChange={handleInputChange}
        placeholder={schema.default ? String(schema.default) : ""}
      />
    </div>
  );
}
