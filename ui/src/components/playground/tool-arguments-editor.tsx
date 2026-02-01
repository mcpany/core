/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { ToolDefinition } from "@/lib/client";
import { SchemaForm } from "./schema-form";
import { ToolPresets } from "./tool-presets";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { JsonView } from "@/components/ui/json-view";
import Ajv, { ErrorObject } from "ajv";
import addFormats from "ajv-formats";

export interface ToolArgumentsEditorProps {
  tool: ToolDefinition;
  value: Record<string, unknown>;
  onChange: (data: Record<string, unknown>, isValid: boolean) => void;
  className?: string;
}

/**
 * ToolArgumentsEditor.
 * A controlled component for editing tool arguments via Form or JSON.
 *
 * @param props - The component props.
 */
export function ToolArgumentsEditor({ tool, value, onChange, className }: ToolArgumentsEditorProps) {
  // Local state to handle draft inputs (especially invalid JSON or partial form updates)
  // We initialize from props.value, but we don't strictly sync back from props.value on every render
  // unless it deeply changes (to avoid cursor jumps or overwriting draft).
  // However, for simplicity in "controlled" mode, we usually rely on parent passing back the value.
  // But here, we have "invalid" intermediate states (jsonInput).

  // Strategy:
  // 1. We keep `formData` (object) and `jsonInput` (string).
  // 2. We track which mode we are in.
  // 3. When `props.value` changes externally (and is different from our `formData`), we update both.
  // 4. When user types, we update local state and call `onChange`.

  const [formData, setFormData] = useState<Record<string, unknown>>(value || {});
  const [jsonInput, setJsonInput] = useState<string>(JSON.stringify(value || {}, null, 2));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [mode, setMode] = useState<"form" | "json" | "schema">("form");

  // Sync from props when tool or value changes significantly
  // We use JSON.stringify to compare content to avoid object reference issues
  useEffect(() => {
      const currentJson = JSON.stringify(formData);
      const propJson = JSON.stringify(value || {});
      if (currentJson !== propJson) {
          setFormData(value || {});
          setJsonInput(JSON.stringify(value || {}, null, 2));
          // We don't change mode
      }
  }, [value, tool.name]); // Reset if tool changes

  // Initialize AJV and compile schema
  const validate = useMemo(() => {
    if (!tool.inputSchema || Object.keys(tool.inputSchema).length === 0) {
        return null;
    }
    const ajv = new Ajv({ allErrors: true, strict: false });
    addFormats(ajv);
    try {
        return ajv.compile(tool.inputSchema);
    } catch (e) {
        console.error("Failed to compile schema:", e);
        return null;
    }
  }, [tool.inputSchema]);

  // Broadcast initial validity on mount or tool change
  useEffect(() => {
      // Defer to next tick to ensure validate is ready?
      // Actually validate is a dependency.
      if (validate) {
          // Check validity silently (no setErrors)
          const valid = validate(formData);
          // If invalid, we don't map errors yet to avoid visual noise
          onChange(formData, !!valid);
      } else {
          // No schema or empty schema -> valid
          onChange(formData, true);
      }
  }, [validate, tool.name]);

  const mapErrors = (ajvErrors: ErrorObject[] | null | undefined): Record<string, string> => {
      if (!ajvErrors) return {};
      const newErrors: Record<string, string> = {};

      ajvErrors.forEach(err => {
          let path = err.instancePath;
          if (path.startsWith('/')) path = path.slice(1);
          path = path.replace(/\//g, '.');
          path = path.replace(/\.(\d+)(?=\.|$)/g, '[$1]');
          path = path.replace(/^(\d+)(?=\.|$)/g, '[$1]');

          if (err.keyword === 'required') {
              const missingProperty = err.params.missingProperty;
              path = path ? `${path}.${missingProperty}` : missingProperty;
              newErrors[path] = "This field is required";
          } else {
              const key = path || "json";
              newErrors[key] = err.message || "Invalid value";
          }
      });
      return newErrors;
  };

  const runValidation = (data: unknown) => {
      if (!validate) return {};
      const valid = validate(data);
      if (!valid) {
          return mapErrors(validate.errors);
      }
      return {};
  };

  const handleTabChange = (value: string) => {
      if (value === "schema") {
          setMode("schema");
          if (mode === "form") {
              setJsonInput(JSON.stringify(formData, null, 2));
          }
          return;
      }

      if (value === "json") {
          if (mode === "form") {
              setJsonInput(JSON.stringify(formData, null, 2));
          }
          setMode("json");
          // Validate immediately
          const validationErrors = runValidation(formData);
          setErrors(validationErrors);
      } else if (value === "form") {
          try {
              const parsed = JSON.parse(jsonInput);
              setFormData(parsed);
              setMode("form");
              const validationErrors = runValidation(parsed);
              setErrors(validationErrors);
              // Update parent with valid data
              onChange(parsed, Object.keys(validationErrors).length === 0);
          } catch (e) {
              setErrors({ "json": "Cannot switch to Form view: Invalid JSON." });
          }
      }
  };

  const handlePresetSelect = (data: Record<string, unknown>) => {
      setFormData(data);
      setJsonInput(JSON.stringify(data, null, 2));
      const validationErrors = runValidation(data);
      setErrors(validationErrors);
      onChange(data, Object.keys(validationErrors).length === 0);
  };

  // Handle Form Change
  const handleFormChange = (newVal: Record<string, unknown>) => {
      setFormData(newVal);
      // Validate on change (optional: debounce?)
      // We validate immediately for responsiveness
      const validationErrors = runValidation(newVal);
      setErrors(validationErrors);
      onChange(newVal, Object.keys(validationErrors).length === 0);
  };

  // Handle JSON Change
  const handleJsonChange = (val: string) => {
      setJsonInput(val);
      try {
          const parsed = JSON.parse(val);
          // Update local formData to avoid re-sync from props triggering reformat
          setFormData(parsed);

          // If valid JSON, check schema
          const validationErrors = runValidation(parsed);
          setErrors(validationErrors);
          // Sync to formData implicitly? No, only on mode switch or forced sync.
          // But parent expects object.
          // If JSON is valid, we should update parent.
          onChange(parsed, Object.keys(validationErrors).length === 0);

          // Also clear JSON syntax error if present
          if (errors.json === "Invalid JSON format") {
              const newErrors = {...errors};
              delete newErrors.json;
              setErrors(newErrors);
          }
      } catch (e) {
          // Syntax error
           setErrors({ "json": "Invalid JSON format" });
           onChange({}, false); // Invalid
      }
  };

  const getCurrentData = () => {
      if (mode === "json") {
          try {
              return JSON.parse(jsonInput);
          } catch {
              return formData;
          }
      }
      return formData;
  };

  return (
    <div className={className}>
      <Tabs value={mode} onValueChange={handleTabChange} className="flex flex-col h-full overflow-hidden">
        <div className="flex items-center justify-between px-1 mb-2 shrink-0">
            <TabsList className="grid w-[300px] grid-cols-3">
                <TabsTrigger value="form">Form</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
                <TabsTrigger value="schema">Schema</TabsTrigger>
            </TabsList>
            <ToolPresets
                toolName={tool.name}
                currentData={getCurrentData()}
                onSelect={handlePresetSelect}
            />
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
                    onChange={handleFormChange}
                    errors={errors}
                 />
             )}
        </TabsContent>

        <TabsContent value="json" className="flex-1 overflow-hidden mt-0">
            <div className="h-full flex flex-col gap-2">
                <Textarea
                    value={jsonInput}
                    onChange={(e) => handleJsonChange(e.target.value)}
                    className="font-mono text-xs flex-1 resize-none"
                    placeholder="{ ... }"
                />
                {errors.json && (
                    <p className="text-xs text-destructive">{errors.json}</p>
                )}
                {mode === "json" && Object.keys(errors).length > 0 && !errors.json && (
                    <div className="text-xs text-destructive max-h-[100px] overflow-y-auto border border-destructive/20 bg-destructive/10 p-2 rounded">
                        <p className="font-semibold mb-1">Validation Errors:</p>
                        <ul className="list-disc pl-4 space-y-1">
                            {Object.entries(errors).map(([path, msg]) => (
                                <li key={path}>
                                    <span className="font-mono">{path || "root"}</span>: {msg}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </div>
        </TabsContent>

        <TabsContent value="schema" className="flex-1 overflow-hidden mt-0">
            <div className="h-full flex flex-col gap-2">
                <JsonView data={tool.inputSchema} className="h-full overflow-auto" maxHeight={0} />
            </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
