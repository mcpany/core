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
  depth?: number;
  id?: string;
}

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON Schema.
 * Supports nested objects and arrays.
 *
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @param props.depth - Current recursion depth (internal use).
 * @param props.id - The ID to use for the input element (for label association).
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, depth = 0, id }: SchemaFormProps) {
  if (!schema) return null;

  // Handle Object (Properties)
  if (schema.type === "object" && schema.properties) {
     const objectValue = value || {};
     return (
        <div className={cn("space-y-4", depth > 0 && "pl-3 border-l-2 border-muted/50")}>
            {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
                const isRequired = schema.required?.includes(key);
                const title = prop.title || key;
                const description = prop.description || "";
                const isComplex = prop.type === "object" || prop.type === "array";

                return (
                    <div key={key} className={cn("space-y-2", isComplex && "pt-2")}>
                        {/* Label */}
                        <Label htmlFor={key} className={cn("flex items-center gap-1", isComplex && "font-semibold text-base")}>
                            {title} {isRequired && <span className="text-destructive">*</span>}
                        </Label>

                        {/* Description for complex types (before content) */}
                        {isComplex && description && (
                             <p className="text-xs text-muted-foreground mb-2">{description}</p>
                        )}

                        {/* Recursive Render */}
                        <SchemaForm
                            schema={prop}
                            value={objectValue[key]}
                            onChange={(v) => onChange({ ...objectValue, [key]: v })}
                            depth={depth + 1}
                            id={key}
                        />

                         {/* Description for simple types (after content) */}
                         {!isComplex && description && (
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
      const arrayValue = Array.isArray(value) ? value : [];

      const addItem = () => {
          let newItem: any = "";
          const itemType = schema.items.type;

          if (itemType === "object") newItem = {};
          else if (itemType === "array") newItem = [];
          else if (itemType === "boolean") newItem = false;
          else if (itemType === "integer" || itemType === "number") newItem = 0;

          onChange([...arrayValue, newItem]);
      };

      const removeItem = (index: number) => {
          const newArray = [...arrayValue];
          newArray.splice(index, 1);
          onChange(newArray);
      };

      const updateItem = (index: number, val: any) => {
          const newArray = [...arrayValue];
          newArray[index] = val;
          onChange(newArray);
      };

      return (
          <div className="space-y-3">
              {arrayValue.length === 0 && (
                  <div className="text-sm text-muted-foreground italic py-2">No items.</div>
              )}
              {arrayValue.map((item: any, index: number) => (
                  <div key={index} className="flex items-start gap-2 group relative">
                      <div className="flex-1 bg-muted/10 p-3 rounded-md border">
                          <SchemaForm
                              schema={schema.items}
                              value={item}
                              onChange={(v) => updateItem(index, v)}
                              depth={depth + 1}
                              // Array items don't have unique IDs easily mapped to a label unless we generate one.
                              // But array items usually don't have a label above them, they are just items.
                          />
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => removeItem(index)}
                        className="text-destructive hover:bg-destructive/10 h-8 w-8 absolute top-1 right-1"
                        title="Remove Item"
                      >
                          <Trash2 className="h-4 w-4" />
                      </Button>
                  </div>
              ))}
              <Button variant="outline" size="sm" onClick={addItem} className="w-full border-dashed text-muted-foreground">
                  <Plus className="mr-2 h-4 w-4" /> Add Item
              </Button>
          </div>
      );
  }

  // --- Primitives ---

  // Determine input type
  let inputType = "text";
  if (schema.format === "password") {
    inputType = "password";
  } else if (
      // Heuristic for sensitive fields if format is not explicit
      id && (
        id.toLowerCase().includes("token") ||
        id.toLowerCase().includes("secret") ||
        id.toLowerCase().includes("key") ||
        id.toLowerCase().includes("password")
      )
  ) {
    inputType = "password";
  } else if (schema.type === "integer" || schema.type === "number") {
    inputType = "number";
  }

  // Boolean
  if (schema.type === "boolean") {
    return (
        <div className="flex items-center space-x-2">
            <Checkbox
                id={id || `check-${depth}-${Math.random()}`}
                checked={value === true || value === "true"}
                onCheckedChange={(c) => onChange(c === true)}
            />
            <span className="text-sm text-muted-foreground">{value ? "Enabled" : "Disabled"}</span>
        </div>
    );
  }

  // Enum
  if (schema.enum) {
    return (
        <Select value={String(value || "")} onValueChange={(v) => onChange(v)}>
            <SelectTrigger id={id}>
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

  // String / Number
  return (
    <Input
        id={id}
        value={value ?? ""}
        onChange={(e) => {
            const val = e.target.value;
            if (schema.type === "integer" || schema.type === "number") {
                 const num = parseFloat(val);
                 onChange(isNaN(num) ? val : num);
            } else {
                 onChange(val);
            }
        }}
        placeholder={schema.default ? String(schema.default) : ""}
        type={inputType}
    />
  );
}
