/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { AlertTriangle } from "lucide-react";

interface SchemaFormProps {
  schema: any;
  value: Record<string, string>;
  onChange: (value: Record<string, string>) => void;
}

/**
 * SchemaForm component.
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.value - The current value.
 * @param props.onChange - Callback function when value changes.
 * @returns The rendered component.
 */
export function SchemaForm({ schema, value, onChange }: SchemaFormProps) {
  if (!schema) return null;

  if (!schema.properties || typeof schema.properties !== 'object') {
      return (
          <div className="flex items-center gap-2 p-4 text-sm text-yellow-500 bg-yellow-500/10 rounded-lg border border-yellow-500/20">
              <AlertTriangle className="h-4 w-4" />
              <span>Invalid configuration schema: Missing properties definition.</span>
          </div>
      );
  }

  const handleChange = (key: string, val: string) => {
    onChange({ ...value, [key]: val });
  };

  return (
    <div className="space-y-4 border rounded-lg p-4 bg-muted/20">
      {Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
        const isRequired = schema.required?.includes(key);
        const description = prop.description || "";
        const title = prop.title || key;

        // Determine input type
        let inputType = "text";
        if (prop.format === "password") {
            inputType = "password";
        } else if (
            key.toLowerCase().includes("token") ||
            key.toLowerCase().includes("secret") ||
            key.toLowerCase().includes("key") ||
            key.toLowerCase().includes("password")
        ) {
            inputType = "password";
        }

        return (
          <div key={key} className="grid gap-2">
            <Label htmlFor={key} className="flex items-center gap-1">
              {title} {isRequired && <span className="text-destructive">*</span>}
            </Label>

            {prop.type === "boolean" ? (
               <div className="flex items-center space-x-2">
                <Checkbox
                    id={key}
                    checked={value[key] === "true"}
                    onCheckedChange={(c) => handleChange(key, String(c))}
                />
                <label
                    htmlFor={key}
                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                    {description || "Enable"}
                </label>
               </div>
            ) : prop.enum ? (
                <Select value={value[key]} onValueChange={(v) => handleChange(key, v)}>
                    <SelectTrigger>
                        <SelectValue placeholder="Select..." />
                    </SelectTrigger>
                    <SelectContent>
                        {prop.enum.map((opt: string) => (
                            <SelectItem key={opt} value={opt}>{opt}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            ) : (
                <Input
                    id={key}
                    value={value[key] || ""}
                    onChange={(e) => handleChange(key, e.target.value)}
                    placeholder={prop.default || `Enter ${title}`}
                    type={inputType}
                />
            )}

            {prop.type !== "boolean" && description && (
                <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
        );
      })}
    </div>
  );
}
