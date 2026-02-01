/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, ChevronDown, ChevronRight, Info } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: string[] | number[];
  default?: unknown;
  format?: string;
  [key: string]: unknown;
}

interface SchemaFormProps {
  schema: Schema;
  value: unknown;
  onChange: (value: unknown) => void;
  name?: string;
  required?: boolean;
  depth?: number;
  path?: string;
}

const getDefaultValue = (schema: Schema) => {
  if (schema.default !== undefined) return schema.default;
  if (schema.enum && schema.enum.length > 0) return schema.enum[0];

  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
  switch (type) {
    case "string": return "";
    case "number":
    case "integer": return 0;
    case "boolean": return false;
    case "object": return {};
    case "array": return [];
    default: return "";
  }
};

export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0, path = "" }: SchemaFormProps) {
  const [isOpen, setIsOpen] = React.useState(true);
  const currentPath = path ? (name ? `${path}.${name}` : path) : (name || "root");
  const inputId = `form-item-${currentPath.replace(/[^a-zA-Z0-9]/g, '-')}`;

  // Initialize undefined values
  useEffect(() => {
    if (value === undefined) {
      onChange(getDefaultValue(schema));
    }
  }, [schema, value, onChange]);

  if (!schema) return null;

  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
  const isObject = type === "object" || !!schema.properties;
  const isArray = type === "array" || !!schema.items;

  const handleObjectChange = (key: string, val: unknown) => {
    onChange({ ...(value as object || {}), [key]: val });
  };

  const handleArrayAdd = () => {
    const newItem = getDefaultValue(schema.items || {});
    onChange([...(value as unknown[] || []), newItem]);
  };

  const handleArrayRemove = (index: number) => {
    const newValue = [...(value as unknown[] || [])];
    newValue.splice(index, 1);
    onChange(newValue);
  };

  const handleArrayChange = (index: number, val: unknown) => {
    const newValue = [...(value as unknown[] || [])];
    newValue[index] = val;
    onChange(newValue);
  };

  // Render Label with Tooltip
  const renderLabel = (id?: string) => (
    <div className="flex items-center gap-2 mb-2">
      {name && (
        <Label htmlFor={id} className={cn("text-sm font-medium", required && "text-foreground")}>
          {name}
          {required && <span className="text-destructive ml-1">*</span>}
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

  // Object Type
  if (isObject) {
    const properties = schema.properties ? Object.entries(schema.properties) : [];

    // If it's the root object (depth 0), just render fields without wrapper unless it has a name
    if (depth === 0 && !name) {
      return (
        <div className="space-y-4">
          {properties.map(([key, propSchema]) => (
            <SchemaForm
              key={key}
              schema={propSchema}
              value={(value as Record<string, unknown>)?.[key]}
              onChange={(val) => handleObjectChange(key, val)}
              name={key}
              required={schema.required?.includes(key)}
              depth={depth + 1}
              path={currentPath}
            />
          ))}
        </div>
      );
    }

    return (
      <div className={cn("border rounded-md p-3 bg-card/30", depth > 0 && "mt-2")}>
         <Collapsible open={isOpen} onOpenChange={setIsOpen}>
            <div className="flex items-center gap-2">
                <CollapsibleTrigger className="p-1 hover:bg-muted rounded">
                     {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                </CollapsibleTrigger>
                {name ? (
                     <div className="flex flex-col">
                        <span className="text-sm font-semibold">{name}</span>
                        {schema.description && <span className="text-[10px] text-muted-foreground">{schema.description}</span>}
                     </div>
                ) : <span className="text-sm italic text-muted-foreground">Object properties</span>}
            </div>
            <CollapsibleContent className="pl-4 pt-3 space-y-4 border-l ml-2.5 mt-2 border-border/50">
                 {properties.map(([key, propSchema]) => (
                    <SchemaForm
                      key={key}
                      schema={propSchema}
                      value={(value as Record<string, unknown>)?.[key]}
                      onChange={(val) => handleObjectChange(key, val)}
                      name={key}
                      required={schema.required?.includes(key)}
                      depth={depth + 1}
                      path={currentPath}
                    />
                  ))}
            </CollapsibleContent>
         </Collapsible>
      </div>
    );
  }

  // Array Type
  if (isArray) {
    return (
      <div className={cn("space-y-2", depth > 0 && "mt-2")}>
        {renderLabel(inputId)}
        <div className="border rounded-md p-3 bg-muted/20">
            {Array.isArray(value) && value.map((item: unknown, index: number) => (
                <div key={index} className="flex items-start gap-2 mb-3 last:mb-0">
                    <div className="flex-1">
                        <SchemaForm
                            schema={schema.items || {}}
                            value={item}
                            onChange={(val) => handleArrayChange(index, val)}
                            depth={depth + 1}
                            path={`${currentPath}.${index}`}
                        />
                    </div>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => handleArrayRemove(index)}
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            <Button
                variant="outline"
                size="sm"
                className="w-full mt-2 border-dashed text-xs h-8"
                onClick={handleArrayAdd}
            >
                <Plus className="mr-2 h-3 w-3" /> Add Item
            </Button>
        </div>
      </div>
    );
  }

  // Enum Type
  if (schema.enum) {
     return (
        <div className={cn("space-y-1.5", depth > 0 && "mt-2")}>
            {renderLabel(inputId)}
            <Select value={value as string} onValueChange={onChange}>
                <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select option" />
                </SelectTrigger>
                <SelectContent>
                    {schema.enum.map((opt: unknown) => (
                        <SelectItem key={String(opt)} value={String(opt)}>
                            {String(opt)}
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
        </div>
     );
  }

  // Boolean Type
  if (type === "boolean") {
      return (
          <div className="flex items-center justify-between py-2 border rounded-md px-3 bg-background/50">
              <div className="flex flex-col gap-0.5">
                  <Label htmlFor={inputId} className={cn("text-sm cursor-pointer", required && "text-foreground")}>
                      {name}
                      {required && <span className="text-destructive ml-1">*</span>}
                  </Label>
                  {schema.description && <span className="text-[10px] text-muted-foreground">{schema.description}</span>}
              </div>
              <Switch
                  id={inputId}
                  checked={!!value}
                  onCheckedChange={onChange}
              />
          </div>
      );
  }

  // Number/Integer
  if (type === "number" || type === "integer") {
      return (
          <div className={cn("space-y-1.5", depth > 0 && "mt-2")}>
              {renderLabel(inputId)}
              <Input
                  id={inputId}
                  type="number"
                  value={value as number}
                  onChange={(e) => onChange(e.target.value === "" ? "" : Number(e.target.value))}
                  placeholder={String(schema.default ?? "")}
                  className="font-mono text-sm"
              />
          </div>
      );
  }

  // String (Default)
  return (
      <div className={cn("space-y-1.5", depth > 0 && "mt-2")}>
          {renderLabel(inputId)}
          <Input
              id={inputId}
              type="text"
              value={value as string || ""}
              onChange={(e) => onChange(e.target.value)}
              placeholder={String(schema.default ?? "")}
              className="font-sans text-sm"
          />
      </div>
  );
}
