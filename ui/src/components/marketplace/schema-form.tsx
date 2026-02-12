/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Plus, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface SchemaFormProps {
  schema: any;
  value: any;
  onChange: (value: any) => void;
  path?: string;
  className?: string;
}

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON schema.
 * Supports nested objects and arrays.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, path = "root", className }: SchemaFormProps) {
  if (!schema) return null;

  // Handle Object
  if (schema.type === "object" && schema.properties) {
    // Ensure value is an object
    const currentVal = (value && typeof value === 'object' && !Array.isArray(value)) ? value : {};

    return (
      <div className={cn("space-y-4", className)}>
        {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
          const isRequired = schema.required?.includes(key);
          const title = prop.title || key;
          const description = prop.description || "";
          const fieldPath = `${path}.${key}`;

          // Determine if we should show a label for this field
          // Objects usually don't need a label if they are nested, unless they have a title
          // But primitives definitely need one.
          const showLabel = true;

          return (
            <div key={key} className={cn("grid gap-2", prop.type === "object" ? "border-l-2 border-primary/20 pl-4 ml-1 my-2" : "")}>
               {showLabel && (
                  <Label htmlFor={fieldPath} className="flex items-center gap-1 font-medium">
                    {title} {isRequired && <span className="text-destructive">*</span>}
                  </Label>
               )}

               {description && (prop.type === "object" || prop.type === "array") && (
                   <p className="text-xs text-muted-foreground mb-1">{description}</p>
               )}

               <SchemaForm
                 schema={prop}
                 value={currentVal[key]}
                 onChange={(v) => onChange({ ...currentVal, [key]: v })}
                 path={fieldPath}
               />

               {description && prop.type !== "object" && prop.type !== "array" && prop.type !== "boolean" && (
                  <p className="text-xs text-muted-foreground">{description}</p>
               )}
            </div>
          );
        })}
      </div>
    );
  }

  // Handle Array
  if (schema.type === "array" && schema.items) {
      const items = Array.isArray(value) ? value : [];

      const addItem = () => {
          // Initialize new item based on item schema type
          let newItem: any = "";
          if (schema.items.type === "object") newItem = {};
          if (schema.items.type === "array") newItem = [];
          if (schema.items.type === "boolean") newItem = false;
          if (schema.items.default !== undefined) newItem = schema.items.default;

          onChange([...items, newItem]);
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
          <div className="space-y-2">
              {items.map((item: any, index: number) => (
                  <div key={index} className="flex gap-2 items-start p-3 border rounded-md bg-muted/20 group relative">
                      <div className="flex-1">
                          <SchemaForm
                            schema={schema.items}
                            value={item}
                            onChange={(v) => updateItem(index, v)}
                            path={`${path}[${index}]`}
                          />
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 text-muted-foreground hover:text-destructive absolute top-2 right-2"
                        onClick={() => removeItem(index)}
                        title="Remove Item"
                      >
                          <Trash2 className="h-4 w-4" />
                      </Button>
                  </div>
              ))}
              <Button variant="outline" size="sm" onClick={addItem} className="w-full border-dashed">
                  <Plus className="mr-2 h-3 w-3" /> Add Item
              </Button>
          </div>
      );
  }

  // Handle Primitives

  // Determine input type
  let inputType = "text";
  if (schema.format === "password") {
      inputType = "password";
  } else if (
      path.toLowerCase().includes("token") ||
      path.toLowerCase().includes("secret") ||
      path.toLowerCase().includes("key") ||
      path.toLowerCase().includes("password")
  ) {
      inputType = "password";
  } else if (schema.type === "integer" || schema.type === "number") {
      inputType = "number";
  }

  if (schema.type === "boolean") {
       return (
           <div className="flex items-center space-x-2 h-9">
              <Checkbox
                  id={path}
                  checked={value === true || value === "true"}
                  onCheckedChange={(c) => onChange(c === true)}
              />
              <label
                  htmlFor={path}
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                  {schema.description || "Enable"}
              </label>
           </div>
       );
  }

  if (schema.enum) {
      return (
          <Select value={String(value || "")} onValueChange={(v) => onChange(v)}>
              <SelectTrigger id={path}>
                  <SelectValue placeholder="Select..." />
              </SelectTrigger>
              <SelectContent>
                  {schema.enum.map((opt: string) => (
                      <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                  ))}
              </SelectContent>
          </Select>
      );
  }

  // Default Text/Number Input
  return (
      <Input
          id={path}
          value={value !== undefined && value !== null ? value : ""}
          onChange={(e) => {
              const val = e.target.value;
              if (schema.type === "integer") {
                  onChange(parseInt(val, 10));
              } else if (schema.type === "number") {
                  onChange(parseFloat(val));
              } else {
                  onChange(val);
              }
          }}
          placeholder={schema.default !== undefined ? String(schema.default) : `Enter value`}
          type={inputType}
      />
  );
}
