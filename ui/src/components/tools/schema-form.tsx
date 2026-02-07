/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Trash2, Plus } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Schema represents a JSON Schema object for form generation.
 */
export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: any[];
  default?: any;
  title?: string;
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

const getDefaultValue = (schema: Schema) => {
    if (schema.default !== undefined) return schema.default;
    const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
    if (type === "boolean") return false;
    if (type === "string") return "";
    if (type === "number" || type === "integer") return 0;
    if (type === "array") return [];
    if (type === "object") return {};
    return undefined;
};

const NumberInput = ({ value, onChange, ...props }: { value: number | undefined, onChange: (val: number | undefined) => void } & Omit<React.ComponentProps<typeof Input>, "value" | "onChange">) => {
    const [localValue, setLocalValue] = useState<string>(value !== undefined ? value.toString() : "");

    useEffect(() => {
        const parsedLocal = localValue === "" ? undefined : parseFloat(localValue);
        if (parsedLocal !== value) {
             setLocalValue(value !== undefined ? value.toString() : "");
        }
    }, [value, localValue]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const val = e.target.value;
        setLocalValue(val);

        if (val === "") {
            onChange(undefined);
            return;
        }

        const parsed = parseFloat(val);
        if (!isNaN(parsed)) {
             onChange(parsed);
        }
    };

    return <Input type="number" value={localValue} onChange={handleChange} {...props} />;
};

/**
 * SchemaForm generates a form from a JSON Schema.
 * @param props - The component props.
 * @param props.schema - The JSON schema to render.
 * @param props.value - The current value of the form.
 * @param props.onChange - Callback when the value changes.
 * @param props.name - Optional name for the field.
 * @param props.required - Whether the field is required.
 * @param props.depth - Recursion depth.
 * @returns The rendered form component.
 */
