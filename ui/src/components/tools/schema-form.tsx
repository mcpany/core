/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useEffect } from "react";
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
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

/**
 * Schema interface representing a JSON Schema structure.
 */
export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  anyOf?: Schema[];
  oneOf?: Schema[];
  allOf?: Schema[];
  enum?: any[];
  default?: any;
  format?: string;
  title?: string;
  [key: string]: any;
}

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  level?: number;
}

/**
 * A dynamic form generator based on JSON Schema.
 */
export function SchemaForm({ schema, value, onChange, name, required, level = 0 }: SchemaFormProps) {
    // Helper to update a specific field in an object
    const handleFieldChange = useCallback((fieldName: string, fieldValue: any) => {
        const newValue = { ...value, [fieldName]: fieldValue };
        // If fieldValue is undefined, we might want to delete the key?
        // For now, keep it simple.
        if (fieldValue === undefined) {
             delete newValue[fieldName];
        }
        onChange(newValue);
    }, [value, onChange]);

    // Initialize default value if missing
    useEffect(() => {
        if (value === undefined && schema.default !== undefined) {
            onChange(schema.default);
        } else if (value === undefined) {
             if (schema.type === "object") {
                 // Don't auto-initialize objects unless they are required?
                 // Let's rely on parent to init if needed, or init as empty object if it's the root.
             } else if (schema.type === "array") {
                 // onChange([]);
             } else if (schema.type === "boolean") {
                 // onChange(false);
             }
        }
    }, [schema.default, value, schema.type]); // Intentionally not adding onChange to deps to avoid loops if careful

    if (!schema) return null;

    const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
    const label = schema.title || name || "Value";
    const description = schema.description;

    // --- ENUM (Select) ---
    if (schema.enum) {
        return (
            <div className={cn("grid gap-2", level > 0 && "mb-4")}>
                <div className="flex items-center gap-2">
                    <Label className={cn(required && "text-destructive font-bold")}>
                        {label} {required && "*"}
                    </Label>
                    {description && (
                        <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                                <Info className="h-4 w-4 text-muted-foreground/70 cursor-help" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-[300px] text-xs">
                                <p>{description}</p>
                            </TooltipContent>
                        </Tooltip>
                    )}
                </div>
                <Select
                    value={value !== undefined ? String(value) : undefined}
                    onValueChange={(v) => {
                        // Try to preserve type if original enum values are numbers/booleans
                        const original = schema.enum?.find(e => String(e) === v);
                        onChange(original !== undefined ? original : v);
                    }}
                >
                    <SelectTrigger>
                        <SelectValue placeholder="Select..." />
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

    // --- BOOLEAN ---
    if (type === "boolean") {
        return (
            <div className={cn("flex items-center justify-between rounded-lg border p-3 shadow-sm", level > 0 && "mb-4")}>
                <div className="space-y-0.5">
                    <Label className={cn("text-base", required && "text-destructive font-bold")}>
                        {label} {required && "*"}
                    </Label>
                     {description && (
                        <p className="text-xs text-muted-foreground">{description}</p>
                    )}
                </div>
                <Switch
                    checked={!!value}
                    onCheckedChange={onChange}
                />
            </div>
        );
    }

    // --- STRING / NUMBER ---
    if (type === "string" || type === "number" || type === "integer") {
        return (
             <div className={cn("grid gap-2", level > 0 && "mb-4")}>
                <div className="flex items-center gap-2">
                    <Label className={cn(required && "text-destructive font-bold")}>
                        {label} {required && "*"}
                    </Label>
                    {description && (
                        <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                                <Info className="h-4 w-4 text-muted-foreground/70 cursor-help" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-[300px] text-xs">
                                <p>{description}</p>
                            </TooltipContent>
                        </Tooltip>
                    )}
                </div>
                <Input
                    type={type === "string" ? "text" : "number"}
                    step={type === "integer" ? "1" : "any"}
                    value={value !== undefined ? value : ""}
                    onChange={(e) => {
                        const v = e.target.value;
                        if (type === "number" || type === "integer") {
                            onChange(v === "" ? undefined : Number(v));
                        } else {
                            onChange(v);
                        }
                    }}
                    placeholder={schema.default ? String(schema.default) : `Enter ${label}...`}
                />
            </div>
        );
    }

    // --- OBJECT ---
    if (type === "object" && schema.properties) {
        // If it's root level (level 0), don't wrap in Card, just list fields
        // If it's nested, wrap in Card
        const Wrapper = level === 0 ? React.Fragment : Card;
        const wrapperProps = level === 0 ? {} : { className: "mb-4 bg-muted/20" };

        const content = (
            <div className={cn("space-y-4", level > 0 && "p-4")}>
                {Object.entries(schema.properties).map(([propName, propSchema]) => (
                    <SchemaForm
                        key={propName}
                        name={propName}
                        schema={propSchema}
                        value={value ? value[propName] : undefined}
                        onChange={(v) => handleFieldChange(propName, v)}
                        required={schema.required?.includes(propName)}
                        level={level + 1}
                    />
                ))}
            </div>
        );

        if (level === 0) return content;

        return (
            <Card className="mb-4 border-dashed shadow-none">
                 <CardHeader className="py-3 px-4 bg-muted/30">
                     <CardTitle className="text-sm font-medium flex items-center gap-2">
                         {label}
                         {description && (
                            <Tooltip delayDuration={300}>
                                <TooltipTrigger asChild>
                                    <Info className="h-3 w-3 text-muted-foreground/70 cursor-help" />
                                </TooltipTrigger>
                                <TooltipContent className="max-w-[300px] text-xs">
                                    <p>{description}</p>
                                </TooltipContent>
                            </Tooltip>
                        )}
                     </CardTitle>
                 </CardHeader>
                 {content}
            </Card>
        );
    }

    // --- ARRAY ---
    if (type === "array" && schema.items) {
        const items = Array.isArray(value) ? value : [];
        const addItem = () => {
            // Add default value if exists, else undefined (which child form handles)
            // But we need a concrete value.
            // Let's pass null or empty string depending on item schema type
            let newVal = undefined;
            const itemSchema = schema.items as Schema;
            if (itemSchema.type === 'string') newVal = "";
            else if (itemSchema.type === 'number') newVal = 0;
            else if (itemSchema.type === 'boolean') newVal = false;
            else if (itemSchema.type === 'object') newVal = {};
            else if (itemSchema.type === 'array') newVal = [];

            onChange([...items, newVal]);
        };

        const removeItem = (index: number) => {
            const newItems = [...items];
            newItems.splice(index, 1);
            onChange(newItems);
        };

        const updateItem = (index: number, v: any) => {
             const newItems = [...items];
             newItems[index] = v;
             onChange(newItems);
        };

        return (
            <div className={cn("space-y-2", level > 0 && "mb-4")}>
                 <div className="flex items-center justify-between">
                     <Label className={cn(required && "text-destructive font-bold")}>
                        {label} (List)
                     </Label>
                      <Button type="button" variant="outline" size="sm" onClick={addItem} className="h-6">
                        <Plus className="h-3 w-3 mr-1" /> Add
                     </Button>
                 </div>
                 {description && <p className="text-xs text-muted-foreground">{description}</p>}

                 <div className="space-y-2 pl-2 border-l-2 border-muted">
                     {items.map((itemValue: any, idx: number) => (
                         <div key={idx} className="flex gap-2 items-start">
                             <div className="flex-1">
                                 <SchemaForm
                                     schema={schema.items as Schema}
                                     value={itemValue}
                                     onChange={(v) => updateItem(idx, v)}
                                     name={`Item ${idx + 1}`}
                                     level={level + 1}
                                 />
                             </div>
                             <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                onClick={() => removeItem(idx)}
                                className="text-destructive h-8 w-8 mt-6" // Align with input
                            >
                                <Trash2 className="h-4 w-4" />
                            </Button>
                         </div>
                     ))}
                     {items.length === 0 && (
                         <div className="text-xs text-muted-foreground italic py-2">No items added.</div>
                     )}
                 </div>
            </div>
        );
    }

    // Fallback for unknown types
    return (
         <div className="text-xs text-destructive p-2 border border-destructive rounded bg-destructive/10">
             Unsupported field type: {String(type)}
         </div>
    );
}
