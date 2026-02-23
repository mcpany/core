/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2, ChevronRight, ChevronDown, Info, HelpCircle, GripVertical } from "lucide-react";
import { cn } from "@/lib/utils";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";

/**
 * Schema definition interface.
 * Represents a JSON Schema object used for form generation.
 */
export interface Schema {
    type?: string | string[];
    description?: string;
    properties?: Record<string, Schema>;
    items?: Schema;
    required?: string[];
    anyOf?: Schema[];
    oneOf?: Schema[];
    enum?: any[];
    default?: any;
    format?: string;
    title?: string;
    [key: string]: any;
}

interface UniversalSchemaFormProps {
    schema: Schema;
    value: any;
    onChange: (value: any) => void;
    errors?: Record<string, string>;
}

interface SchemaFieldProps {
    path: string;
    schema: Schema;
    value: any;
    onChange: (path: string, value: any) => void;
    errors?: Record<string, string>;
    required?: boolean;
    label?: string;
    level?: number;
    isLast?: boolean;
}

/**
 * UniversalSchemaForm component.
 * Renders a form based on a JSON schema with improved UX for nested objects and arrays.
 */
export function UniversalSchemaForm({ schema, value, onChange, errors }: UniversalSchemaFormProps) {
    return (
        <TooltipProvider delayDuration={300}>
            <div className="space-y-4 pb-4">
                <SchemaField
                    path=""
                    schema={schema}
                    value={value}
                    onChange={(_, v) => onChange(v)}
                    errors={errors}
                    level={0}
                />
            </div>
        </TooltipProvider>
    );
}

/**
 * Helper to display field label with optional tooltip and required indicator.
 */
function FieldLabel({ label, required, description, error, className, htmlFor }: { label?: string, required?: boolean, description?: string, error?: string, className?: string, htmlFor?: string }) {
    if (!label) return null;
    return (
        <div className={cn("flex flex-col gap-1 mb-1.5", className)}>
            <div className="flex items-center gap-2">
                <Label htmlFor={htmlFor} className={cn("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", error && "text-destructive")}>
                    {label}
                    {required && <span className="text-destructive ml-0.5">*</span>}
                </Label>
                {description && (
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <HelpCircle className="h-3.5 w-3.5 text-muted-foreground cursor-help opacity-70 hover:opacity-100 transition-opacity" />
                        </TooltipTrigger>
                        <TooltipContent className="max-w-xs text-xs">
                            <p>{description}</p>
                        </TooltipContent>
                    </Tooltip>
                )}
            </div>
            {error && (
                <p className="text-[11px] font-medium text-destructive animate-in slide-in-from-top-1 fade-in duration-200">
                    {error}
                </p>
            )}
        </div>
    );
}

/**
 * Recursive Schema Field Component
 */
