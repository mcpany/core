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
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  schema: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onChange: (value: any) => void;
  level?: number;
  title?: string;
  path?: string;
}

/**
 * SchemaForm component.
 * Recursively renders a form based on a JSON schema.
 * Supports: string, boolean, enum, object, array.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, level = 0, title, path = "root" }: SchemaFormProps) {
  if (!schema) return null;

  const isRoot = level === 0;

  // -- ARRAY TYPE --
  if (schema.type === "array") {
    const items = Array.isArray(value) ? value : [];
    const itemSchema = schema.items || { type: "string" };

    const handleAddItem = () => {
      // Default value for new item based on type
      let newItem: unknown = "";
      if (itemSchema.type === "object") newItem = {};
      if (itemSchema.type === "array") newItem = [];
      if (itemSchema.type === "boolean") newItem = false;
      if (itemSchema.default !== undefined) newItem = itemSchema.default;

      onChange([...items, newItem]);
    };

    const handleRemoveItem = (index: number) => {
      const newItems = [...items];
      newItems.splice(index, 1);
      onChange(newItems);
    };

    const handleItemChange = (index: number, val: unknown) => {
      const newItems = [...items];
      newItems[index] = val;
      onChange(newItems);
    };

    return (
      <div className={cn("space-y-2", isRoot ? "" : "border-l-2 border-muted pl-4 my-2")}>
        <div className="flex items-center justify-between">
           <Label className={cn(level === 0 ? "text-base font-semibold" : "text-sm font-medium")}>
               {title || schema.title || "List"}
           </Label>
           <Button variant="outline" size="sm" onClick={handleAddItem} type="button">
             <Plus className="h-3 w-3 mr-1" /> Add Item
           </Button>
        </div>
        {schema.description && <p className="text-xs text-muted-foreground mb-2">{schema.description}</p>}

        <div className="space-y-3">
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {items.map((item: any, index: number) => (
                <div key={index} className="flex gap-2 items-start group">
                    <div className="flex-1">
                        <SchemaForm
                            schema={itemSchema}
                            value={item}
                            onChange={(v) => handleItemChange(index, v)}
                            level={level + 1}
                            title={`Item ${index + 1}`}
                            path={`${path}[${index}]`}
                        />
                    </div>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive mt-1"
                        onClick={() => handleRemoveItem(index)}
                        type="button"
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            {items.length === 0 && (
                <div className="text-sm text-muted-foreground italic p-2 border border-dashed rounded text-center">
                    No items.
                </div>
            )}
        </div>
      </div>
    );
  }

  // -- OBJECT TYPE --
  if (schema.type === "object" || schema.properties) {
    const currentVal = value || {};

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleChange = (key: string, val: any) => {
      onChange({ ...currentVal, [key]: val });
    };

    return (
        <div className={cn("space-y-4", !isRoot && "my-2")}>
             {(!isRoot && (title || schema.title)) && (
                 <h4 className="font-medium text-sm flex items-center gap-2 mb-2 text-primary">
                    {title || schema.title}
                 </h4>
             )}
             {(!isRoot && schema.description) && (
                 <p className="text-xs text-muted-foreground -mt-1 mb-2">{schema.description}</p>
             )}

            <div className={cn("grid gap-4", !isRoot && "pl-4 border-l-2 border-muted")}>
                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                {Object.entries(schema.properties || {}).map(([key, prop]: [string, any]) => {
                    const isRequired = schema.required?.includes(key);
                    const propTitle = prop.title || key;

                    return (
                        <div key={key}>
                            {/* Recursive Call */}
                            <SchemaForm
                                schema={prop}
                                value={currentVal[key]}
                                onChange={(v) => handleChange(key, v)}
                                level={level + 1}
                                title={`${propTitle}${isRequired ? ' *' : ''}`}
                                path={`${path}.${key}`}
                            />
                        </div>
                    );
                })}
            </div>
        </div>
    );
  }

  // -- PRIMITIVE TYPES --
  const fieldId = `field-${path}`; // Use unique path for ID

  return (
    <div className="grid gap-2">
       {title && (
           <Label htmlFor={fieldId} className="flex items-center gap-1">
                {title}
           </Label>
       )}

        {schema.type === "boolean" ? (
           <div className="flex items-center space-x-2 h-9">
            <Checkbox
                id={fieldId}
                checked={value === true || value === "true"}
                onCheckedChange={(c) => onChange(c)}
            />
            <label
                htmlFor={fieldId}
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
                {schema.description || "Enable"}
            </label>
           </div>
        ) : schema.enum ? (
            <Select value={value} onValueChange={(v) => onChange(v)}>
                <SelectTrigger id={fieldId}>
                    <SelectValue placeholder="Select..." />
                </SelectTrigger>
                <SelectContent>
                    {schema.enum.map((opt: string) => (
                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                    ))}
                </SelectContent>
            </Select>
        ) : (
            <Input
                id={fieldId}
                value={value || ""}
                onChange={(e) => onChange(e.target.value)}
                placeholder={schema.default || schema.placeholder || `Enter ${title}`}
                type={schema.format === "password" || (title && title.toLowerCase().includes("token")) ? "password" : "text"}
            />
        )}

        {schema.type !== "boolean" && schema.description && (
            <p className="text-xs text-muted-foreground">{schema.description}</p>
        )}
    </div>
  );
}
