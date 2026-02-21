/**
 * Copyright 2025 Author(s) of MCP Any
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
import { FileInput } from "@/components/ui/file-input";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
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
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";

interface SchemaFieldProps {
    path: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    schema: any;
    value: unknown;
    onChange: (path: string, value: unknown) => void;
    errors?: Record<string, string>;
    required?: boolean;
    label?: string;
    level?: number;
}

/**
 * SchemaForm component.
 * Renders a form based on a JSON schema.
 */
export function SchemaForm({ schema, value, onChange, errors }: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    schema: any;
    value: unknown;
    onChange: (value: unknown) => void;
    errors?: Record<string, string>;
}) {
    return (
        <TooltipProvider delayDuration={300}>
            <div className="space-y-6 pb-8">
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
function FieldLabel({ label, required, description, error, className }: { label?: string, required?: boolean, description?: string, error?: string, className?: string }) {
    if (!label) return null;
    return (
        <div className={cn("flex flex-col gap-1.5 mb-2", className)}>
            <div className="flex items-center gap-2">
                <Label className={cn("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", error && "text-destructive")}>
                    {label}
                    {required && <span className="text-destructive ml-1">*</span>}
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
            {description && !error && (
                 <p className="text-[11px] text-muted-foreground">{description}</p>
            )}
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
function SchemaField({ path, schema, value, onChange, errors, required, label, level = 0 }: SchemaFieldProps) {
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
        const options = schema.oneOf || schema.anyOf;
        // Try to detect which option matches the current value
        // Simple heuristic: if option has properties that match keys in value
        const [selectedIndex, setSelectedIndex] = useState(0);

        // On mount, try to guess the index if we have a value
        useEffect(() => {
            if (value && typeof value === 'object') {
                // Find best match based on property overlap
                let bestIdx = 0;
                let maxMatch = -1;

                options.forEach((opt: any, idx: number) => {
                    if (!opt.properties) return;
                    // Count how many keys in value are present in opt.properties
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
        }, []);

        const selectedSchema = options[selectedIndex];

        return (
            <Card className={cn("border-dashed", hasError && "border-destructive/50")}>
                <CardHeader className="p-3 pb-2 bg-muted/20">
                    <div className="flex items-center justify-between">
                        <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            {label || "Option"} (One Of)
                        </Label>
                        <Select
                            value={String(selectedIndex)}
                            onValueChange={(idx) => {
                                setSelectedIndex(Number(idx));
                                // Reset value when switching type? Maybe not always desirable but safer.
                                // onChange(path, undefined);
                            }}
                        >
                            <SelectTrigger className="h-7 w-[180px] text-xs">
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
                </CardHeader>
                <CardContent className="p-4 pt-4">
                     <SchemaField
                        path={path}
                        schema={selectedSchema}
                        value={value}
                        onChange={onChange}
                        errors={errors}
                        level={level}
                     />
                </CardContent>
            </Card>
        );
    }

    // --- Enum (Select) ---
    if (schema.enum) {
        return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} />
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
        if (schema.contentEncoding === "base64" || schema.format === "binary") {
             return (
                <div className="w-full">
                    <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} />
                    <FileInput
                        id={path}
                        value={(value as string) || undefined}
                        onChange={(v) => onChange(path, v)}
                        accept={schema.contentMediaType}
                        className={cn(hasError && "border-destructive")}
                    />
                </div>
            );
        }

         return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} />
                <Input
                    id={path}
                    value={(value as string) || ""}
                    onChange={(e) => onChange(path, e.target.value)}
                    placeholder={schema.examples?.[0] || "Enter text"}
                    className={cn(hasError && "border-destructive focus-visible:ring-destructive/20")}
                />
            </div>
        );
    }

    if (type === "integer" || type === "number") {
        return (
            <div className="w-full">
                <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} />
                <Input
                    id={path}
                    type="number"
                    step={type === "integer" ? "1" : "any"}
                    value={(value as number) || ""}
                    onChange={(e) => onChange(path, e.target.value === "" ? undefined : Number(e.target.value))}
                    placeholder="0"
                    className={cn(hasError && "border-destructive focus-visible:ring-destructive/20")}
                />
            </div>
        );
    }

    if (type === "boolean") {
        return (
             <div className="flex items-center justify-between py-2 border rounded-md px-3 bg-muted/10">
                <div className="space-y-0.5">
                    <Label htmlFor={path} className="text-sm font-medium">{label}</Label>
                    {description && <p className="text-[10px] text-muted-foreground">{description}</p>}
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

        // Root level object (level 0) - Render flat or clean
        if (level === 0) {
            return (
                <div className="space-y-5">
                    {hasProperties ? (
                        Object.keys(properties).map((key) => (
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
                        ))
                    ) : (
                        <div className="text-sm text-muted-foreground italic p-4 border border-dashed rounded bg-muted/20 text-center">
                            No properties defined.
                        </div>
                    )}
                </div>
            );
        }

        // Nested Object - Render as Card/Group
        return (
            <Card className={cn("overflow-hidden shadow-none", hasError ? "border-destructive" : "border-border/60")}>
                <CardHeader className="px-4 py-3 bg-muted/30 border-b">
                    <div className="flex items-center gap-2">
                        <Label className="text-sm font-semibold">{label}</Label>
                        {isRequired && <Badge variant="outline" className="text-[10px] h-4 px-1 py-0 border-destructive/50 text-destructive">Required</Badge>}
                    </div>
                    {description && <CardDescription className="text-xs mt-0">{description}</CardDescription>}
                </CardHeader>
                <CardContent className="p-4 space-y-4 bg-card/50">
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
                </CardContent>
            </Card>
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
            <Card className={cn("overflow-hidden shadow-none", hasError ? "border-destructive" : "border-border/60")}>
                 <CardHeader className="px-4 py-3 bg-muted/30 border-b flex flex-row items-center justify-between space-y-0">
                    <div className="space-y-1">
                        <div className="flex items-center gap-2">
                            <Label className="text-sm font-semibold">{label}</Label>
                             <Badge variant="secondary" className="text-[10px] h-4 px-1">List</Badge>
                             {isRequired && <span className="text-destructive">*</span>}
                        </div>
                        {description && <CardDescription className="text-xs">{description}</CardDescription>}
                    </div>
                    <Button type="button" variant="outline" size="sm" onClick={handleAdd} className="h-7 text-xs gap-1 shadow-sm">
                        <Plus className="size-3" /> Add Item
                    </Button>
                </CardHeader>
                <CardContent className="p-0">
                    {arrayValue.length === 0 ? (
                        <div className="text-xs text-muted-foreground italic p-8 text-center bg-muted/5">
                            No items in list.
                        </div>
                    ) : (
                        <div className="divide-y">
                            {arrayValue.map((item, index) => (
                                <div key={index} className="flex gap-3 items-start p-3 hover:bg-muted/10 transition-colors group">
                                     <div className="mt-2 text-muted-foreground/30">
                                         <GripVertical className="size-4" />
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
                                        className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 -mr-1"
                                        title="Remove Item"
                                    >
                                        <Trash2 className="size-4" />
                                    </Button>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        );
    }

    // Fallback
    return (
        <div className="w-full">
            <FieldLabel label={label} required={isRequired} description={description} error={errors?.[path]} />
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