function SchemaField({ path, schema, value, onChange, errors, required, label, level = 0, isLast }: SchemaFieldProps) {
    let type = schema.type;
    // Normalize type array (e.g., ["string", "null"] -> "string")
    if (Array.isArray(type)) {
        type = type.find((t: string) => t !== "null");
    }

    const description = schema.description;
    const isRequired = required;
    const hasError = !!errors?.[path];

    // --- Handle oneOf / anyOf ---
    if (schema.oneOf || schema.anyOf) {
        const options = schema.oneOf || schema.anyOf || [];
        // Try to detect which option matches the current value
        const [selectedIndex, setSelectedIndex] = useState(0);

        // Heuristic detection on mount
        useEffect(() => {
            if (value && typeof value === 'object') {
                let bestIdx = 0;
                let maxMatch = -1;

                options.forEach((opt: any, idx: number) => {
                    if (!opt.properties) return;
                    const keys = Object.keys(value as object);
                    let matches = 0;
                    keys.forEach(k => {
                        if (k in opt.properties) matches++;
                    });

                    if (matches > maxMatch) {
                        maxMatch = matches;
                        bestIdx = idx;
                    }
                });
                setSelectedIndex(bestIdx);
            }
        }, []); // eslint-disable-line react-hooks/exhaustive-deps

        const selectedSchema = options[selectedIndex];

        return (
            <div className={cn("border rounded-md p-3 bg-muted/10", hasError && "border-destructive/50")}>
                <div className="flex items-center justify-between mb-3">
                    <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                        {label || schema.title || "Option"}
                    </Label>
                    <Select
                        value={String(selectedIndex)}
                        onValueChange={(idx) => {
                            setSelectedIndex(Number(idx));
                            // Optional: clear value on type switch?
                            // onChange(path, {});
                        }}
                    >
                        <SelectTrigger className="h-7 w-[200px] text-xs bg-background">
                            <SelectValue placeholder="Select Type" />
                        </SelectTrigger>
                        <SelectContent>
                            {options.map((opt: any, idx: number) => (
                                <SelectItem key={idx} value={String(idx)} className="text-xs">
                                    {opt.title || opt.description || `Option ${idx + 1}`}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
                <SchemaField
                    path={path}
                    schema={selectedSchema}
                    value={value}
                    onChange={onChange}
                    errors={errors}
                    level={level}
                />
            </div>
        );
    }

    // --- Enum (Select) ---
    if (schema.enum) {
        return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} htmlFor={path} />
                <Select value={value as string} onValueChange={(v) => onChange(path, v)}>
                    <SelectTrigger id={path} className={cn(hasError && "border-destructive ring-destructive/20")}>
                        <SelectValue placeholder="Select option..." />
                    </SelectTrigger>
                    <SelectContent>
                        {schema.enum.map((val: string) => (
                            <SelectItem key={val} value={val}>{val}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
        );
    }

    // --- Simple Types ---

    if (type === "string") {
        const isPassword = schema.format === "password" ||
                           (label && /password|secret|token|key/i.test(label));
        const isMultiline = schema.format === "multiline" || (label && /description|content/i.test(label));

        return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} htmlFor={path} />
                {isMultiline ? (
                     <Textarea
                        id={path}
                        value={(value as string) || ""}
                        onChange={(e) => onChange(path, e.target.value)}
                        placeholder={schema.default || schema.examples?.[0] || "Enter text"}
                        className={cn("min-h-[80px]", hasError && "border-destructive focus-visible:ring-destructive/20")}
                    />
                ) : (
                    <Input
                        id={path}
                        type={isPassword ? "password" : "text"}
                        value={(value as string) || ""}
                        onChange={(e) => onChange(path, e.target.value)}
                        placeholder={schema.default || schema.examples?.[0] || "Enter text"}
                        className={cn(hasError && "border-destructive focus-visible:ring-destructive/20")}
                    />
                )}
            </div>
        );
    }

    if (type === "integer" || type === "number") {
        return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} htmlFor={path} />
                <Input
                    id={path}
                    type="number"
                    step={type === "integer" ? "1" : "any"}
                    value={(value as number) || ""}
                    onChange={(e) => onChange(path, e.target.value === "" ? undefined : Number(e.target.value))}
                    placeholder={schema.default ? String(schema.default) : "0"}
                    className={cn(hasError && "border-destructive focus-visible:ring-destructive/20")}
                />
            </div>
        );
    }

    if (type === "boolean") {
        return (
             <div className="flex items-center justify-between py-2 rounded-md">
                <div className="space-y-0.5">
                    <Label htmlFor={path} className="text-sm font-medium">{label}</Label>
                    {description && <p className="text-[11px] text-muted-foreground">{description}</p>}
                </div>
                <Switch
                    id={path}
                    checked={!!value}
                    onCheckedChange={(v) => onChange(path, v)}
                />
            </div>
        );
    }

    // --- Complex Types: Object ---
    if (type === "object") {
        const properties = schema.properties || {};
        const requiredFields = schema.required || [];
        const objectValue = (value as Record<string, unknown>) || {};

        const handleChildChange = (key: string, childValue: unknown) => {
            const newValue = { ...objectValue, [key]: childValue };
            onChange(path, newValue);
        };

        const hasProperties = Object.keys(properties).length > 0;

        // Root level object (level 0) - Render flat
        if (level === 0) {
            return (
                <div className="space-y-5">
                    {hasProperties ? (
                        Object.entries(properties).map(([key, propSchema]) => (
                            <SchemaField
                                key={key}
                                path={path ? `${path}.${key}` : key}
                                schema={propSchema}
                                value={objectValue[key]}
                                onChange={(_, v) => handleChildChange(key, v)}
                                errors={errors}
                                required={requiredFields.includes(key)}
                                label={propSchema.title || key}
                                level={level + 1}
                            />
                        ))
                    ) : (
                        <div className="text-sm text-muted-foreground italic p-4 border border-dashed rounded bg-muted/20 text-center">
                            No properties defined.
                        </div>
                    )}
                </div>
            );
        }

        // Nested Object - Render as Collapsible Section
        return (
            <Collapsible defaultOpen={true} className={cn("w-full group", hasError ? "border-l-2 border-destructive pl-3" : "pl-2 border-l border-border/50")}>
                <div className="flex items-center justify-between py-2">
                    <CollapsibleTrigger className="flex items-center gap-2 hover:bg-muted/50 p-1 rounded -ml-1">
                        <ChevronDown className="h-3 w-3 text-muted-foreground transition-transform group-data-[state=closed]:-rotate-90" />
                        <span className="text-sm font-semibold">{label || "Object"}</span>
                        {isRequired && <span className="text-destructive text-xs">*</span>}
                    </CollapsibleTrigger>
                    {description && (
                        <Tooltip>
                            <TooltipTrigger>
                                <Info className="h-3 w-3 text-muted-foreground/50" />
                            </TooltipTrigger>
                            <TooltipContent><p>{description}</p></TooltipContent>
                        </Tooltip>
                    )}
                </div>
                <CollapsibleContent className="space-y-4 pt-1 pb-2 pl-2">
                     {Object.entries(properties).map(([key, propSchema]) => (
                        <SchemaField
                            key={key}
                            path={path ? `${path}.${key}` : key}
                            schema={propSchema}
                            value={objectValue[key]}
                            onChange={(_, v) => handleChildChange(key, v)}
                            errors={errors}
                            required={requiredFields.includes(key)}
                            label={propSchema.title || key}
                            level={level + 1}
                        />
                    ))}
                </CollapsibleContent>
            </Collapsible>
        );
    }

    // --- Complex Types: Array ---
    if (type === "array") {
        const itemsSchema = schema.items || {};
        const arrayValue: unknown[] = Array.isArray(value) ? value : [];

        const handleAdd = () => {
            onChange(path, [...arrayValue, undefined]);
        };

        const handleRemove = (index: number) => {
            const newArray = [...arrayValue];
            newArray.splice(index, 1);
            onChange(path, newArray);
        };

        const handleItemChange = (index: number, val: unknown) => {
            const newArray = [...arrayValue];
            newArray[index] = val;
            onChange(path, newArray);
        };

        return (
            <div className={cn("w-full space-y-2", hasError ? "border-l-2 border-destructive pl-3" : "pl-2 border-l border-border/50")}>
                 <div className="flex items-center justify-between py-1">
                    <div className="flex items-center gap-2">
                        <Label className="text-sm font-semibold">{label || "List"}</Label>
                        <Badge variant="secondary" className="text-[10px] h-4 px-1 rounded-sm">Array</Badge>
                        {isRequired && <span className="text-destructive">*</span>}
                    </div>
                    <Button type="button" variant="ghost" size="sm" onClick={handleAdd} className="h-6 text-xs gap-1 hover:bg-muted">
                        <Plus className="size-3" /> Add Item
                    </Button>
                </div>
                {description && <p className="text-xs text-muted-foreground -mt-1 mb-2">{description}</p>}

                <div className="space-y-2">
                    {arrayValue.length === 0 ? (
                        <div className="text-xs text-muted-foreground italic p-2 bg-muted/20 rounded text-center">
                            Empty list.
                        </div>
                    ) : (
                        arrayValue.map((item, index) => (
                            <div key={index} className="flex gap-2 items-start group relative pl-2">
                                <div className="mt-2.5 text-muted-foreground/30 absolute left-0">
                                     <div className="w-1 h-1 rounded-full bg-current" />
                                </div>
                                <div className="flex-1 min-w-0">
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
                                    size="icon"
                                    onClick={() => handleRemove(index)}
                                    className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                                    title="Remove Item"
                                >
                                    <Trash2 className="size-3" />
                                </Button>
                            </div>
                        ))
                    )}
                </div>
            </div>
        );
    }

    // Fallback
    return (
        <div className="w-full">
            <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} htmlFor={path} />
            <Input
                id={path}
                value={(value as string) || ""}
                onChange={(e) => onChange(path, e.target.value)}
                placeholder="Value"
                className={cn(hasError && "border-destructive")}
            />
        </div>
    );
}
