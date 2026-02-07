/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

// Re-use Schema definition from SchemaViewer if possible, but redefine for clarity here
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
}

/**
 * A recursive form component that renders inputs based on a JSON Schema.
 */
export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0 }: SchemaFormProps) {
  // Determine type
  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Helper to handle change
  const handleChange = (newValue: any) => {
    onChange(newValue);
  };

  // Render Label with Tooltip
  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-1.5">
      <Label className={cn("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", required && "text-foreground")}>
        {name}
        {required && <span className="text-red-500 ml-1">*</span>}
      </Label>
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

  // 1. Enum (Select)
  if (schema.enum) {
    return (
      <div className="flex flex-col gap-1.5 w-full">
        {name && renderLabel()}
        <Select
          value={value !== undefined ? String(value) : undefined}
          onValueChange={(val) => {
              // Try to cast back to original type if number/boolean
              if (type === "number" || type === "integer") handleChange(Number(val));
              else if (type === "boolean") handleChange(val === "true");
              else handleChange(val);
          }}
        >
          <SelectTrigger className="w-full">
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

  // 2. Boolean (Switch)
  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between space-x-2 border rounded-md p-3">
        <div className="flex flex-col gap-1">
            <Label className="text-base">{name}</Label>
            {schema.description && <span className="text-xs text-muted-foreground">{schema.description}</span>}
        </div>
        <Switch
          checked={value === true}
          onCheckedChange={handleChange}
        />
      </div>
    );
  }

  // 3. String (Input / Textarea)
  if (type === "string") {
    // If format implies long text, use Textarea? For now default to Input
    return (
      <div className="flex flex-col gap-1.5 w-full">
        {name && renderLabel()}
        <Input
          value={value || ""}
          onChange={(e) => handleChange(e.target.value)}
          placeholder={schema.default ? String(schema.default) : ""}
        />
      </div>
    );
  }

  // 4. Number (Input type="number")
  if (type === "number" || type === "integer") {
    return (
      <div className="flex flex-col gap-1.5 w-full">
        {name && renderLabel()}
        <Input
          type="number"
          value={value !== undefined ? value : ""}
          onChange={(e) => {
            const val = e.target.value;
            handleChange(val === "" ? undefined : Number(val));
          }}
          placeholder={schema.default ? String(schema.default) : ""}
        />
      </div>
    );
  }

  // 5. Object (Recursion)
  if (type === "object" || schema.properties) {
    const properties = schema.properties || {};
    const currentValue = value || {};

    const handlePropChange = (propKey: string, propValue: any) => {
        const newValue = { ...currentValue };
        if (propValue === undefined) {
            delete newValue[propKey];
        } else {
            newValue[propKey] = propValue;
        }
        onChange(newValue);
    };

    return (
      <div className={cn("space-y-4 w-full", depth > 0 && "pl-4 border-l-2 border-muted")}>
        {name && (
            <div className="flex items-center gap-2 mb-2">
                 <h4 className="font-semibold text-sm">{name}</h4>
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
        )}
        <div className="grid gap-4">
            {Object.entries(properties).map(([key, propSchema]) => (
            <SchemaForm
                key={key}
                name={key}
                schema={propSchema}
                value={currentValue[key]}
                onChange={(val) => handlePropChange(key, val)}
                required={schema.required?.includes(key)}
                depth={depth + 1}
            />
            ))}
        </div>
      </div>
    );
  }

  // 6. Array (List)
  if (type === "array" && schema.items) {
      const items = Array.isArray(value) ? value : [];
      const itemSchema = schema.items;

      const addItem = () => {
          const newItem = itemSchema.default !== undefined ? itemSchema.default : (
              itemSchema.type === "string" ? "" :
              itemSchema.type === "number" ? 0 :
              itemSchema.type === "boolean" ? false :
              itemSchema.type === "object" ? {} :
              null
          );
          handleChange([...items, newItem]);
      };

      const removeItem = (index: number) => {
          const newItems = [...items];
          newItems.splice(index, 1);
          handleChange(newItems);
      };

      const updateItem = (index: number, val: any) => {
          const newItems = [...items];
          newItems[index] = val;
          handleChange(newItems);
      };

      return (
        <div className="space-y-2 w-full">
            {name && renderLabel()}
            <div className="space-y-2 pl-2 border-l-2 border-dashed border-muted">
                {items.map((item: any, idx: number) => (
                    <div key={idx} className="flex gap-2 items-start">
                        <div className="flex-1">
                            <SchemaForm
                                schema={itemSchema}
                                value={item}
                                onChange={(val) => updateItem(idx, val)}
                                depth={depth + 1}
                            />
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => removeItem(idx)}
                            className="h-8 w-8 text-muted-foreground hover:text-destructive mt-1"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
                <Button
                    variant="outline"
                    size="sm"
                    onClick={addItem}
                    className="w-full border-dashed text-muted-foreground hover:text-foreground"
                >
                    <Plus className="mr-2 h-4 w-4" /> Add Item
                </Button>
            </div>
        </div>
      );
  }

  // Fallback for unknown types or complex combinations
  return (
    <div className="flex flex-col gap-1.5 w-full">
       {name && renderLabel()}
       <div className="text-xs text-muted-foreground italic border border-dashed p-2 rounded">
          Complex type ({String(type)}) not fully supported in form view.
          Please use JSON mode.
       </div>
    </div>
  );
}
