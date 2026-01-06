/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface SchemaFieldProps {
    path: string;
    schema: any;
    value: any;
    onChange: (path: string, value: any) => void;
    errors?: Record<string, string>;
    required?: boolean;
    label?: string;
    level?: number;
}

export function SchemaForm({ schema, value, onChange, errors }: {
    schema: any;
    value: any;
    onChange: (value: any) => void;
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

function SchemaField({ path, schema, value, onChange, errors, required, label, level = 0 }: SchemaFieldProps) {
    const type = schema.type;
    const description = schema.description;
    const isRequired = required;

    // Primitive Types
    if (schema.enum) {
        return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                <Select value={value} onValueChange={(v) => onChange(path, v)}>
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
         return (
            <FieldWrapper label={label} description={description} required={isRequired} error={errors?.[path]} level={level} inputId={path}>
                <Input
                    id={path}
                    value={value || ""}
                    onChange={(e) => onChange(path, e.target.value)}
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
                    value={value || ""}
                    onChange={(e) => onChange(path, e.target.value === "" ? undefined : Number(e.target.value))}
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
                        onCheckedChange={(v) => onChange(path, v)}
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
        const objectValue = value || {};

        const handleChildChange = (key: string, childValue: any) => {
            const newValue = { ...objectValue, [key]: childValue };
            onChange(path, newValue);
        };

        const content = (
            <div className={cn("space-y-3", level > 0 && "pl-4 border-l my-2")}>
                {Object.keys(properties).map((key) => (
                    <SchemaField
                        key={key}
                        path={path ? `${path}.${key}` : key}
                        schema={properties[key]}
                        value={objectValue[key]}
                        onChange={(_, v) => handleChildChange(key, v)}
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
        const arrayValue: any[] = Array.isArray(value) ? value : [];

        const handleAdd = () => {
            onChange(path, [...arrayValue, undefined]);
        };

        const handleRemove = (index: number) => {
            const newArray = [...arrayValue];
            newArray.splice(index, 1);
            onChange(path, newArray);
        };

        const handleItemChange = (index: number, val: any) => {
            const newArray = [...arrayValue];
            newArray[index] = val;
            onChange(path, newArray);
        };

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
                value={value || ""}
                onChange={(e) => onChange(path, e.target.value)}
                placeholder="Unknown Type"
                className={cn(errors?.[path] && "border-red-500")}
            />
        </FieldWrapper>
    );
}

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
