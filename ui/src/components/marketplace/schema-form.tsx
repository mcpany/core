/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, ChevronDown, ChevronRight } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

interface SchemaFormProps {
  schema: any;
  value: any;
  onChange: (value: any) => void;
  // If true, renders the top-level container style (border, etc.)
  root?: boolean;
}

/**
 * SchemaForm component that recursively renders a JSON Schema form.
 * Supports: string, boolean, enum, object (nested), array (nested).
 */
export function SchemaForm({ schema, value, onChange, root = true }: SchemaFormProps) {
  if (!schema) return null;

  // --- OBJECT TYPE ---
  if (schema.type === "object" && schema.properties) {
    const currentVal = value || {};

    const content = (
        <div className="space-y-4">
            {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
                const isRequired = schema.required?.includes(key);
                const description = prop.description || "";
                const title = prop.title || key;

                return (
                    <div key={key} className="space-y-2">
                        {/* Only show label for non-object/array or if they don't handle their own headers well */}
                        {prop.type !== "object" && prop.type !== "array" && (
                            <Label className="flex items-center gap-1">
                                {title} {isRequired && <span className="text-destructive">*</span>}
                            </Label>
                        )}

                        <SchemaForm
                            schema={prop}
                            value={currentVal[key]}
                            onChange={(newVal) => onChange({ ...currentVal, [key]: newVal })}
                            root={false}
                        />

                        {prop.type !== "object" && prop.type !== "array" && description && (
                            <p className="text-xs text-muted-foreground">{description}</p>
                        )}
                    </div>
                );
            })}
        </div>
    );

    if (root) {
        return (
            <div className="space-y-4 border rounded-lg p-4 bg-muted/20">
                {content}
            </div>
        );
    }

    // If it's a nested object, wrap it in a card/fieldset style
    if (schema.title) {
        return (
            <Card className="border-l-4 border-l-primary/20">
                <CardHeader className="py-3 px-4 bg-muted/10">
                    <Label className="font-semibold text-sm">{schema.title}</Label>
                    {schema.description && <p className="text-xs text-muted-foreground">{schema.description}</p>}
                </CardHeader>
                <CardContent className="p-4">
                    {content}
                </CardContent>
            </Card>
        );
    }

    return content;
  }

  // --- ARRAY TYPE ---
  if (schema.type === "array" && schema.items) {
      const currentVal = Array.isArray(value) ? value : [];

      const addItem = () => {
          // If items is a string type, default to ""
          // If object, default to {}
          let defaultItem = "";
          if (schema.items.type === "object") defaultItem = {};
          if (schema.items.type === "boolean") defaultItem = false as any;
          if (schema.items.type === "integer" || schema.items.type === "number") defaultItem = 0 as any;

          onChange([...currentVal, defaultItem]);
      };

      const removeItem = (index: number) => {
          onChange(currentVal.filter((_, i) => i !== index));
      };

      const updateItem = (index: number, newVal: any) => {
          const newArr = [...currentVal];
          newArr[index] = newVal;
          onChange(newArr);
      };

      return (
          <div className="space-y-2">
              <div className="flex items-center justify-between">
                  <Label>{schema.title || "Items"}</Label>
                  <Button type="button" variant="outline" size="sm" onClick={addItem} className="h-7 text-xs">
                      <Plus className="mr-1 h-3 w-3" /> Add
                  </Button>
              </div>
              <div className="space-y-2">
                  {currentVal.map((item: any, idx: number) => (
                      <div key={idx} className="flex gap-2 items-start group">
                          <div className="flex-1">
                              <SchemaForm
                                  schema={schema.items}
                                  value={item}
                                  onChange={(v) => updateItem(idx, v)}
                                  root={false}
                              />
                          </div>
                          <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => removeItem(idx)}
                              className="h-9 w-9 text-muted-foreground hover:text-destructive shrink-0 mt-0.5"
                          >
                              <Trash2 className="h-4 w-4" />
                          </Button>
                      </div>
                  ))}
                  {currentVal.length === 0 && (
                      <div className="text-xs text-muted-foreground italic p-2 border border-dashed rounded text-center">
                          No items added.
                      </div>
                  )}
              </div>
          </div>
      );
  }

  // --- PRIMITIVES ---

  // Enum
  if (schema.enum) {
      return (
        <Select value={value} onValueChange={onChange}>
            <SelectTrigger>
                <SelectValue placeholder="Select..." />
            </SelectTrigger>
            <SelectContent>
                {schema.enum.map((opt: string) => (
                    <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                ))}
            </SelectContent>
        </Select>
      );
  }

  // Boolean
  if (schema.type === "boolean") {
       return (
        <div className="flex items-center space-x-2 h-9">
            <Checkbox
                id={`chk-${Math.random()}`} // simplistic id
                checked={value === true}
                onCheckedChange={onChange}
            />
            <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                {schema.title || "Enable"}
            </label>
        </div>
       );
  }

  // String / Number / Integer
  let inputType = "text";
  if (schema.type === "integer" || schema.type === "number") inputType = "number";
  if (schema.format === "password") inputType = "password";

  // Heuristic for password fields if format is missing
  if (schema.type === "string" && !schema.format && schema.title) {
      const lower = schema.title.toLowerCase();
      if (lower.includes("token") || lower.includes("secret") || lower.includes("key") || lower.includes("password")) {
          inputType = "password";
      }
  }

  return (
    <Input
        value={value || ""}
        onChange={(e) => {
            if (schema.type === "integer" || schema.type === "number") {
                onChange(e.target.valueAsNumber);
            } else {
                onChange(e.target.value);
            }
        }}
        placeholder={schema.default || ""}
        type={inputType}
    />
  );
}
