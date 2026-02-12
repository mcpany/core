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
  schema: any;
  value: any;
  onChange: (value: any) => void;
  level?: number;
  path?: string;
  title?: string;
  required?: boolean;
}

/**
 * SchemaForm component for rendering JSON Schema forms recursively.
 * Supports objects, arrays, and primitive types.
 */
export function SchemaForm({ schema, value, onChange, level = 0, path = "", title, required }: SchemaFormProps) {
  if (!schema) return null;

  // Ensure value is initialized correctly based on type
  React.useEffect(() => {
    if (value === undefined) {
      if (schema.type === "object" && !schema.default) {
        onChange({});
      } else if (schema.type === "array" && !schema.default) {
        onChange([]);
      } else if (schema.default !== undefined) {
        onChange(schema.default);
      }
    }
  }, [schema, value, onChange]);

  const displayTitle = title || schema.title || path.split('.').pop() || "Field";

  // Handle Primitive Types
  if (schema.type === "string" || schema.type === "integer" || schema.type === "number" || schema.type === "boolean") {
    return (
      <SchemaField
        name={displayTitle}
        schema={schema}
        value={value}
        onChange={onChange}
        required={!!required}
      />
    );
  }

  // Handle Object Type
  if (schema.type === "object" && schema.properties) {
    const currentVal = value || {};

    return (
      <div className={`space-y-4 ${level > 0 ? "border-l-2 border-muted pl-4 ml-1" : ""}`}>
        {level > 0 && (
            <div className="space-y-1">
                <Label className="flex items-center gap-1 text-xs uppercase text-muted-foreground font-mono">
                    {displayTitle} {required && <span className="text-destructive">*</span>}
                </Label>
                {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
            </div>
        )}
        {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
          const isRequired = schema.required?.includes(key);
          const propValue = currentVal[key];

          const handlePropChange = (newVal: any) => {
             const updated = { ...currentVal, [key]: newVal };
             onChange(updated);
          };

          return (
            <div key={key} className="space-y-2">
               <SchemaForm
                schema={prop}
                value={propValue}
                onChange={handlePropChange}
                level={level + 1}
                path={`${path}.${key}`}
                title={prop.title || key}
                required={isRequired}
              />
            </div>
          );
        })}
      </div>
    );
  }

  // Handle Array Type
  if (schema.type === "array" && schema.items) {
    const currentList = Array.isArray(value) ? value : [];

    const handleAddItem = () => {
        // Initialize new item based on items schema
        const newItem = schema.items.default !== undefined
            ? schema.items.default
            : schema.items.type === "object" ? {} : "";
        onChange([...currentList, newItem]);
    };

    const handleRemoveItem = (index: number) => {
        const newList = [...currentList];
        newList.splice(index, 1);
        onChange(newList);
    };

    const handleItemChange = (index: number, val: any) => {
        const newList = [...currentList];
        newList[index] = val;
        onChange(newList);
    };

    return (
      <Card className="bg-muted/10 border-dashed">
          <CardHeader className="p-3 pb-0 flex flex-row items-center justify-between">
              <div className="space-y-0.5">
                  <CardTitle className="text-sm font-medium flex items-center gap-1">
                      {displayTitle} {required && <span className="text-destructive">*</span>}
                  </CardTitle>
                  {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
              </div>
              <Button variant="ghost" size="sm" onClick={handleAddItem} className="h-6 gap-1">
                  <Plus className="h-3 w-3" /> Add
              </Button>
          </CardHeader>
          <CardContent className="p-3 space-y-3">
              {currentList.length === 0 && (
                  <div className="text-xs text-muted-foreground text-center py-2 italic">
                      No items.
                  </div>
              )}
              {currentList.map((item: any, index: number) => (
                  <div key={index} className="flex gap-2 items-start group">
                      <div className="flex-1 min-w-0">
                          <SchemaForm
                              schema={schema.items}
                              value={item}
                              onChange={(val) => handleItemChange(index, val)}
                              level={level + 1}
                              path={`${path}[${index}]`}
                              title={`Item ${index + 1}`}
                          />
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive shrink-0 mt-0.5"
                        onClick={() => handleRemoveItem(index)}
                        title="Remove Item"
                      >
                          <Trash2 className="h-4 w-4" />
                      </Button>
                  </div>
              ))}
          </CardContent>
      </Card>
    );
  }

  return <div className="text-red-500 text-xs">Unsupported schema type: {schema.type}</div>;
}


interface SchemaFieldProps {
    name: string;
    schema: any;
    value: any;
    onChange: (val: any) => void;
    required: boolean;
}

function SchemaField({ name, schema, value, onChange, required }: SchemaFieldProps) {
    const description = schema.description || "";
    const inputType = schema.format === "password" || name.toLowerCase().includes("token") || name.toLowerCase().includes("key") ? "password" : "text";

    if (schema.type === "boolean") {
        return (
            <div className="flex items-center space-x-2">
                <Checkbox
                    id={name}
                    checked={value === true || value === "true"}
                    onCheckedChange={(c) => onChange(c)}
                />
                <div className="grid gap-1.5 leading-none">
                    <label
                        htmlFor={name}
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                        {name}
                    </label>
                    {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
                </div>
            </div>
        );
    }

    if (schema.enum) {
        return (
            <div className="grid gap-1.5">
                <Label htmlFor={name} className="flex items-center gap-1">
                    {name} {required && <span className="text-destructive">*</span>}
                </Label>
                <Select value={value || ""} onValueChange={onChange}>
                    <SelectTrigger id={name}>
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

    // Default Input (Text/Number)
    return (
        <div className="grid gap-1.5">
            <Label htmlFor={name} className="flex items-center gap-1">
                {name} {required && <span className="text-destructive">*</span>}
            </Label>
            <Input
                id={name}
                value={value || ""}
                onChange={(e) => {
                    const val = e.target.value;
                    if (schema.type === "integer" || schema.type === "number") {
                         // Only update if valid number or empty
                         if (val === "" || !isNaN(Number(val))) {
                             onChange(val === "" ? undefined : Number(val));
                         }
                    } else {
                        onChange(val);
                    }
                }}
                placeholder={schema.default ? String(schema.default) : `Enter ${name}`}
                type={schema.type === "integer" || schema.type === "number" ? "number" : inputType}
            />
            {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
        </div>
    );
}
