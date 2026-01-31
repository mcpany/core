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
import { Plus, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface DynamicFormProps {
  schema: any;
  value: any;
  onChange: (value: any) => void;
  className?: string;
  parentPath?: string;
}

export function DynamicForm({
  schema,
  value,
  onChange,
  className,
  parentPath = "",
}: DynamicFormProps) {
  if (!schema) return null;

  // Handle Object
  if (schema.type === "object" || (schema.properties && !schema.type)) {
    return (
      <div className={cn("space-y-4", className)}>
        {Object.entries(schema.properties || {}).map(([key, propSchema]: [string, any]) => {
          const fieldPath = parentPath ? `${parentPath}.${key}` : key;
          const isRequired = schema.required?.includes(key);
          const currentValue = value ? value[key] : undefined;

          return (
            <div key={key} className="space-y-2">
              <Label className={cn("text-xs font-medium", isRequired && "text-foreground")}>
                {key} {isRequired && <span className="text-red-500">*</span>}
              </Label>
              {propSchema.description && (
                <p className="text-[10px] text-muted-foreground mb-1">
                  {propSchema.description}
                </p>
              )}
              <DynamicForm
                schema={propSchema}
                value={currentValue}
                onChange={(newValue) => {
                  const updated = { ...(value || {}) };
                  if (newValue === undefined) {
                      // Optionally remove key if undefined?
                      // For now, keep it simple.
                      delete updated[key];
                  } else {
                      updated[key] = newValue;
                  }
                  onChange(updated);
                }}
                parentPath={fieldPath}
              />
            </div>
          );
        })}
      </div>
    );
  }

  // Handle Enum
  if (schema.enum) {
    return (
      <Select
        value={value ? String(value) : ""}
        onValueChange={(val) => {
           // Try to infer type
           if (schema.type === 'integer' || schema.type === 'number') {
               onChange(Number(val));
           } else {
               onChange(val);
           }
        }}
      >
        <SelectTrigger className="h-8">
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
    );
  }

  // Handle Boolean
  if (schema.type === "boolean") {
    return (
      <div className="flex items-center space-x-2">
        <Switch
          checked={!!value}
          onCheckedChange={(checked) => onChange(checked)}
        />
        <span className="text-xs text-muted-foreground">{value ? "True" : "False"}</span>
      </div>
    );
  }

  // Handle Array
  if (schema.type === "array") {
     const items = Array.isArray(value) ? value : [];
     const itemSchema = schema.items;

     const addItem = () => {
         onChange([...items, undefined]); // Initialize with undefined, child form will set default
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
         <div className="space-y-2 border rounded p-2 bg-muted/10">
             {items.map((item: any, index: number) => (
                 <div key={index} className="flex items-start gap-2">
                     <div className="flex-1">
                        <DynamicForm
                            schema={itemSchema}
                            value={item}
                            onChange={(val) => updateItem(index, val)}
                            parentPath={`${parentPath}[${index}]`}
                        />
                     </div>
                     <Button variant="ghost" size="icon" onClick={() => removeItem(index)} className="h-8 w-8 text-destructive">
                         <X className="h-4 w-4" />
                     </Button>
                 </div>
             ))}
             <Button variant="outline" size="sm" onClick={addItem} className="w-full border-dashed h-8 text-xs">
                 <Plus className="mr-2 h-3 w-3" /> Add Item
             </Button>
         </div>
     )
  }

  // Handle Number/Integer
  if (schema.type === "number" || schema.type === "integer") {
    return (
      <Input
        type="number"
        className="h-8 text-sm"
        value={value ?? ""}
        onChange={(e) => {
            const val = e.target.value;
            if (val === "") onChange(undefined);
            else onChange(Number(val));
        }}
        placeholder="Enter number..."
      />
    );
  }

  // Handle String (Default)
  return (
    <Input
      type="text"
      className="h-8 text-sm"
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value)}
      placeholder="Enter text..."
    />
  );
}
