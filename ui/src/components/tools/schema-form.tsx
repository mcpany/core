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
import { Plus, Trash2, Info } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

/**
 * Schema definition (subset of JSON Schema).
 */
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
  depth?: number;
}

const MAX_DEPTH = 10;

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON Schema.
 *
 * @param props - The component props.
 * @returns The rendered form.
 */
export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0 }: SchemaFormProps) {
  if (!schema || depth > MAX_DEPTH) return null;

  // Helper to handle field changes
  const handleChange = (newValue: any) => {
    onChange(newValue);
  };

  const label = (
    <div className="flex items-center gap-2 mb-2">
      <Label className={cn(depth === 0 ? "text-base font-semibold" : "text-sm font-medium")}>
        {name || "Root"}
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
  if (schema.enum && Array.isArray(schema.enum)) {
    // Ensure value is a string for Select
    const currentValue = value === undefined || value === null ? "" : String(value);

    return (
      <div className="space-y-1">
        {name && label}
        <Select
          value={currentValue}
          onValueChange={(val) => {
              // Try to cast back to original type if number/boolean
              if (schema.type === 'number' || schema.type === 'integer') {
                  handleChange(Number(val));
              } else if (schema.type === 'boolean') {
                  handleChange(val === 'true');
              } else {
                  handleChange(val);
              }
          }}
        >
          <SelectTrigger className="w-full">
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
  if (schema.type === "boolean") {
    return (
      <div className="flex items-center justify-between space-y-0 rounded-md border p-3">
        <div className="space-y-0.5">
           <Label className="text-sm font-medium">
             {name || "Enabled"}
             {required && <span className="text-red-500 ml-1">*</span>}
           </Label>
           {schema.description && (
             <p className="text-[0.8rem] text-muted-foreground">
               {schema.description}
             </p>
           )}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={handleChange}
        />
      </div>
    );
  }

  // 3. Object (Recursive)
  if (schema.type === "object" || (schema.properties && !schema.type)) {
    const properties = schema.properties || {};
    const requiredFields = schema.required || [];

    // Ensure value is object
    const currentValue = typeof value === 'object' && value !== null ? value : {};

    const handlePropChange = (propKey: string, propValue: any) => {
      const newValue = { ...currentValue, [propKey]: propValue };
      handleChange(newValue);
    };

    const content = (
        <div className="space-y-4">
            {Object.entries(properties).map(([key, propSchema]) => (
            <SchemaForm
                key={key}
                name={key}
                schema={propSchema}
                value={currentValue[key]}
                onChange={(val) => handlePropChange(key, val)}
                required={requiredFields.includes(key)}
                depth={depth + 1}
            />
            ))}
        </div>
    );

    if (depth === 0) {
        return content;
    }

    return (
      <Card className="bg-muted/10 border-dashed">
        <CardHeader className="p-3 pb-0">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
              {name}
              {required && <span className="text-red-500">*</span>}
          </CardTitle>
          {schema.description && <p className="text-xs text-muted-foreground">{schema.description}</p>}
        </CardHeader>
        <CardContent className="p-3 pt-2">
            {content}
        </CardContent>
      </Card>
    );
  }

  // 4. Array (Dynamic List)
  if (schema.type === "array" && schema.items) {
      const items = Array.isArray(value) ? value : [];

      const addItem = () => {
          // Initialize new item with default or empty
          const newItem = schema.items?.default !== undefined ? schema.items.default :
                          schema.items?.type === 'object' ? {} :
                          schema.items?.type === 'array' ? [] :
                          schema.items?.type === 'boolean' ? false :
                          schema.items?.type === 'number' ? 0 :
                          "";
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
          <div className="space-y-2">
              <div className="flex items-center justify-between">
                  {label}
                  <Button variant="outline" size="sm" onClick={addItem} type="button">
                      <Plus className="h-3 w-3 mr-1" /> Add
                  </Button>
              </div>

              <div className="space-y-2 pl-2 border-l-2 border-muted">
                  {items.map((item: any, idx: number) => (
                      <div key={idx} className="flex gap-2 items-start">
                          <div className="flex-1">
                              <SchemaForm
                                  schema={schema.items!} // non-null assertion as we checked schema.items
                                  value={item}
                                  onChange={(val) => updateItem(idx, val)}
                                  depth={depth + 1}
                              />
                          </div>
                          <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-muted-foreground hover:text-destructive mt-1"
                              onClick={() => removeItem(idx)}
                              title="Remove item"
                          >
                              <Trash2 className="h-4 w-4" />
                          </Button>
                      </div>
                  ))}
                  {items.length === 0 && (
                      <div className="text-xs text-muted-foreground italic p-2 border border-dashed rounded text-center">
                          No items
                      </div>
                  )}
              </div>
          </div>
      );
  }

  // 5. Number/Integer
  if (schema.type === "number" || schema.type === "integer") {
    return (
      <div className="space-y-1">
        {name && label}
        <Input
          type="number"
          value={value === undefined ? "" : value}
          onChange={(e) => {
              const val = e.target.value;
              if (val === "") {
                  handleChange(undefined);
              } else {
                  handleChange(Number(val));
              }
          }}
          placeholder={schema.default !== undefined ? String(schema.default) : "0"}
        />
      </div>
    );
  }

  // 6. String (Default)
  return (
    <div className="space-y-1">
      {name && label}
      <Input
        type="text"
        value={value || ""}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={schema.default !== undefined ? String(schema.default) : ""}
      />
    </div>
  );
}
