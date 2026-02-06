/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { memo, useRef, useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { FileInput } from "@/components/ui/file-input";

interface SchemaFieldProps {
    path: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    schema: any; // Using any for schema as it can be recursive and variable
    value: unknown;
    onChange: (path: string, value: unknown) => void;
    errors?: Record<string, string>;
    required?: boolean;
    label?: string;
    level?: number;
    // Optimization props
    parentChange?: (key: string, value: unknown) => void;
    name?: string;
}

/**
 * SchemaForm component.
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @param props.errors - The error message or object.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, errors }: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    schema: any; // Using any for schema as it can be recursive and variable
    value: unknown;
    onChange: (value: unknown) => void;
    errors?: Record<string, string>;
}) {
    return (
        <div className="space-y-4">
            <SchemaField
                path=""
                schema={schema}
                value={value}
                onChange={(_, v) => onChange(v)}
                errors={errors}
                level={0}
            />
        </div>
    );
}

// ⚡ BOLT: Memoize SchemaField and use stable callbacks to prevent O(N) re-renders.
// Randomized Selection from Top 5 High-Impact Targets
/**
 * SchemaField component.
 * @param props - The component props.
 * @param props.path - The path property.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @param props.errors - The error message or object.
 * @param props.required - Whether the field is required.
 * @param props.label - The label property.
 * @param props.level - The level property.
 * @param props.parentChange - Optimized callback for parent updates.
 * @param props.name - The field name/key.
 * @returns The rendered component.
 */
const SchemaField = memo(function SchemaFieldInner({ path, schema, value, onChange, errors, required, label, level = 0, parentChange, name }: SchemaFieldProps) {
    // Optimization: Use ref to track value for stable callbacks
    const valueRef = useRef(value);
    valueRef.current = value;

    // Optimization: Create a stable change handler
    const handleChange = useCallback((p: string, v: unknown) => {
        if (parentChange && name !== undefined) {
            parentChange(name, v);
        } else {
            onChange(p, v);
        }
    }, [parentChange, name, onChange]);

    const type = schema.type;
    const description = schema.description;
    const isRequired = required;

    // Primitive Types
    if (schema.enum) {
        return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                <Select value={value as string} onValueChange={(v) => handleChange(path, v)}>
                    <SelectTrigger id={path} className={cn(errors?.[path] && "border-red-500")}>
                        <SelectValue placeholder="Select..." />
                    </SelectTrigger>
                    <SelectContent>
                        {schema.enum.map((val: string) => (
                            <SelectItem key={val} value={val}>{val}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </FieldWrapper>
        );
    }

    if (type === "string") {
        if (schema.contentEncoding === "base64" || schema.format === "binary") {
             return (
                <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                    <FileInput
                        id={path}
                        value={(value as string) || undefined}
                        onChange={(v) => handleChange(path, v)}
                        accept={schema.contentMediaType}
                        className={cn(errors?.[path] && "border-red-500")}
                    />
                </FieldWrapper>
            );
        }

         return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                <Input
                    id={path}
                    value={(value as string) || ""}
                    onChange={(e) => handleChange(path, e.target.value)}
                    placeholder={label || "Enter text"}
                    className={cn(errors?.[path] && "border-red-500")}
                />
            </FieldWrapper>
        );
    }

    if (type === "integer" || type === "number") {
        return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                <Input
                    id={path}
                    type="number"
                    step={type === "integer" ? "1" : "any"}
                    value={(value as number) || ""}
                    onChange={(e) => handleChange(path, e.target.value === "" ? undefined : Number(e.target.value))}
                    placeholder="0"
                    className={cn(errors?.[path] && "border-red-500")}
                />
            </FieldWrapper>
        );
    }

    if (type === "boolean") {
        return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                 <div className="flex items-center gap-2">
                    <Switch
                        id={path}
                        checked={!!value}
                        onCheckedChange={(v) => handleChange(path, v)}
                    />
                    <span className="text-sm text-muted-foreground">{value ? "True" : "False"}</span>
                </div>
            </FieldWrapper>
        );
    }

    // Complex Types: Object
    if (type === "object") {
        const properties = schema.properties || {};
        const requiredFields = schema.required || [];
        // Ensure value is object
        const objectValue = (value as Record<string, unknown>) || {};

        // Optimization: Use useCallback with ref to prevent re-creation
        const handleChildChange = useCallback((key: string, childValue: unknown) => {
            const currentValue = (valueRef.current as Record<string, unknown>) || {};
            const newValue = { ...currentValue, [key]: childValue };
            // We still need to call our 'handleChange' (which calls parent)
            // But here we are constructing the new object value for *this* field.
            if (parentChange && name !== undefined) {
                 parentChange(name, newValue);
            } else {
                 onChange(path, newValue);
            }
        }, [path, parentChange, name, onChange]);

        const content = (
            <div className={cn("space-y-3", level > 0 && "pl-4 border-l my-2")}>
                {Object.keys(properties).map((key) => (
                    <SchemaField
                        key={key}
                        path={path ? `${path}.${key}` : key}
                        name={key}
                        schema={properties[key]}
                        value={objectValue[key]}
                        onChange={onChange} // Pass original for fallback, but parentChange takes precedence
                        parentChange={handleChildChange}
                        errors={errors}
                        required={requiredFields.includes(key)}
                        label={key}
                        level={level + 1}
                    />
                ))}
            </div>
        );

        if (level === 0) return content; // Root object doesn't need wrapper

        return (
            <div className="space-y-1">
                 <Label className="flex items-center gap-1 font-semibold">
                    {label}
                    {isRequired && <span className="text-red-500">*</span>}
                    <span className="text-xs font-normal text-muted-foreground">({type})</span>
                </Label>
                 {description && <p className="text-[10px] text-muted-foreground mb-1">{description}</p>}
                {content}
            </div>
        );
    }

    // Complex Types: Array
    if (type === "array") {
        const itemsSchema = schema.items || {};
        const arrayValue: unknown[] = Array.isArray(value) ? value : [];

        // Optimization: Use useCallback with ref
        const handleAdd = useCallback(() => {
            const currentArray = Array.isArray(valueRef.current) ? valueRef.current : [];
            const newArray = [...currentArray, undefined];
            if (parentChange && name !== undefined) parentChange(name, newArray);
            else onChange(path, newArray);
        }, [path, parentChange, name, onChange]);

        const handleRemove = useCallback((index: number) => {
             const currentArray = Array.isArray(valueRef.current) ? valueRef.current : [];
             const newArray = [...currentArray];
             newArray.splice(index, 1);
             if (parentChange && name !== undefined) parentChange(name, newArray);
             else onChange(path, newArray);
        }, [path, parentChange, name, onChange]);

        const handleItemChange = useCallback((index: number, val: unknown) => {
             const currentArray = Array.isArray(valueRef.current) ? valueRef.current : [];
             const newArray = [...currentArray];
             newArray[index] = val;
             if (parentChange && name !== undefined) parentChange(name, newArray);
             else onChange(path, newArray);
        }, [path, parentChange, name, onChange]);

        return (
            <div className="space-y-2">
                <div className="flex items-center justify-between">
                     <Label className="flex items-center gap-1">
                        {label}
                        {isRequired && <span className="text-red-500">*</span>}
                         <span className="text-xs font-normal text-muted-foreground">([])</span>
                    </Label>
                    <Button type="button" variant="outline" size="sm" onClick={handleAdd} className="h-6 gap-1 text-xs">
                        <Plus className="size-3" /> Add Item
                    </Button>
                </div>
                 {description && <p className="text-[10px] text-muted-foreground">{description}</p>}

                <div className="space-y-2 pl-2">
                    {arrayValue.map((item, index) => (
                        <div key={index} className="flex gap-2 items-start group">
                            <div className="flex-1">
                                <SchemaField
                                    path={`${path}[${index}]`}
                                    name={index.toString()} // Array indices as names?
                                    // Special handling: parentChange expects (key, value). For array, we need index.
                                    // But handleItemChange expects (index, val).
                                    // So we can't directly use parentChange={handleItemChange}.
                                    // We need to construct a specific callback for each item?
                                    // Or genericize handleItemChange?
                                    // Let's rely on standard onChange for array items for now to keep it simple, or fix it.
                                    // If we use standard onChange, we lose optimization for array items.
                                    // Let's create a per-index callback? No, that's O(N).
                                    // Let's adapt handleItemChange.
                                    // If we pass parentChange={(k, v) => handleItemChange(Number(k), v)}
                                    // This creates a new function.
                                    // We can use a custom component for Array Item wrapper?
                                    // Let's just use `onChange` for Array items for now. Array re-ordering is tricky anyway.
                                    schema={itemsSchema}
                                    value={item}
                                    onChange={(_, v) => handleItemChange(index, v)}
                                    errors={errors}
                                    level={level + 1}
                                    label={`Item ${index + 1}`}
                                />
                            </div>
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRemove(index)}
                                className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive mt-6"
                            >
                                <Trash2 className="size-4" />
                            </Button>
                        </div>
                    ))}
                    {arrayValue.length === 0 && (
                        <div className="text-xs text-muted-foreground italic p-2 border border-dashed rounded bg-muted/20 text-center">
                            No items
                        </div>
                    )}
                </div>
            </div>
        );
    }

    // Fallback
    return (
        <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
             <Input
                id={path}
                value={(value as string) || ""}
                onChange={(e) => handleChange(path, e.target.value)}
                placeholder="Unknown Type"
                className={cn(errors?.[path] && "border-red-500")}
            />
        </FieldWrapper>
    );
});
SchemaField.displayName = "SchemaField";

/**
 * FieldWrapper component.
 * @param props - The component props.
 * @param props.children - The child components.
 * @param props.label - The label property.
 * @param props.description - The description property.
 * @param props.required - Whether the field is required.
 * @param props.error - The error message or object.
 * @param props.level - The level property.
 * @param props.inputId - The unique identifier for input.
 * @returns The rendered component.
 */
function FieldWrapper({ children, label, description, required, error, level, inputId }: {
    children: React.ReactNode, label?: string, description?: string, required?: boolean, error?: string, level: number, inputId?: string
}) {
    if (level === 0 && !label) return <>{children}</>;

    return (
        <div className="space-y-1">
            {label && (
                <Label htmlFor={inputId} className="flex items-center gap-1">
                    {label}
                    {required && <span className="text-red-500">*</span>}
                </Label>
            )}
            {description && (
                <p className="text-[10px] text-muted-foreground">{description}</p>
            )}
            {children}
            {error && <span className="text-xs text-red-500">{error}</span>}
        </div>
    );
}
