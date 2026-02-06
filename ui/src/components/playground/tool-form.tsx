/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ToolDefinition, PromptDefinition } from "@/lib/client";
import { SchemaForm } from "./schema-form";
import { ToolPresets } from "./tool-presets";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { JsonView } from "@/components/ui/json-view";
import Ajv, { ErrorObject } from "ajv";
import addFormats from "ajv-formats";

interface ToolFormProps {
  definition: ToolDefinition | PromptDefinition;
  onSubmit: (data: Record<string, unknown>) => void;
  onCancel: () => void;
}

/**
 * ToolForm.
 *
 * @param onCancel - The onCancel.
 */
export function ToolForm({ definition, onSubmit, onCancel }: ToolFormProps) {
  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [jsonInput, setJsonInput] = useState<string>("{}");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [mode, setMode] = useState<"form" | "json" | "schema">("form");

  // Initialize AJV and compile schema
  const validate = useMemo(() => {
    if (!definition.inputSchema || Object.keys(definition.inputSchema).length === 0) {
        return null;
    }
    const ajv = new Ajv({ allErrors: true, strict: false });
    addFormats(ajv);
    try {
        return ajv.compile(definition.inputSchema);
    } catch (e) {
        console.error("Failed to compile schema:", e);
        return null;
    }
  }, [definition.inputSchema]);

  const mapErrors = (ajvErrors: ErrorObject[] | null | undefined): Record<string, string> => {
      if (!ajvErrors) return {};
      const newErrors: Record<string, string> = {};

      ajvErrors.forEach(err => {
          let path = err.instancePath;

          // Convert /foo/bar/0/baz to foo.bar[0].baz
          // Remove leading slash
          if (path.startsWith('/')) path = path.slice(1);

          // Replace slashes with dots
          path = path.replace(/\//g, '.');

          // Replace .number with [number] for arrays
          // We need to be careful not to replace object keys that happen to be numbers,
          // but strictly speaking in JS dot notation is fine for number keys too,
          // however SchemaForm uses [x] for arrays.
          // AJV instancePath for arrays is like /arr/0.
          path = path.replace(/\.(\d+)(?=\.|$)/g, '[$1]');

          // Also handle the case where it starts with a number (root array)
          path = path.replace(/^(\d+)(?=\.|$)/g, '[$1]');

          if (err.keyword === 'required') {
              const missingProperty = err.params.missingProperty;
              path = path ? `${path}.${missingProperty}` : missingProperty;
              newErrors[path] = "This field is required";
          } else {
              // Use a fallback key if path is empty (root error)
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

    const validationErrors = runValidation(finalData);

    if (Object.keys(validationErrors).length > 0) {
        setErrors(validationErrors);
        return;
    }

    setErrors({});
    onSubmit(finalData);
  };

  const handleTabChange = (value: string) => {
      if (value === "schema") {
          setMode("schema");
          // Update JSON input from form data if coming from form, so returning to JSON works
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
          // Validate immediately on switch
          setErrors(runValidation(formData));
      } else if (value === "form") {
          try {
              const parsed = JSON.parse(jsonInput);
              setFormData(parsed);
              setMode("form");
              setErrors(runValidation(parsed));
          } catch (e) {
              setErrors({ "json": "Cannot switch to Form view: Invalid JSON." });
              // Do NOT switch mode
          }
      }
  };

  const handlePresetSelect = (data: Record<string, unknown>) => {
      setFormData(data);
      setJsonInput(JSON.stringify(data, null, 2));
      // Validation will trigger via useEffect
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

  // Real-time validation for JSON mode
  useEffect(() => {
      if (mode === "json") {
          try {
              const parsed = JSON.parse(jsonInput);
              setErrors(runValidation(parsed));
          } catch (e) {
               // Don't show parse error while typing, only on submit or switch.
               // Or maybe show it? Textarea handles change events.
               // Let's just clear schema errors but keep "json" error if user manually set it via submit?
               // Actually, if it's invalid JSON, we can't validate schema.
          }
      } else {
          // Real-time validation for Form mode
          // We could debounce this if needed, but for small forms it's fine.
          // Check if we want to show errors immediately or only after first submit attempt?
          // Typically immediate is annoying if fields are empty.
          // But existing code passed errors prop to SchemaForm.
          // Let's only validate if there are already errors (meaning user tried to submit or switched tabs with invalid data)
          // OR if we want "Premium" feel, maybe validate only fields that are dirty?
          // For now, to keep it simple and consistent with "Playground Schema Validation" goal:
          // We will validate on change IF we already have errors.
          if (Object.keys(errors).length > 0) {
             setErrors(runValidation(formData));
          }
      }
  }, [formData, jsonInput, mode]);
  // removed errors from deps to avoid infinite loop if setErrors is called.
  // actually including formData/jsonInput is enough.

  return (
    <form onSubmit={handleSubmit} className="space-y-4 py-2 flex flex-col h-[60vh]">
      <Tabs value={mode} onValueChange={handleTabChange} className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-1 mb-2">
            <TabsList className="grid w-[300px] grid-cols-3">
                <TabsTrigger value="form">Form</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
                <TabsTrigger value="schema">Schema</TabsTrigger>
            </TabsList>
            <ToolPresets
                toolName={definition.name}
                currentData={getCurrentData()}
                onSelect={handlePresetSelect}
            />
        </div>

        <TabsContent value="form" className="flex-1 overflow-y-auto pr-2 mt-0">
             {(!definition.inputSchema || !definition.inputSchema.properties || Object.keys(definition.inputSchema.properties).length === 0) ? (
                 <div className="text-sm text-muted-foreground italic p-1">
                     This tool takes no arguments.
                 </div>
             ) : (
                 <SchemaForm
                    schema={definition.inputSchema}
                    value={formData}
                    onChange={(val) => {
                        setFormData(val as Record<string, unknown>);
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
                        // Optional: Clear syntax error on type
                        if (errors.json === "Invalid JSON format") {
                            const newErrors = {...errors};
                            delete newErrors.json;
                            setErrors(newErrors);
                        }
                    }}
                    className="font-mono text-xs flex-1 resize-none"
                    placeholder="{ ... }"
                />
                {errors.json && (
                    <p className="text-xs text-destructive">{errors.json}</p>
                )}
                {/* Show other schema errors in JSON mode too, maybe as a list? */}
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
                <JsonView data={definition.inputSchema} className="h-full overflow-auto" maxHeight={0} />
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
