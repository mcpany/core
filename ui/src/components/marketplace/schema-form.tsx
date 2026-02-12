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
  showLabel?: boolean;
  required?: boolean;
}

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON Schema.
 * Supports objects, arrays, strings, numbers, booleans, and enums.
 *
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @param props.path - Unique path for ID generation (e.g. "root.server.host").
 * @param props.showLabel - Whether to render the label for this field.
 * @param props.required - Whether this field is marked as required by the parent.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, path = "root", showLabel = true, required = false }: SchemaFormProps) {
  if (!schema) return null;

  // Initialize value if undefined
  const safeValue = value ?? (
      schema.type === "array" ? [] :
      schema.type === "object" ? {} :
      schema.default ?? ""
  );

  const title = schema.title || (path.includes('.') ? path.split('.').pop() : "");
  const description = schema.description;

  // OBJECT
  if (schema.type === "object" && schema.properties) {
    return (
      <div className={cn("space-y-2", path !== "root" && "border-l-2 border-muted pl-4 mt-2")}>
        {showLabel && title && (
           <Label className="font-semibold text-base mb-2 block">{title}</Label>
        )}
        {description && <p className="text-xs text-muted-foreground mb-3">{description}</p>}

        <div className="space-y-4">
            {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
            const isFieldRequired = schema.required?.includes(key);
            return (
                <SchemaForm
                    key={key}
                    schema={prop}
                    value={safeValue[key]}
                    onChange={(v) => onChange({ ...safeValue, [key]: v })}
                    path={`${path}.${key}`}
                    required={isFieldRequired}
                />
            );
            })}
        </div>
      </div>
    );
  }

  // ARRAY
  if (schema.type === "array") {
      const items = Array.isArray(safeValue) ? safeValue : [];
      return (
          <div className={cn("space-y-2", path !== "root" && "ml-1")}>
              {showLabel && (
                  <Label className="font-semibold">{title}</Label>
              )}
              {description && <p className="text-xs text-muted-foreground">{description}</p>}

              <div className="space-y-2">
                {items.map((item: any, index: number) => (
                    <div key={index} className="flex items-start gap-2 group relative">
                        <div className="flex-1 bg-muted/10 rounded-md p-2 border border-transparent hover:border-border transition-colors">
                            <SchemaForm
                                schema={schema.items}
                                value={item}
                                onChange={(v) => {
                                    const newItems = [...items];
                                    newItems[index] = v;
                                    onChange(newItems);
                                }}
                                path={`${path}.${index}`}
                                showLabel={schema.items.type === "object"} // Only show labels for objects in array
                            />
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-muted-foreground hover:text-destructive shrink-0 mt-1"
                            onClick={() => {
                                const newItems = items.filter((_: any, i: number) => i !== index);
                                onChange(newItems);
                            }}
                            title="Remove Item"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
              </div>
              <Button
                variant="outline"
                size="sm"
                className="w-full border-dashed text-xs h-8"
                onClick={() => {
                     const newItem = schema.items.type === "object" ? {} :
                                     schema.items.type === "array" ? [] : "";
                     onChange([...items, newItem]);
                }}
              >
                  <Plus className="mr-2 h-3 w-3" /> Add {schema.items.title || "Item"}
              </Button>
          </div>
      )
  }

  // PRIMITIVES

  // Boolean
  if (schema.type === "boolean") {
       return (
        <div className="flex items-center space-x-2 py-1">
            <Checkbox
                id={path}
                checked={value === true || value === "true"}
                onCheckedChange={(c) => onChange(c)}
            />
            <Label htmlFor={path} className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer">
                {title || "Enabled"}
            </Label>
        </div>
       )
  }

  // Enum
  if (schema.enum) {
      return (
        <div className="space-y-1.5">
            {showLabel && (
                <Label htmlFor={path} className="flex items-center gap-1">
                    {title} {required && <span className="text-destructive">*</span>}
                </Label>
            )}
            <Select value={value} onValueChange={onChange}>
                <SelectTrigger id={path}>
                    <SelectValue placeholder="Select..." />
                </SelectTrigger>
                <SelectContent>
                    {schema.enum.map((opt: string) => (
                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                    ))}
                </SelectContent>
            </Select>
            {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
        </div>
      )
  }

  // String / Number / Password
  let inputType = "text";
  if (schema.type === "integer" || schema.type === "number") inputType = "number";
  if (schema.format === "password" || (title && title.toLowerCase().includes("token"))) inputType = "password";

  return (
    <div className="space-y-1.5">
        {showLabel && (
            <Label htmlFor={path} className="flex items-center gap-1">
                {title} {required && <span className="text-destructive">*</span>}
            </Label>
        )}
        <Input
            id={path}
            value={value || ""}
            onChange={(e) => {
                const val = e.target.value;
                onChange(inputType === "number" ? Number(val) : val);
            }}
            placeholder={schema.default ? String(schema.default) : ""}
            type={inputType}
            className="bg-background"
        />
        {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
    </div>
  );
}
