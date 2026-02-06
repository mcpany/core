/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
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
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { Info } from "lucide-react";
import { cn } from "@/lib/utils";

interface SchemaProperty {
  type?: string;
  description?: string;
  enum?: string[];
  default?: any;
  items?: SchemaProperty;
  properties?: Record<string, SchemaProperty>;
  required?: string[];
  [key: string]: any;
}

interface DynamicFormProps {
  schema: SchemaProperty;
  value: Record<string, any>;
  onChange: (value: Record<string, any>) => void;
  className?: string;
}

/**
 * DynamicForm generates a form based on a JSON Schema.
 */
export function DynamicForm({ schema, value, onChange, className }: DynamicFormProps) {
  // Ensure value is an object
  const safeValue = value || {};

  const handleFieldChange = (key: string, val: any) => {
    const newValue = { ...safeValue, [key]: val };
    // Remove key if value is empty/undefined (optional, but keeps JSON clean)
    // However, for controlled inputs, we usually want to keep the key.
    onChange(newValue);
  };

  const renderField = (key: string, propSchema: SchemaProperty, required: boolean) => {
    const fieldId = `field-${key}`;
    const currentValue = safeValue[key] ?? propSchema.default ?? "";

    const label = (
        <div className="flex items-center gap-2 mb-1.5">
            <Label htmlFor={fieldId} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
                {key}
            </Label>
            {propSchema.description && (
                <TooltipProvider>
                    <Tooltip delayDuration={300}>
                        <TooltipTrigger asChild>
                            <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help opacity-70 hover:opacity-100 transition-opacity" />
                        </TooltipTrigger>
                        <TooltipContent className="max-w-xs text-xs">
                            <p>{propSchema.description}</p>
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            )}
        </div>
    );

    // Enum (Select)
    if (propSchema.enum) {
      return (
        <div key={key} className="space-y-1">
          {label}
          <Select
            value={currentValue}
            onValueChange={(v) => handleFieldChange(key, v)}
          >
            <SelectTrigger id={fieldId} className="w-full">
              <SelectValue placeholder="Select..." />
            </SelectTrigger>
            <SelectContent>
              {propSchema.enum.map((opt) => (
                <SelectItem key={opt} value={opt}>
                  {opt}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      );
    }

    // Boolean (Switch)
    if (propSchema.type === "boolean") {
        return (
            <div key={key} className="flex items-center justify-between rounded-lg border p-3 shadow-sm bg-muted/20">
                <div className="space-y-0.5">
                    <Label htmlFor={fieldId} className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
                        {key}
                    </Label>
                    {propSchema.description && (
                        <p className="text-[10px] text-muted-foreground">
                            {propSchema.description}
                        </p>
                    )}
                </div>
                <Switch
                    id={fieldId}
                    checked={currentValue === true}
                    onCheckedChange={(checked) => handleFieldChange(key, checked)}
                />
            </div>
        );
    }

    // Number
    if (propSchema.type === "integer" || propSchema.type === "number") {
        return (
            <div key={key} className="space-y-1">
                {label}
                <Input
                    id={fieldId}
                    type="number"
                    value={currentValue}
                    onChange={(e) => {
                        const val = e.target.value === "" ? "" : Number(e.target.value);
                        handleFieldChange(key, val);
                    }}
                    placeholder={`Enter ${propSchema.type}...`}
                />
            </div>
        );
    }

    // Default: String
    return (
      <div key={key} className="space-y-1">
        {label}
        <Input
            id={fieldId}
            type="text"
            value={currentValue}
            onChange={(e) => handleFieldChange(key, e.target.value)}
            placeholder={propSchema.default ? `Default: ${propSchema.default}` : ""}
        />
      </div>
    );
  };

  if (!schema || !schema.properties) {
      return <div className="text-sm text-muted-foreground p-4 text-center italic">No configurable properties found.</div>;
  }

  const requiredFields = new Set(schema.required || []);

  return (
    <div className={cn("space-y-4 p-1", className)}>
      {Object.entries(schema.properties).map(([key, prop]) =>
        renderField(key, prop, requiredFields.has(key))
      )}
    </div>
  );
}
