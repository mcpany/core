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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { Plus, Trash2, Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

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
  depth?: number;
}

function NumberInput({ value, onChange, placeholder }: { value: number | undefined, onChange: (val: number | undefined) => void, placeholder?: string }) {
  const [localValue, setLocalValue] = useState(value?.toString() ?? "");

  // Sync with external value changes
  useEffect(() => {
    if (value === undefined) {
      if (localValue !== "") setLocalValue("");
      return;
    }
    const currentNum = parseFloat(localValue);
    // Update local value only if the parsed local value differs from the prop value
    // This allows intermediate states like "1." to persist without being overwritten by "1"
    if (currentNum !== value) {
      setLocalValue(value.toString());
    }
  }, [value, localValue]);

  return (
    <Input
      type="number"
      value={localValue}
      onChange={(e) => {
        const val = e.target.value;
        setLocalValue(val);
        if (val === "") {
          onChange(undefined);
        } else {
          const num = Number(val);
          if (!isNaN(num)) {
             onChange(num);
          }
        }
      }}
      placeholder={placeholder}
    />
  );
}

export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0 }: SchemaFormProps) {
  // Helper to update object properties
  const updateProperty = (key: string, val: any) => {
    const newValue = { ...value, [key]: val };
    // If val is undefined/empty string and optional, maybe remove key?
    // For now, keep it simple.
    onChange(newValue);
  };

  // Helper to update array items
  const updateArrayItem = (index: number, val: any) => {
    const newArray = [...(value || [])];
    newArray[index] = val;
    onChange(newArray);
  };

  const addArrayItem = () => {
    const newArray = [...(value || [])];
    // Default value based on items schema
    let defaultValue: any = "";
    if (schema.items?.type === "object") defaultValue = {};
    if (schema.items?.type === "boolean") defaultValue = false;
    if (schema.items?.type === "number") defaultValue = 0;

    newArray.push(defaultValue);
    onChange(newArray);
  };

  const removeArrayItem = (index: number) => {
    const newArray = [...(value || [])];
    newArray.splice(index, 1);
    onChange(newArray);
  };

  // Render Label with Tooltip
  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-1.5">
      {name && (
        <Label className={cn("text-sm font-medium", required && "text-foreground")}>
          {name}
          {required && <span className="text-destructive ml-0.5">*</span>}
        </Label>
      )}
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

  // Type: Object
  if (schema.type === "object" || schema.properties) {
    const properties = schema.properties || {};
    // Ensure value is object
    const objectValue = value && typeof value === "object" && !Array.isArray(value) ? value : {};

    // If depth > 0, wrap in a card or visual container
    const Container = depth > 0 ? Card : React.Fragment;
    const containerProps = depth > 0 ? { className: "bg-muted/10 border-muted-foreground/20" } : {};
    const Content = depth > 0 ? CardContent : React.Fragment;
    const contentProps = depth > 0 ? { className: "p-4 space-y-4" } : {};

    return (
      <div className={cn("space-y-2", depth > 0 && "mt-2")}>
        {name && renderLabel()}
        <Container {...containerProps}>
          {/* @ts-ignore */}
          <Content {...contentProps}>
            <div className={cn("space-y-4", depth === 0 && "mt-1")}>
              {Object.entries(properties).map(([key, propSchema]) => (
                <SchemaForm
                  key={key}
                  name={key}
                  schema={propSchema}
                  value={objectValue[key]}
                  onChange={(val) => updateProperty(key, val)}
                  required={schema.required?.includes(key)}
                  depth={depth + 1}
                />
              ))}
            </div>
          </Content>
        </Container>
      </div>
    );
  }

  // Type: Array
  if (schema.type === "array" || schema.items) {
    const arrayValue = Array.isArray(value) ? value : [];

    return (
      <div className="space-y-2">
        {name && renderLabel()}
        <div className="space-y-2 pl-2 border-l-2 border-muted">
          {arrayValue.map((item: any, index: number) => (
            <div key={index} className="flex gap-2 items-start">
              <div className="flex-1">
                <SchemaForm
                  schema={schema.items || {}}
                  value={item}
                  onChange={(val) => updateArrayItem(index, val)}
                  depth={depth + 1}
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10 mt-1"
                onClick={() => removeArrayItem(index)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
          <Button
            variant="outline"
            size="sm"
            onClick={addArrayItem}
            className="w-full border-dashed"
          >
            <Plus className="h-3 w-3 mr-1" /> Add Item
          </Button>
        </div>
      </div>
    );
  }

  // Type: Enum (Select)
  if (schema.enum) {
    return (
      <div className="space-y-1.5">
        {name && renderLabel()}
        <Select value={value} onValueChange={onChange}>
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

  // Type: Boolean (Switch)
  if (schema.type === "boolean") {
    return (
      <div className="flex items-center justify-between space-x-2 border rounded-md p-3 bg-muted/20">
        <div className="space-y-0.5">
           {name && <Label className={cn("text-sm font-medium", required && "text-foreground")}>{name}{required && "*"}</Label>}
           {schema.description && <p className="text-[10px] text-muted-foreground">{schema.description}</p>}
        </div>
        <Switch checked={!!value} onCheckedChange={onChange} />
      </div>
    );
  }

  // Type: Number / Integer
  if (schema.type === "number" || schema.type === "integer") {
    return (
      <div className="space-y-1.5">
        {name && renderLabel()}
        <NumberInput
          value={value}
          onChange={onChange}
          placeholder={String(schema.default ?? "")}
        />
      </div>
    );
  }

  // Type: String (Default)
  return (
    <div className="space-y-1.5">
      {name && renderLabel()}
      <Input
        type="text"
        value={value ?? ""}
        onChange={(e) => onChange(e.target.value)}
        placeholder={String(schema.default ?? "")}
      />
    </div>
  );
}
