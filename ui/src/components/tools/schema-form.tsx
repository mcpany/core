/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Trash2, Plus, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useState } from "react";

// Helper for Number Inputs to handle "1." vs 1 state
const NumberInput = ({ value, onChange, placeholder }: { value: any, onChange: (val: any) => void, placeholder?: string }) => {
  const [inputValue, setInputValue] = useState(value?.toString() ?? "");

  useEffect(() => {
    if (value === undefined) {
      setInputValue("");
    } else {
      // Avoid resetting if the parsed input matches the value (handling "1." vs 1)
      const currentParsed = parseFloat(inputValue);
      // Compare roughly to handle float precision if needed, but strict for now
      if (currentParsed !== value && !isNaN(Number(value))) {
        setInputValue(String(value));
      }
    }
  }, [value]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value;
    setInputValue(raw);

    const num = parseFloat(raw);
    if (!isNaN(num)) {
      onChange(num);
    } else if (raw === "") {
      onChange(undefined);
    }
  };

  return (
    <Input
      type="number"
      value={inputValue}
      onChange={handleChange}
      placeholder={placeholder}
    />
  );
};

/**
 * Schema represents a JSON Schema object used for defining tool input parameters.
 */
export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: any[];
  default?: any;
  format?: string;
  [key: string]: any;
}

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  path?: string[];
}

/**
 * SchemaForm component generates a dynamic form based on a JSON Schema.
 * It supports nested objects, arrays, and basic types (string, number, boolean).
 *
 * @param props - The component props.
 * @param props.schema - The JSON schema definition.
 * @param props.value - The current value of the form.
 * @param props.onChange - Callback when value changes.
 * @param props.name - The name of the field (optional).
 * @param props.required - Whether the field is required.
 * @param props.path - The path to the current field (for debugging/nesting).
 * @returns The rendered form component.
 */
export function SchemaForm({ schema, value, onChange, name, required = false, path = [] }: SchemaFormProps) {
  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  const handleChange = (newValue: any) => {
    onChange(newValue);
  };

  // Ensure value matches type (basic sanitization)
  // If value is undefined, we might want to set default if provided
  useEffect(() => {
    if (value === undefined && schema.default !== undefined) {
      onChange(schema.default);
    }
  }, [value, schema.default]); // Removed onChange to avoid infinite loop if onChange identity changes

  if (!schema) return null;

  // Render Label with Tooltip
  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-2">
      <Label className={cn(required && "font-semibold")}>
        {name || "Root"}
        {required && <span className="text-red-500 ml-1">*</span>}
      </Label>
      {schema.description && (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>
            <Info className="h-3 w-3 text-muted-foreground/70 hover:text-foreground transition-colors cursor-help" />
          </TooltipTrigger>
          <TooltipContent className="max-w-[300px] text-xs">
            <p>{schema.description}</p>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  // String / Number / Integer
  if (type === "string" || type === "number" || type === "integer") {
    if (schema.enum) {
       return (
        <div className="grid gap-1.5">
          {renderLabel()}
          <Select
            value={value !== undefined ? String(value) : ""}
            onValueChange={(val) => {
                if (type === "number" || type === "integer") {
                    handleChange(Number(val));
                } else {
                    handleChange(val);
                }
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

    return (
      <div className="grid gap-1.5">
        {renderLabel()}
        {type === "number" || type === "integer" ? (
          <NumberInput
            value={value}
            onChange={handleChange}
            placeholder={schema.default ? String(schema.default) : ""}
          />
        ) : (
          <Input
            type="text"
            value={value !== undefined ? value : ""}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={schema.default ? String(schema.default) : ""}
          />
        )}
      </div>
    );
  }

  // Boolean
  if (type === "boolean") {
    return (
      <div className="flex items-center justify-between rounded-lg border p-3 shadow-sm">
        <div className="space-y-0.5">
          {renderLabel()}
        </div>
        <Switch
          checked={!!value}
          onCheckedChange={handleChange}
        />
      </div>
    );
  }

  // Object
  if (type === "object") {
    const properties = schema.properties || {};
    // If it's the root object (no name), don't wrap in Card unless forced
    const Wrapper = name ? Card : "div";
    const wrapperProps = name ? { className: cn("border-dashed mt-4") } : { className: "space-y-4" };

    return (
      // @ts-ignore
      <Wrapper {...wrapperProps}>
        {name && (
            <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    {name}
                    {schema.description && <span className="text-xs font-normal text-muted-foreground">- {schema.description}</span>}
                </CardTitle>
            </CardHeader>
        )}
        <div className={name ? "p-4 space-y-4" : "space-y-4"}>
          {Object.entries(properties).map(([key, propSchema]) => (
            <SchemaForm
              key={key}
              name={key}
              schema={propSchema}
              value={value ? value[key] : undefined}
              onChange={(newPropValue) => {
                 const newValue = { ...value, [key]: newPropValue };
                 handleChange(newValue);
              }}
              required={schema.required?.includes(key)}
              path={[...path, key]}
            />
          ))}
        </div>
      </Wrapper>
    );
  }

  // Array
  if (type === "array") {
      const items = schema.items;
      if (!items) return null;

      const currentList = Array.isArray(value) ? value : [];

      const addItem = () => {
          // Add default or empty based on item type
          let newItem = undefined;
          if (items.type === "string") newItem = "";
          if (items.type === "number") newItem = 0;
          if (items.type === "boolean") newItem = false;
          if (items.type === "object") newItem = {};
          if (items.type === "array") newItem = [];
          if (items.default !== undefined) newItem = items.default;

          handleChange([...currentList, newItem]);
      };

      const removeItem = (index: number) => {
          const newList = [...currentList];
          newList.splice(index, 1);
          handleChange(newList);
      };

      const updateItem = (index: number, val: any) => {
          const newList = [...currentList];
          newList[index] = val;
          handleChange(newList);
      };

      return (
        <Card className="border-dashed mt-4">
             <CardHeader className="py-3 px-4 bg-muted/20 border-b flex flex-row items-center justify-between">
                <div className="space-y-0.5">
                    <CardTitle className="text-sm font-medium">{name || "List"}</CardTitle>
                    {schema.description && <p className="text-xs text-muted-foreground">{schema.description}</p>}
                </div>
                <Button size="sm" variant="outline" onClick={addItem} type="button">
                    <Plus className="h-4 w-4 mr-1" /> Add Item
                </Button>
            </CardHeader>
            <CardContent className="p-4 space-y-4">
                {currentList.length === 0 && (
                    <div className="text-sm text-muted-foreground text-center py-4 italic">
                        No items in list.
                    </div>
                )}
                {currentList.map((item: any, idx: number) => (
                    <div key={idx} className="flex items-start gap-2 p-2 border rounded bg-background/50">
                        <div className="flex-1">
                             <SchemaForm
                                schema={items}
                                value={item}
                                onChange={(val) => updateItem(idx, val)}
                                path={[...path, String(idx)]}
                             />
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => removeItem(idx)}
                            className="text-destructive hover:text-destructive mt-6"
                            title="Remove Item"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
                ))}
            </CardContent>
        </Card>
      );
  }

  return (
    <div className="text-xs text-red-500">
      Unsupported type: {String(type)}
    </div>
  );
}