export function SchemaForm({ schema, value, onChange, name, required, depth = 0 }: SchemaFormProps) {
    // Determine effective type
    const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
    const id = React.useId();

    const handleChange = (newValue: any) => {
        onChange(newValue);
    };

    const label = (
        <div className="flex items-center gap-2 mb-1.5">
            <Label htmlFor={id} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
                {schema.title || name || (depth === 0 ? "Parameters" : "Value")}
            </Label>
            {schema.description && (
                <span className="text-xs text-muted-foreground truncate max-w-[300px]" title={schema.description}>
                    - {schema.description}
                </span>
            )}
        </div>
    );

    // Enum (Select)
    if (schema.enum) {
        return (
            <div className="space-y-1">
                {label}
                <Select value={value !== undefined ? String(value) : ""} onValueChange={(v) => handleChange(v)}>
                    <SelectTrigger id={id}>
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

    // Boolean (Switch)
    if (type === "boolean") {
        return (
            <div className="flex items-center justify-between space-x-2 py-2 border rounded-md px-3 bg-muted/20">
                <div className="space-y-0.5">
                   <Label htmlFor={id} className={cn("text-base", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
                        {schema.title || name || "Value"}
                    </Label>
                    {schema.description && <p className="text-xs text-muted-foreground">{schema.description}</p>}
                </div>
                <Switch id={id} checked={!!value} onCheckedChange={handleChange} />
            </div>
        );
    }

    // String (Input)
    if (type === "string") {
        return (
            <div className="space-y-1">
                {label}
                <Input
                    id={id}
                    value={value || ""}
                    onChange={(e) => handleChange(e.target.value)}
                    placeholder={schema.default ? String(schema.default) : ""}
                />
            </div>
        );
    }

    // Number/Integer (Input type="number")
    if (type === "number" || type === "integer") {
         return (
            <div className="space-y-1">
                {label}
                <NumberInput
                    id={id}
                    value={value}
                    onChange={(val) => {
                        if (type === "integer" && val !== undefined) {
                            handleChange(Math.floor(val));
                        } else {
                            handleChange(val);
                        }
                    }}
                    step={type === "integer" ? "1" : "any"}
                />
            </div>
        );
    }

    // Object (Nested Form)
    if (type === "object" && schema.properties) {
        const objectValue = value || {};

        // If it's the root object (depth 0), we might want to skip the card wrapper for a cleaner look
        // or ensure it fits the dialog. But usually a Card is good for grouping.
        // Let's use a cleaner layout for root.

        if (depth === 0) {
             return (
                <div className="space-y-4">
                    {Object.entries(schema.properties).map(([key, propSchema]) => (
                        <SchemaForm
                            key={key}
                            name={key}
                            schema={propSchema}
                            value={objectValue[key]}
                            onChange={(newValue) => handleChange({ ...objectValue, [key]: newValue })}
                            required={schema.required?.includes(key)}
                            depth={depth + 1}
                        />
                    ))}
                </div>
            );
        }

        return (
            <Card className={cn("border-l-4 mt-2", depth % 2 === 0 ? "border-l-primary/20" : "border-l-secondary/20")}>
                <CardHeader className="py-2 px-4 bg-muted/10">
                    <CardTitle className="text-sm font-medium flex items-center gap-2">
                        {schema.title || name || "Object"}
                        {required && <span className="text-red-500">*</span>}
                    </CardTitle>
                    {schema.description && <p className="text-xs text-muted-foreground">{schema.description}</p>}
                </CardHeader>
                <CardContent className="p-4 space-y-4">
                    {Object.entries(schema.properties).map(([key, propSchema]) => (
                        <SchemaForm
                            key={key}
                            name={key}
                            schema={propSchema}
                            value={objectValue[key]}
                            onChange={(newValue) => handleChange({ ...objectValue, [key]: newValue })}
                            required={schema.required?.includes(key)}
                            depth={depth + 1}
                        />
                    ))}
                </CardContent>
            </Card>
        );
    }

    // Array (Dynamic List)
    if (type === "array" && schema.items) {
        const arrayValue = Array.isArray(value) ? value : [];
        const addItem = () => {
            const defaultValue = getDefaultValue(schema.items!);
            handleChange([...arrayValue, defaultValue]);
        };
        const removeItem = (index: number) => {
            const newArray = [...arrayValue];
            newArray.splice(index, 1);
            handleChange(newArray);
        };
        const updateItem = (index: number, val: any) => {
             const newArray = [...arrayValue];
             newArray[index] = val;
             handleChange(newArray);
        };

        return (
            <div className="space-y-2 mt-2">
                <div className="flex items-center justify-between">
                     {label}
                     <Button type="button" variant="outline" size="sm" onClick={addItem} className="h-7 text-xs">
                        <Plus className="mr-1 h-3 w-3" /> Add Item
                     </Button>
                </div>
                <div className="space-y-2 pl-2 border-l-2 border-dashed border-muted">
                    {arrayValue.map((item: any, index: number) => (
                        <div key={index} className="flex items-start gap-2 group relative pr-8">
                            <div className="flex-1">
                                <SchemaForm
                                    schema={schema.items!}
                                    value={item}
                                    onChange={(v) => updateItem(index, v)}
                                    depth={depth + 1}
                                    name={`Item ${index + 1}`}
                                />
                            </div>
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-destructive opacity-50 group-hover:opacity-100 transition-opacity absolute right-0 top-0"
                                onClick={() => removeItem(index)}
                            >
                                <Trash2 className="h-4 w-4" />
                            </Button>
                        </div>
                    ))}
                    {arrayValue.length === 0 && (
                        <div className="text-xs text-muted-foreground italic py-2">
                            No items.
                        </div>
                    )}
                </div>
            </div>
        );
    }

    // Fallback
    return (
        <div className="p-2 border border-red-200 bg-red-50 text-red-800 text-xs rounded">
            Unsupported type: {String(type)}
        </div>
    );
}
