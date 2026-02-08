/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { Schema } from "./schema-viewer";
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
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ChevronRight, ChevronDown, Plus, Trash2, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  depth?: number;
  label?: string;
}

/**
 * SchemaForm component.
 * Renders a dynamic form based on a JSON schema.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange, name, required, depth = 0, label }: SchemaFormProps) {
  const [isOpen, setIsOpen] = useState(true);

  // Helper to update field in object
  const handleFieldChange = (key: string, newValue: any) => {
    const updated = { ...(value || {}) };
    if (newValue === undefined) {
      delete updated[key];
    } else {
      updated[key] = newValue;
    }
    onChange(updated);
  };

  // Helper for array changes
  const handleArrayChange = (index: number, newValue: any) => {
    const updated = [...(value || [])];
    updated[index] = newValue;
    onChange(updated);
  };

  const handleArrayAdd = () => {
    const updated = [...(value || [])];
    // Add default value based on item schema
    updated.push(getDefaultValue(schema.items));
    onChange(updated);
  };

  const handleArrayRemove = (index: number) => {
    const updated = [...(value || [])];
    updated.splice(index, 1);
    onChange(updated);
  };

  const getDefaultValue = (s?: Schema): any => {
    if (!s) return "";
    if (s.default !== undefined) return s.default;
    if (s.type === "string") return "";
    if (s.type === "number" || s.type === "integer") return 0;
    if (s.type === "boolean") return false;
    if (s.type === "object") return {};
    if (s.type === "array") return [];
    return "";
  };

  if (!schema) return null;

  const displayLabel = label || name;
  const description = schema.description;

  // STRING
  if (schema.type === "string") {
    // Enum
    if (schema.enum) {
       return (
        <div className="space-y-2">
            <Label className="text-sm font-medium flex items-center gap-2">
                {displayLabel}
                {required && <span className="text-red-500">*</span>}
                {description && (
                    <Tooltip delayDuration={300}>
                        <TooltipTrigger asChild>
                            <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                        </TooltipTrigger>
                        <TooltipContent className="max-w-[300px] text-xs">
                            <p>{description}</p>
                        </TooltipContent>
                    </Tooltip>
                )}
            </Label>
            <Select
                value={value || ""}
                onValueChange={(v) => onChange(v)}
            >
                <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select..." />
                </SelectTrigger>
                <SelectContent>
                    {schema.enum.map((opt: string) => (
                        <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                    ))}
                </SelectContent>
            </Select>
        </div>
       )
    }

    return (
      <div className="space-y-2">
        <Label className="text-sm font-medium flex items-center gap-2">
            {displayLabel}
            {required && <span className="text-red-500">*</span>}
            {description && (
                <Tooltip delayDuration={300}>
                    <TooltipTrigger asChild>
                        <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent className="max-w-[300px] text-xs">
                        <p>{description}</p>
                    </TooltipContent>
                </Tooltip>
            )}
        </Label>
        <Input
          value={value || ""}
          onChange={(e) => onChange(e.target.value)}
          placeholder={String(schema.default || description || "")}
        />
      </div>
    );
  }

  // NUMBER / INTEGER
  if (schema.type === "number" || schema.type === "integer") {
    return (
      <div className="space-y-2">
        <Label className="text-sm font-medium flex items-center gap-2">
            {displayLabel}
            {required && <span className="text-red-500">*</span>}
            {description && (
                <Tooltip delayDuration={300}>
                    <TooltipTrigger asChild>
                        <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent className="max-w-[300px] text-xs">
                        <p>{description}</p>
                    </TooltipContent>
                </Tooltip>
            )}
        </Label>
        <Input
          type="number"
          value={value || ""}
          onChange={(e) => onChange(Number(e.target.value))}
          placeholder={String(schema.default || "0")}
        />
      </div>
    );
  }

  // BOOLEAN
  if (schema.type === "boolean") {
    return (
       <div className="flex items-center justify-between space-x-2 border p-3 rounded-md">
           <Label className="flex flex-col space-y-1">
               <span className="text-sm font-medium flex items-center gap-2">
                   {displayLabel}
                   {required && <span className="text-red-500">*</span>}
                   {description && (
                        <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                                <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-[300px] text-xs">
                                <p>{description}</p>
                            </TooltipContent>
                        </Tooltip>
                    )}
               </span>
               {description && <span className="text-xs text-muted-foreground">{description}</span>}
           </Label>
           <Switch
                checked={!!value}
                onCheckedChange={onChange}
           />
       </div>
    );
  }

  // OBJECT
  if (schema.type === "object" || schema.properties) {
      const properties = schema.properties || {};
      const hasProps = Object.keys(properties).length > 0;

      if (!hasProps) return null;

      // Root object (depth 0) doesn't need a collapsible wrapper if it's the main container
      // but recursive calls do.
      if (depth === 0 && !name) {
          return (
            <div className="space-y-4 pt-2">
                {Object.entries(properties).map(([key, propSchema]) => (
                    <SchemaForm
                        key={key}
                        name={key}
                        schema={propSchema}
                        value={value?.[key]}
                        onChange={(v) => handleFieldChange(key, v)}
                        required={schema.required?.includes(key)}
                        depth={depth + 1}
                    />
                ))}
            </div>
          );
      }

      return (
          <Collapsible
             open={isOpen}
             onOpenChange={setIsOpen}
             className={cn("space-y-2", depth > 0 && "ml-2 border-l pl-4")}
          >
              <div className="flex items-center justify-between">
                  <CollapsibleTrigger asChild>
                      <Button variant="ghost" size="sm" className="p-0 h-auto hover:bg-transparent justify-start font-semibold">
                          {isOpen ? <ChevronDown className="h-4 w-4 mr-2" /> : <ChevronRight className="h-4 w-4 mr-2" />}
                          {displayLabel || "Parameters"}
                      </Button>
                  </CollapsibleTrigger>
              </div>

              <CollapsibleContent className="space-y-4 pt-2">
                  {Object.entries(properties).map(([key, propSchema]) => (
                      <SchemaForm
                        key={key}
                        name={key}
                        schema={propSchema}
                        value={value?.[key]}
                        onChange={(v) => handleFieldChange(key, v)}
                        required={schema.required?.includes(key)}
                        depth={depth + 1}
                      />
                  ))}
              </CollapsibleContent>
          </Collapsible>
      );
  }

  // ARRAY
  if (schema.type === "array" && schema.items) {
      return (
          <div className={cn("space-y-2", depth > 0 && "ml-2 border-l pl-4")}>
              <Label className="text-sm font-medium flex items-center gap-2">
                  {displayLabel}
                  {required && <span className="text-red-500">*</span>}
                  <span className="text-xs text-muted-foreground">({(value || []).length} items)</span>
              </Label>

              <div className="space-y-2">
                  {(value || []).map((item: any, index: number) => (
                      <div key={index} className="flex items-start gap-2">
                          <div className="flex-1">
                              <SchemaForm
                                  schema={schema.items!}
                                  value={item}
                                  onChange={(v) => handleArrayChange(index, v)}
                                  depth={depth + 1}
                                  label={`Item ${index + 1}`}
                              />
                          </div>
                          <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-destructive mt-8"
                              onClick={() => handleArrayRemove(index)}
                          >
                              <Trash2 className="h-4 w-4" />
                          </Button>
                      </div>
                  ))}
                  <Button variant="outline" size="sm" onClick={handleArrayAdd} className="w-full border-dashed">
                      <Plus className="h-3 w-3 mr-2" /> Add Item
                  </Button>
              </div>
          </div>
      );
  }

  return (
      <div className="text-xs text-muted-foreground p-2 border border-dashed rounded">
          Unsupported type: {schema.type}
      </div>
  );
}
