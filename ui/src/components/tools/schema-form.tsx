/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Plus, Trash2, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

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
  root?: boolean;
  name?: string;
  required?: boolean;
}

/**
 * A recursive form builder that generates UI based on a JSON Schema.
 */
export function SchemaForm({ schema, value, onChange, root = false, name, required = false }: SchemaFormProps) {
  // Determine type
  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  // Handle default value on mount if value is undefined
  useEffect(() => {
    if (value === undefined && schema.default !== undefined) {
      onChange(schema.default);
    }
  }, []);

  const description = schema.description || schema.title;

  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-2">
      <Label className={cn(root ? "text-base font-semibold" : "text-sm font-medium")}>
        {name || "Root"}
        {required && <span className="text-red-500 ml-1">*</span>}
      </Label>
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
    </div>
  );

  // Handle Enum
  if (schema.enum) {
    return (
      <div className="space-y-1 mb-4">
        {name && renderLabel()}
        <Select
          value={value !== undefined ? String(value) : ""}
          onValueChange={(val) => {
            // Try to convert back to original type if number/boolean
            if (typeof schema.enum![0] === 'number') onChange(Number(val));
            else if (typeof schema.enum![0] === 'boolean') onChange(val === 'true');
            else onChange(val);
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select an option" />
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

  // Handle Boolean
  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between space-x-2 mb-4 p-2 rounded-md border bg-muted/20">
        <div className="flex flex-col space-y-1">
            <Label className="text-sm font-medium">
                {name || "Enabled"}
                {required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {description && <span className="text-xs text-muted-foreground">{description}</span>}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={onChange}
        />
      </div>
    );
  }

  // Handle Object
  if (type === "object" || schema.properties) {
    const properties = schema.properties || {};
    const requiredFields = schema.required || [];

    if (root) {
        return (
            <div className="space-y-4">
                {Object.entries(properties).map(([key, propSchema]) => (
                    <SchemaForm
                        key={key}
                        name={key}
                        schema={propSchema}
                        value={value?.[key]}
                        onChange={(val) => onChange({ ...value, [key]: val })}
                        required={requiredFields.includes(key)}
                    />
                ))}
            </div>
        );
    }

    return (
      <Accordion type="single" collapsible className="w-full mb-4 border rounded-md bg-muted/10">
        <AccordionItem value="item-1" className="border-0">
          <AccordionTrigger className="px-4 py-2 hover:no-underline">
            <div className="flex items-center gap-2">
                <span className="font-medium">{name || "Object"}</span>
                {required && <span className="text-red-500 text-xs">*</span>}
                <span className="text-xs text-muted-foreground font-normal">({Object.keys(properties).length} fields)</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="px-4 pb-4 pt-2 border-t">
             <div className="space-y-4 mt-2">
                {Object.entries(properties).map(([key, propSchema]) => (
                    <SchemaForm
                        key={key}
                        name={key}
                        schema={propSchema}
                        value={value?.[key]}
                        onChange={(val) => onChange({ ...value, [key]: val })}
                        required={requiredFields.includes(key)}
                    />
                ))}
            </div>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    );
  }

  // Handle Array
  if (type === "array") {
    const items = value || [];
    const itemSchema = schema.items || {};

    const addItem = () => {
      onChange([...items, itemSchema.default !== undefined ? itemSchema.default : (itemSchema.type === "object" ? {} : "")]);
    };

    const removeItem = (index: number) => {
      const newItems = [...items];
      newItems.splice(index, 1);
      onChange(newItems);
    };

    const updateItem = (index: number, val: any) => {
      const newItems = [...items];
      newItems[index] = val;
      onChange(newItems);
    };

    return (
      <div className="space-y-2 mb-4">
        {name && renderLabel()}
        <div className="space-y-2">
            {items.map((item: any, idx: number) => (
                <div key={idx} className="flex gap-2 items-start">
                    <div className="flex-1">
                        <SchemaForm
                            schema={itemSchema}
                            value={item}
                            onChange={(val) => updateItem(idx, val)}
                        />
                    </div>
                    <Button variant="ghost" size="icon" onClick={() => removeItem(idx)} className="h-9 w-9 text-destructive hover:text-destructive hover:bg-destructive/10">
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            <Button variant="outline" size="sm" onClick={addItem} className="w-full border-dashed">
                <Plus className="h-3 w-3 mr-1" /> Add Item
            </Button>
        </div>
      </div>
    );
  }

  // Handle String/Number/Integer
  return (
    <div className="space-y-1 mb-4">
      {name && renderLabel()}
      {type === "integer" || type === "number" ? (
        <Input
          type="number"
          value={value ?? ""}
          onChange={(e) => {
             const val = e.target.value;
             if (val === "") onChange(undefined);
             else onChange(type === "integer" ? parseInt(val) : parseFloat(val));
          }}
          placeholder={schema.default ? String(schema.default) : "0"}
        />
      ) : (
          schema.format === "date-time" || schema.format === "date" ? (
              <Input
                type={schema.format === "date" ? "date" : "datetime-local"}
                value={value ?? ""}
                onChange={(e) => onChange(e.target.value)}
              />
          ) : (
             <Input
                value={value ?? ""}
                onChange={(e) => onChange(e.target.value)}
                placeholder={schema.default || (description ? `Enter ${name}...` : "")}
              />
          )
      )}
    </div>
  );
}
