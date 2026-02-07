/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export interface Schema {
  type?: string;
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

/**
 * A dynamic form generator based on JSON Schema.
 */
export function SchemaForm({ schema, value, onChange, name, required, className }: SchemaFormProps) {
  // Initialize default value if undefined
  useEffect(() => {
    if (value === undefined && schema.default !== undefined) {
      onChange(schema.default);
    }
  }, [schema.default, value, onChange]);

  if (!schema) return null;

  const description = schema.description;
  const isRequired = required;

  // Helper to update object properties
  const handlePropertyChange = (prop: string, newValue: any) => {
    const newObj = { ...value } || {};
    if (newValue === undefined) {
      delete newObj[prop];
    } else {
      newObj[prop] = newValue;
    }
    onChange(newObj);
  };

  // Helper to update array items
  const handleArrayChange = (index: number, newValue: any) => {
    const newArray = [...(value || [])];
    newArray[index] = newValue;
    onChange(newArray);
  };

  const addArrayItem = () => {
    const newArray = [...(value || [])];
    // Try to determine a good initial value based on items schema
    let initialItemValue = undefined;
    if (schema.items) {
        if (schema.items.default !== undefined) initialItemValue = schema.items.default;
        else if (schema.items.type === 'string') initialItemValue = "";
        else if (schema.items.type === 'boolean') initialItemValue = false;
        else if (schema.items.type === 'object') initialItemValue = {};
        else if (schema.items.type === 'array') initialItemValue = [];
    }
    newArray.push(initialItemValue);
    onChange(newArray);
  };

  const removeArrayItem = (index: number) => {
    const newArray = [...(value || [])];
    newArray.splice(index, 1);
    onChange(newArray);
  };

  // Render Label with Tooltip
  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-1.5">
      {name && (
        <Label className={cn("text-sm font-medium", isRequired && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
          {name}
        </Label>
      )}
      {description && (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
          </TooltipTrigger>
          <TooltipContent className="max-w-[300px] text-xs">
            <p>{description}</p>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  // ENUM (Select)
  if (schema.enum) {
    return (
      <div className={className}>
        {renderLabel()}
        <Select
          value={value === undefined ? "" : String(value)}
          onValueChange={(val) => {
             // Try to convert back to original type if number/boolean
             let typedVal: any = val;
             if (schema.type === 'number' || schema.type === 'integer') typedVal = Number(val);
             if (schema.type === 'boolean') typedVal = val === 'true';
             onChange(typedVal);
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            {schema.enum.map((option: any) => (
              <SelectItem key={String(option)} value={String(option)}>
                {String(option)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    );
  }

  // BOOLEAN (Switch)
  if (schema.type === "boolean") {
    return (
      <div className={cn("flex items-center justify-between rounded-lg border p-3 bg-muted/20", className)}>
        <div className="space-y-0.5">
           {name && (
            <Label className={cn("text-sm font-medium", isRequired && "text-red-500")}>
              {name} {isRequired && "*"}
            </Label>
           )}
           {description && <p className="text-xs text-muted-foreground">{description}</p>}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
        />
      </div>
    );
  }

  // STRING (Input)
  if (schema.type === "string") {
    return (
      <div className={className}>
        {renderLabel()}
        <Input
          value={value || ""}
          onChange={(e) => onChange(e.target.value)}
          placeholder={schema.default ? String(schema.default) : ""}
        />
      </div>
    );
  }

  // NUMBER / INTEGER (Input type="number")
  if (schema.type === "number" || schema.type === "integer") {
    return (
      <div className={className}>
        {renderLabel()}
        <Input
          type="number"
          value={value === undefined ? "" : value}
          onChange={(e) => {
            const val = e.target.value;
            onChange(val === "" ? undefined : Number(val));
          }}
          placeholder={schema.default ? String(schema.default) : ""}
        />
      </div>
    );
  }

  // OBJECT (Card + Recursive)
  if (schema.type === "object" || schema.properties) {
    const properties = schema.properties || {};
    return (
      <Card className={cn("bg-muted/10", className)}>
         {(name || description) && (
            <CardHeader className="px-4 py-3 border-b bg-muted/20">
                {name && <CardTitle className="text-sm font-medium">{name}</CardTitle>}
                {description && <p className="text-xs text-muted-foreground">{description}</p>}
            </CardHeader>
         )}
        <CardContent className="p-4 space-y-4">
          {Object.entries(properties).map(([key, propSchema]) => (
            <SchemaForm
              key={key}
              name={key}
              schema={propSchema}
              value={value ? value[key] : undefined}
              onChange={(newVal) => handlePropertyChange(key, newVal)}
              required={schema.required?.includes(key)}
            />
          ))}
        </CardContent>
      </Card>
    );
  }

  // ARRAY (List + Recursive)
  if (schema.type === "array" && schema.items) {
    const items = value || [];
    return (
      <div className={cn("space-y-2", className)}>
        {renderLabel()}
        <div className="space-y-2">
            {items.map((item: any, index: number) => (
                <div key={index} className="flex items-start gap-2">
                    <div className="flex-1">
                        <SchemaForm
                            schema={schema.items!}
                            value={item}
                            onChange={(newVal) => handleArrayChange(index, newVal)}
                        />
                    </div>
                    <Button variant="ghost" size="icon" onClick={() => removeArrayItem(index)} className="h-8 w-8 text-destructive mt-8">
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
        </div>
        <Button variant="outline" size="sm" onClick={addArrayItem} className="w-full border-dashed">
            <Plus className="mr-2 h-4 w-4" /> Add Item
        </Button>
      </div>
    );
  }

  // Fallback for unknown types
  return (
    <div className={className}>
      {renderLabel()}
      <div className="p-2 border border-yellow-500/50 bg-yellow-500/10 rounded text-xs text-yellow-600 dark:text-yellow-400">
        Unsupported type: {schema.type}
      </div>
    </div>
  );
}
