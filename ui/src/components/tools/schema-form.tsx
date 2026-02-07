/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { Plus, Trash2, HelpCircle } from "lucide-react";
import { Schema } from "./schema-viewer";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface SchemaFormProps {
  schema: Schema;
  value: any;
  onChange: (value: any) => void;
  name?: string;
  required?: boolean;
  depth?: number;
}

export function SchemaForm({ schema, value, onChange, name, required = false, depth = 0 }: SchemaFormProps) {
  // Ensure value is initialized correctly based on type if undefined
  useEffect(() => {
    if (value === undefined) {
      if (schema.default !== undefined) {
        onChange(schema.default);
      } else if (schema.type === "boolean") {
        onChange(false);
      } else if (schema.type === "array") {
        onChange([]);
      } else if (schema.type === "object") {
        onChange({});
      }
    }
  }, [schema, value, onChange]);

  const type = Array.isArray(schema.type) ? schema.type[0] : schema.type;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let newValue: any = e.target.value;
    if (type === "number" || type === "integer") {
      newValue = e.target.value === "" ? undefined : Number(e.target.value);
    }
    onChange(newValue);
  };

  const renderLabel = () => (
    <div className="flex items-center gap-2 mb-2">
      <Label className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>
        {name || "Root"}
      </Label>
      {schema.description && (
        <Tooltip delayDuration={300}>
          <TooltipTrigger asChild>
            <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
          </TooltipTrigger>
          <TooltipContent className="max-w-[300px] text-xs">
            <p>{schema.description}</p>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );

  if (schema.enum) {
    return (
      <div className="space-y-1">
        {name && renderLabel()}
        <Select value={value} onValueChange={onChange}>
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

  switch (type) {
    case "string":
      return (
        <div className="space-y-1">
          {name && renderLabel()}
          <Input
            value={value || ""}
            onChange={handleInputChange}
            placeholder={schema.description || name}
          />
        </div>
      );

    case "number":
    case "integer":
      return (
        <div className="space-y-1">
          {name && renderLabel()}
          <Input
            type="number"
            value={value !== undefined ? value : ""}
            onChange={handleInputChange}
            placeholder={schema.description || name}
          />
        </div>
      );

    case "boolean":
      return (
        <div className="flex items-center justify-between space-y-0 py-2">
          {name && <Label className={cn("text-sm font-medium", required && "after:content-['*'] after:ml-0.5 after:text-red-500")}>{name}</Label>}
          <Switch
            checked={!!value}
            onCheckedChange={onChange}
          />
        </div>
      );

    case "object":
        const properties = schema.properties || {};
        const propertyKeys = Object.keys(properties);

        // If it's a nested object (depth > 0), wrap in a card/border
        const Wrapper = depth > 0 ? Card : React.Fragment;
        const wrapperProps = depth > 0 ? { className: "border-dashed bg-muted/10" } : {};
        const ContentWrapper = depth > 0 ? CardContent : React.Fragment;
        const contentProps = depth > 0 ? { className: "p-4 space-y-4" } : { className: "space-y-4" };

        return (
            <div className="space-y-2">
                {name && renderLabel()}
                {/* @ts-expect-error Wrapper can be Fragment which doesn't accept className */}
                <Wrapper {...wrapperProps}>
                     {/* @ts-expect-error ContentWrapper can be Fragment which doesn't accept className */}
                    <ContentWrapper {...contentProps}>
                        {propertyKeys.map((key) => (
                            <SchemaForm
                                key={key}
                                name={key}
                                schema={properties[key]}
                                value={value ? value[key] : undefined}
                                onChange={(newPropValue) => {
                                    const newObj = { ...value, [key]: newPropValue };
                                    // Clean up undefined values to keep JSON clean?
                                    // Or keep them. Usually undefined is JSON.stringify'd away.
                                    onChange(newObj);
                                }}
                                required={schema.required?.includes(key)}
                                depth={depth + 1}
                            />
                        ))}
                    </ContentWrapper>
                </Wrapper>
            </div>
        );

    case "array":
        const itemsSchema = schema.items;
        if (!itemsSchema) return null;

        const items = Array.isArray(value) ? value : [];

        const addItem = () => {
            // Initialize new item based on schema type
            let newItem = undefined;
             if (itemsSchema.type === "string") newItem = "";
             else if (itemsSchema.type === "number") newItem = 0;
             else if (itemsSchema.type === "boolean") newItem = false;
             else if (itemsSchema.type === "object") newItem = {};
             else if (itemsSchema.type === "array") newItem = [];

            onChange([...items, newItem]);
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
            <div className="space-y-2">
                {name && renderLabel()}
                <div className="space-y-2 pl-2 border-l-2 border-muted">
                    {items.map((item: any, idx: number) => (
                        <div key={idx} className="flex gap-2 items-start">
                            <div className="flex-1">
                                <SchemaForm
                                    schema={itemsSchema}
                                    value={item}
                                    onChange={(val) => updateItem(idx, val)}
                                    depth={depth + 1}
                                />
                            </div>
                            <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => removeItem(idx)}
                                className="h-8 w-8 text-destructive mt-1"
                            >
                                <Trash2 className="h-4 w-4" />
                            </Button>
                        </div>
                    ))}
                    <Button variant="outline" size="sm" onClick={addItem} className="w-full border-dashed">
                        <Plus className="mr-2 h-4 w-4" /> Add Item
                    </Button>
                </div>
            </div>
        );

    default:
      return (
        <div className="p-2 border border-red-200 bg-red-50 text-red-500 rounded text-xs">
          Unsupported type: {type}
        </div>
      );
  }
}
