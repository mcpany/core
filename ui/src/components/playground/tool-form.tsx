/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ToolDefinition } from "@/lib/client";
import { SchemaForm } from "./schema-form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";

interface ToolFormProps {
  tool: ToolDefinition;
  onSubmit: (data: Record<string, unknown>) => void;
  onCancel: () => void;
}

export function ToolForm({ tool, onSubmit, onCancel }: ToolFormProps) {
  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [jsonInput, setJsonInput] = useState<string>("{}");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [mode, setMode] = useState<"form" | "json">("form");

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const validate = (schema: any, data: any, path: string = ""): Record<string, string> => {
    let newErrors: Record<string, string> = {};

    if (!schema) return newErrors;

    const required = schema.required || [];
    const properties = schema.properties || {};

    // Check required fields at this level
    for (const field of required) {
        const fieldPath = path ? `${path}.${field}` : field;
        const value = data?.[field];
        if (value === undefined || value === "" || value === null) {
            newErrors[fieldPath] = "This field is required";
        }
    }

    // Recurse into object properties
    if (schema.type === "object" && properties) {
        for (const key of Object.keys(properties)) {
             const fieldPath = path ? `${path}.${key}` : key;
             const fieldSchema = properties[key];

             // Validate if field is present (handles optional nested objects correctly)
             if (data?.[key] !== undefined) {
                 const childErrors = validate(fieldSchema, data[key], fieldPath);
                 newErrors = { ...newErrors, ...childErrors };
             }
        }
    }

    // Recurse into array items
    if (schema.type === "array" && schema.items && Array.isArray(data)) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        data.forEach((item: any, index: number) => {
            const itemPath = `${path}[${index}]`;
            const itemErrors = validate(schema.items, item, itemPath);
            newErrors = { ...newErrors, ...itemErrors };
        });
    }

    return newErrors;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    let finalData = formData;

    if (mode === "json") {
        try {
            finalData = JSON.parse(jsonInput);
        } catch (err) {
            setErrors({ "json": "Invalid JSON format" });
            return;
        }
    }

    const validationErrors = validate(tool.inputSchema, finalData);

    if (Object.keys(validationErrors).length > 0) {
        setErrors(validationErrors);
        return;
    }

    setErrors({});
    onSubmit(finalData);
  };

  const handleTabChange = (value: string) => {
      if (value === "json") {
          setJsonInput(JSON.stringify(formData, null, 2));
          setMode("json");
      } else {
          try {
              const parsed = JSON.parse(jsonInput);
              setFormData(parsed);
              setMode("form");
              setErrors({});
          } catch (e) {
              setErrors({ "json": "Cannot switch to Form view: Invalid JSON." });
              // Do NOT switch mode
          }
      }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 py-2 flex flex-col h-[60vh]">
      <Tabs value={mode} onValueChange={handleTabChange} className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-1 mb-2">
            <TabsList className="grid w-[200px] grid-cols-2">
                <TabsTrigger value="form">Form</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
            </TabsList>
        </div>

        <TabsContent value="form" className="flex-1 overflow-y-auto pr-2 mt-0">
             {(!tool.inputSchema || !tool.inputSchema.properties || Object.keys(tool.inputSchema.properties).length === 0) ? (
                 <div className="text-sm text-muted-foreground italic p-1">
                     This tool takes no arguments.
                 </div>
             ) : (
                 <SchemaForm
                    schema={tool.inputSchema}
                    value={formData}
                    onChange={(val) => {
                        setFormData(val);
                    }}
                    errors={errors}
                 />
             )}
        </TabsContent>

        <TabsContent value="json" className="flex-1 overflow-hidden mt-0">
            <div className="h-full flex flex-col gap-2">
                <Textarea
                    value={jsonInput}
                    onChange={(e) => {
                        setJsonInput(e.target.value);
                        setErrors({});
                    }}
                    className="font-mono text-xs flex-1 resize-none"
                    placeholder="{ ... }"
                />
                {errors.json && (
                    <p className="text-xs text-destructive">{errors.json}</p>
                )}
            </div>
        </TabsContent>
      </Tabs>

      <div className="flex justify-end gap-2 pt-4 border-t mt-auto">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit">
          Build Command
        </Button>
      </div>
    </form>
  );
}
