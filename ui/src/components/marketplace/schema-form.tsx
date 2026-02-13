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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface SchemaFormProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  schema: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any; // Can be object, string, boolean, array, etc.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onChange: (value: any) => void;
  title?: string;
  isRoot?: boolean;
  fieldId?: string;
}

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON Schema.
 *
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @param props.title - Optional title override.
 * @param props.isRoot - Whether this is the root form (avoids outer card).
 * @param props.fieldId - Optional ID for the input field (for labels).
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, title, isRoot = true, fieldId }: SchemaFormProps) {
  if (!schema) return null;

  // Handle Enum (takes precedence over primitive types)
  if (schema.enum) {
       const description = schema.description || "";
       return (
            <div className="grid gap-1">
                <Select value={String(value || "")} onValueChange={onChange}>
                    <SelectTrigger>
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
       );
  }

  // Handle Object
  if (schema.type === "object" || (schema.properties && !schema.type)) {
      const properties = schema.properties || {};
      const currentValue = value || {};

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const handleChange = (key: string, val: any) => {
        onChange({ ...currentValue, [key]: val });
      };

      const content = (
          <div className="space-y-4">
            {Object.entries(properties).map(([key, prop]: [string, any]) => {
              const isRequired = schema.required?.includes(key);
              const propTitle = prop.title || key;

              return (
                <div key={key} className="space-y-2">
                    {/* Only show label if it's not a complex nested object which has its own header, OR if we want to show it anyway */}
                    {prop.type !== "object" && prop.type !== "array" && (
                         <Label htmlFor={key} className="flex items-center gap-1">
                            {propTitle} {isRequired && <span className="text-destructive">*</span>}
                         </Label>
                    )}

                    <SchemaForm
                        schema={prop}
                        value={currentValue[key]}
                        onChange={(v) => handleChange(key, v)}
                        title={propTitle}
                        isRoot={false}
                        fieldId={key}
                    />
                </div>
              );
            })}
          </div>
      );

      if (isRoot) {
          return content;
      }

      return (
          <Card className="border-dashed">
             <CardHeader className="py-3 px-4 bg-muted/20">
                 <CardTitle className="text-sm font-medium">{title || schema.title || "Configuration"}</CardTitle>
             </CardHeader>
             <CardContent className="p-4">
                 {content}
             </CardContent>
          </Card>
      );
  }

  // Handle Array
  if (schema.type === "array") {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const items = Array.isArray(value) ? value : [];
      const itemSchema = schema.items;

      const addItem = () => {
          // Initialize primitive types with empty defaults if possible
          let initial = undefined;
          if (itemSchema.type === "string") initial = "";
          if (itemSchema.type === "object") initial = {};
          if (itemSchema.type === "array") initial = [];

          onChange([...items, initial]);
      };

      const removeItem = (index: number) => {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const newItems = [...items];
          newItems.splice(index, 1);
          onChange(newItems);
      };

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const updateItem = (index: number, val: any) => {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const newItems = [...items];
          newItems[index] = val;
          onChange(newItems);
      };

      return (
          <Card className="border-dashed">
              <CardHeader className="py-3 px-4 bg-muted/20 flex flex-row items-center justify-between space-y-0">
                  <div className="flex flex-col gap-0.5">
                    <CardTitle className="text-sm font-medium">{title || schema.title || "List"}</CardTitle>
                    {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
                  </div>
                  <Button variant="outline" size="sm" onClick={addItem} type="button">
                      <Plus className="mr-2 h-3 w-3" /> Add Item
                  </Button>
              </CardHeader>
              <CardContent className="p-4 space-y-4">
                  {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                  {items.map((item: any, index: number) => (
                      <div key={index} className="flex items-start gap-2">
                          <div className="flex-1">
                              <SchemaForm
                                  schema={itemSchema}
                                  value={item}
                                  onChange={(v) => updateItem(index, v)}
                                  isRoot={false}
                                  title={`Item ${index + 1}`}
                              />
                          </div>
                          <Button variant="ghost" size="icon" onClick={() => removeItem(index)} className="text-destructive mt-1">
                              <Trash2 className="h-4 w-4" />
                          </Button>
                      </div>
                  ))}
                  {items.length === 0 && (
                      <div className="text-sm text-muted-foreground text-center py-2 italic">
                          No items.
                      </div>
                  )}
              </CardContent>
          </Card>
      );
  }

  // Handle Primitives
  const description = schema.description || "";

  // String / Password
  if (schema.type === "string" || !schema.type) { // Default to string
       let inputType = "text";
       if (schema.format === "password") inputType = "password";
       if (schema.title?.toLowerCase().includes("password") || schema.title?.toLowerCase().includes("token") || schema.title?.toLowerCase().includes("key")) inputType = "password";

       return (
           <div className="grid gap-1">
               <Input
                    id={fieldId}
                    value={value || ""}
                    onChange={(e) => onChange(e.target.value)}
                    placeholder={schema.default || description || ""}
                    type={inputType}
                />
                {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
           </div>
       );
  }

  // Boolean
  if (schema.type === "boolean") {
      return (
        <div className="flex items-center space-x-2">
            <Checkbox
                id={fieldId}
                checked={value === true || value === "true"} // Handle string "true" if coming from older config
                onCheckedChange={(c) => onChange(c)}
            />
            <label htmlFor={fieldId} className="text-sm font-medium leading-none cursor-pointer">
                {description || "Enable"}
            </label>
        </div>
      );
  }

  // Integer / Number
  if (schema.type === "integer" || schema.type === "number") {
      return (
          <div className="grid gap-1">
               <Input
                    id={fieldId}
                    type="number"
                    value={value || ""}
                    onChange={(e) => onChange(Number(e.target.value))}
                    placeholder={String(schema.default || "")}
                />
                {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
           </div>
      );
  }

  return <div className="text-red-500 text-xs">Unsupported type: {schema.type}</div>;
}
