/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ToolDefinition } from "@/lib/client";
import { SchemaForm } from "./schema-form";

interface ToolFormProps {
  tool: ToolDefinition;
  onSubmit: (data: any) => void;
  onCancel: () => void;
}

export function ToolForm({ tool, onSubmit, onCancel }: ToolFormProps) {
  const [formData, setFormData] = useState<any>({});
  const [errors, setErrors] = useState<Record<string, string>>({});

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
        data.forEach((item, index) => {
            const itemPath = `${path}[${index}]`;
            const itemErrors = validate(schema.items, item, itemPath);
            newErrors = { ...newErrors, ...itemErrors };
        });
    }

    return newErrors;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const validationErrors = validate(tool.inputSchema, formData);

    if (Object.keys(validationErrors).length > 0) {
        setErrors(validationErrors);
        return;
    }

    setErrors({});
    onSubmit(formData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 py-2 flex flex-col h-[60vh]">
      <div className="flex-1 overflow-y-auto pr-2">
         {(!tool.inputSchema || !tool.inputSchema.properties || Object.keys(tool.inputSchema.properties).length === 0) ? (
             <div className="text-sm text-muted-foreground italic">
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
      </div>

      <div className="flex justify-end gap-2 pt-4 border-t mt-auto">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit">
          Run Tool
        </Button>
      </div>
    </form>
  );
}
