/**
 * Copyright 2025 Author(s) of MCP Any
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
import { Trash2, Plus, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

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
  className?: string;
}

export function SchemaForm({
  schema,
  value,
  onChange,
  name,
  required = false,
  className,
}: SchemaFormProps) {
  if (!schema) return null;

  // Handle "type": ["string", "null"] case
  let type = schema.type;
  if (Array.isArray(type)) {
    type = type.find((t) => t !== "null");
  }

  const label = name ? (
    <div className="flex items-center gap-2 mb-1.5">
      <Label className={cn(required && "text-foreground font-semibold")}>
        {name}
        {required && <span className="text-destructive ml-1">*</span>}
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
  ) : null;

  // String / Number / Integer
  if (type === "string" || type === "number" || type === "integer") {
    // Enum
    if (schema.enum) {
      return (
        <div className={cn("grid gap-1.5", className)}>
          {label}
          <Select
            value={value === undefined ? "" : String(value)}
            onValueChange={(val) => {
               if (type === "number" || type === "integer") {
                   const num = parseFloat(val);
                   onChange(isNaN(num) ? undefined : num);
               } else {
                   onChange(val);
               }
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select..." />
            </SelectTrigger>
            <SelectContent>
              {schema.enum.map((opt) => (
                <SelectItem key={String(opt)} value={String(opt)}>
                  {String(opt)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      );
    }

    // Default Input
    return (
      <div className={cn("grid gap-1.5", className)}>
        {label}
        <Input
          type={type === "number" || type === "integer" ? "number" : "text"}
          value={value === undefined ? "" : value}
          onChange={(e) => {
            const val = e.target.value;
            if (type === "number" || type === "integer") {
                const num = parseFloat(val);
                onChange(isNaN(num) && val === "" ? undefined : isNaN(num) ? val : num);
            } else {
                onChange(val);
            }
          }}
          placeholder={schema.default ? `Default: ${schema.default}` : ""}
        />
      </div>
    );
  }

  // Boolean
  if (type === "boolean") {
    return (
      <div className={cn("flex items-center justify-between rounded-lg border p-3", className)}>
        <div className="space-y-0.5">
          <Label className="text-base">
            {name}
            {required && <span className="text-destructive ml-1">*</span>}
          </Label>
          {schema.description && (
             <p className="text-xs text-muted-foreground">{schema.description}</p>
          )}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
        />
      </div>
    );
  }

  // Object
  if (type === "object" || schema.properties) {
    const properties = schema.properties || {};
    const currentValue = value || {};

    const handlePropChange = (key: string, newVal: any) => {
      const updated = { ...currentValue, [key]: newVal };
      // Optional: Clean up undefined values?
      // For now, keep them to allow clearing inputs
      onChange(updated);
    };

    return (
      <Card className={cn("w-full bg-muted/20", className)}>
         {name && (
            <div className="p-3 border-b bg-muted/30 flex items-center gap-2">
                <span className="font-medium text-sm">{name}</span>
                {required && <span className="text-destructive text-xs">*</span>}
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
        <CardContent className="p-4 space-y-4">
          {Object.entries(properties).map(([key, propSchema]) => (
            <SchemaForm
              key={key}
              name={key}
              schema={propSchema as Schema}
              value={currentValue[key]}
              onChange={(val) => handlePropChange(key, val)}
              required={schema.required?.includes(key)}
            />
          ))}
        </CardContent>
      </Card>
    );
  }

  // Array
  if (type === "array" && schema.items) {
      const items = (value || []) as any[];
      const itemSchema = schema.items as Schema;

      const addItem = () => {
          onChange([...items, undefined]);
      };

      const removeItem = (index: number) => {
          const newItems = [...items];
          newItems.splice(index, 1);
          onChange(newItems);
      };

      const updateItem = (index: number, val: any) => {
          const newItems = [...items];
          newItems[index] = val;
          onChange(newItems);
      };

      return (
          <Card className={cn("w-full bg-muted/20", className)}>
               <div className="p-3 border-b bg-muted/30 flex items-center justify-between">
                   <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{name || "List"}</span>
                        {required && <span className="text-destructive text-xs">*</span>}
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
                   <Button variant="ghost" size="sm" onClick={addItem} className="h-6 w-6 p-0">
                       <Plus className="h-4 w-4" />
                   </Button>
               </div>
               <CardContent className="p-4 space-y-3">
                   {items.map((item, index) => (
                       <div key={index} className="flex gap-2 items-start">
                           <div className="flex-1">
                                <SchemaForm
                                    schema={itemSchema}
                                    value={item}
                                    onChange={(val) => updateItem(index, val)}
                                    className="mb-0" // Remove margin bottom for nested item
                                />
                           </div>
                           <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => removeItem(index)}
                                className="h-8 w-8 text-destructive hover:bg-destructive/10 mt-1" // Align with input
                           >
                               <Trash2 className="h-4 w-4" />
                           </Button>
                       </div>
                   ))}
                   {items.length === 0 && (
                       <div className="text-xs text-muted-foreground text-center py-2 italic">
                           No items. Click + to add.
                       </div>
                   )}
               </CardContent>
          </Card>
      );
  }

  // Fallback
  return (
      <div className="text-xs text-red-500">
          Unsupported type: {String(type)}
      </div>
  );
}
