/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect, useId } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, ChevronRight, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
import { Schema } from "./schema-viewer";

interface SchemaFormProps {
  schema: Schema;
  value: any; // eslint-disable-line @typescript-eslint/no-explicit-any
  onChange: (value: any) => void; // eslint-disable-line @typescript-eslint/no-explicit-any
  name?: string;
  depth?: number;
  required?: boolean;
}

export function SchemaForm({ schema, value, onChange, name, depth = 0, required = false }: SchemaFormProps) {
  // Initialize value if undefined based on type
  useEffect(() => {
    // Only initialize if value is strictly undefined.
    // Null might be a valid value for some schema types, but usually we want to respect the schema.
    if (value === undefined) {
      if (schema.default !== undefined) {
        onChange(schema.default);
      } else if (schema.type === "object") {
        onChange({});
      } else if (schema.type === "array") {
        onChange([]);
      }
      // We don't auto-init primitives to avoid over-writing parent state if it's controlled elsewhere or optional
    }
  }, []); // Run once on mount

  const handleChange = (newValue: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
    onChange(newValue);
  };

  const handleObjectChange = (key: string, val: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
    const newObj = { ...(value || {}), [key]: val };
    onChange(newObj);
  };

  const [isOpen, setIsOpen] = useState(true);
  const id = useId();

  if (!schema) return null;

  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Handle Enum
  if (schema.enum) {
      return (
        <div className={cn("grid gap-2", depth > 0 && "mb-4")}>
            {name && <Label htmlFor={id} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
            <Select value={String(value || "")} onValueChange={handleChange}>
                <SelectTrigger id={id} className="w-full">
                    <SelectValue placeholder="Select an option" />
                </SelectTrigger>
                <SelectContent>
                    {schema.enum.map((opt: any) => ( // eslint-disable-line @typescript-eslint/no-explicit-any
                        <SelectItem key={String(opt)} value={String(opt)}>
                            {String(opt)}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
            {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
        </div>
      );
  }

  // Handle String
  if (type === "string") {
      return (
        <div className={cn("grid gap-2", depth > 0 && "mb-4")}>
            {name && <Label htmlFor={id} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
            <Input
                id={id}
                value={value || ""}
                onChange={(e) => handleChange(e.target.value)}
                placeholder={schema.description || name}
            />
            {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
        </div>
      );
  }

  // Handle Number/Integer
  if (type === "number" || type === "integer") {
       return (
        <div className={cn("grid gap-2", depth > 0 && "mb-4")}>
            {name && <Label htmlFor={id} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
            <Input
                id={id}
                type="number"
                value={value || ""}
                onChange={(e) => handleChange(Number(e.target.value))}
                placeholder={schema.description || name}
            />
            {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
        </div>
      );
  }

  // Handle Boolean
  if (type === "boolean") {
       return (
        <div className={cn("flex items-center justify-between space-x-2 border p-3 rounded-md", depth > 0 && "mb-4")}>
            <div className="space-y-0.5">
                {name && <Label htmlFor={id} className={cn("text-base", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
                {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
            </div>
            <Switch
                id={id}
                checked={!!value}
                onCheckedChange={handleChange}
            />
        </div>
      );
  }

  // Handle Object
  if (type === "object" || schema.properties) {
      const properties = schema.properties || {};
      return (
          <div className={cn(depth > 0 ? "border-l-2 border-muted pl-4 mt-2" : "", "mb-4")}>
              {name && (
                  <div className="flex items-center gap-2 mb-2 cursor-pointer select-none" onClick={() => setIsOpen(!isOpen)}>
                      {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                      <Label className={cn("text-base font-semibold cursor-pointer")}>{name}</Label>
                      {required && <span className="text-red-500">*</span>}
                  </div>
              )}
               {isOpen && (
                   <div className="space-y-2">
                       {Object.entries(properties).map(([key, propSchema]) => (
                           <SchemaForm
                               key={key}
                               name={key}
                               schema={propSchema}
                               value={value?.[key]}
                               onChange={(val) => handleObjectChange(key, val)}
                               depth={depth + 1}
                               required={schema.required?.includes(key)}
                           />
                       ))}
                   </div>
               )}
          </div>
      );
  }

  // Handle Array
  if (type === "array" && schema.items) {
       const items = value || [];
       const addItem = () => {
           // Determine default value for new item
           const itemType = schema.items?.type;
           const newItem = itemType === "object" ? {} :
                           itemType === "array" ? [] :
                           itemType === "string" ? "" :
                           itemType === "number" || itemType === "integer" ? 0 :
                           itemType === "boolean" ? false : "";
           handleChange([...items, newItem]);
       };
       const removeItem = (index: number) => {
           const newItems = [...items];
           newItems.splice(index, 1);
           handleChange(newItems);
       };
       const updateItem = (index: number, val: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
           const newItems = [...items];
           newItems[index] = val;
           handleChange(newItems);
       };

       return (
           <div className={cn(depth > 0 ? "border-l-2 border-muted pl-4 mt-2" : "", "mb-4")}>
               <div className="flex items-center justify-between mb-2 select-none">
                   <div className="flex items-center gap-2" onClick={() => setIsOpen(!isOpen)}>
                       {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                       <Label className="text-base font-semibold cursor-pointer">{name || "List"}</Label>
                       {required && <span className="text-red-500">*</span>}
                       <span className="text-xs text-muted-foreground ml-2">({items.length} items)</span>
                   </div>
                   <Button variant="outline" size="sm" onClick={addItem} className="h-6 gap-1">
                       <Plus className="h-3 w-3" /> Add
                   </Button>
               </div>

               {isOpen && (
                   <div className="space-y-3">
                       {items.map((item: any, index: number) => ( // eslint-disable-line @typescript-eslint/no-explicit-any
                           <div key={index} className="flex gap-2 items-start group">
                               <div className="flex-1">
                                   <SchemaForm
                                       name={`Item ${index + 1}`}
                                       schema={schema.items as Schema}
                                       value={item}
                                       onChange={(val) => updateItem(index, val)}
                                       depth={depth + 1}
                                   />
                               </div>
                               <Button variant="ghost" size="icon" onClick={() => removeItem(index)} className="h-8 w-8 text-destructive opacity-50 group-hover:opacity-100">
                                   <Trash2 className="h-4 w-4" />
                               </Button>
                           </div>
                       ))}
                       {items.length === 0 && (
                           <div className="text-xs text-muted-foreground italic p-2 border border-dashed rounded text-center">
                               No items. Click Add to create one.
                           </div>
                       )}
                   </div>
               )}
           </div>
       );
  }

  // Fallback for unknown types
  return (
      <div className={cn("grid gap-2", depth > 0 && "mb-4")}>
          {name && <Label className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
          <div className="text-xs text-muted-foreground p-2 bg-muted/20 rounded">
              Unsupported type: {String(type)}
          </div>
      </div>
  );
}
