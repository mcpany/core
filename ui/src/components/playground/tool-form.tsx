/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ToolDefinition } from "@/lib/client";
import { SchemaForm } from "./schema-form";
import { ToolPresets } from "./tool-presets";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { JsonView } from "@/components/ui/json-view";
import Ajv, { ErrorObject } from "ajv";
import addFormats from "ajv-formats";
import { Check, Copy, Code } from "lucide-react";
import { generateCurlCommand, generatePythonCommand } from "@/lib/code-gen";
import dynamic from "next/dynamic";
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

const SyntaxHighlighter = dynamic(
    () => import("react-syntax-highlighter").then((mod) => mod.Prism),
    {
        ssr: false,
        loading: () => <div className="p-4 bg-muted h-32 animate-pulse rounded" />,
    }
);

interface ToolFormProps {
  tool: ToolDefinition;
  onSubmit: (data: Record<string, unknown>) => void;
  onCancel: () => void;
}

/**
 * ToolForm.
 *
 * @param onCancel - The onCancel.
 */
export function ToolForm({ tool, onSubmit, onCancel }: ToolFormProps) {
  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [jsonInput, setJsonInput] = useState<string>("{}");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [mode, setMode] = useState<"form" | "json" | "schema" | "code">("form");
  const [copiedCurl, setCopiedCurl] = useState(false);
  const [copiedPython, setCopiedPython] = useState(false);

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
      // Common logic: update state based on current mode before switching
      if (mode === "form") {
          setJsonInput(JSON.stringify(formData, null, 2));
      } else if (mode === "json") {
          try {
              const parsed = JSON.parse(jsonInput);
              setFormData(parsed);
              // Validate immediately on switch
              setErrors(runValidation(parsed));
          } catch (e) {
              setErrors({ "json": "Cannot switch: Invalid JSON." });
              return; // Do NOT switch mode
          }
      }

      setMode(value as any);
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

  const handleCopy = (text: string, type: 'curl' | 'python') => {
      navigator.clipboard.writeText(text);
      if (type === 'curl') {
          setCopiedCurl(true);
          setTimeout(() => setCopiedCurl(false), 2000);
      } else {
          setCopiedPython(true);
          setTimeout(() => setCopiedPython(false), 2000);
      }
  };

  // Real-time validation for JSON mode
  useEffect(() => {
      if (mode === "json") {
          try {
              const parsed = JSON.parse(jsonInput);
              setErrors(runValidation(parsed));
          } catch (e) {
               // Ignore parse errors during typing
          }
      } else if (mode === "form") {
          if (Object.keys(errors).length > 0) {
             setErrors(runValidation(formData));
          }
      }
  }, [formData, jsonInput, mode]);

  return (
    <form onSubmit={handleSubmit} className="space-y-4 py-2 flex flex-col h-[60vh]">
      <Tabs value={mode} onValueChange={handleTabChange} className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-1 mb-2">
            <TabsList className="grid w-[400px] grid-cols-4">
                <TabsTrigger value="form">Form</TabsTrigger>
                <TabsTrigger value="json">JSON</TabsTrigger>
                <TabsTrigger value="schema">Schema</TabsTrigger>
                <TabsTrigger value="code" className="flex gap-1 items-center">
                    <Code className="w-3 h-3" /> Code
                </TabsTrigger>
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

        <TabsContent value="code" className="flex-1 overflow-hidden mt-0">
             <div className="h-full flex flex-col gap-4 overflow-y-auto pr-2">
                 {/* Curl Section */}
                 <div className="space-y-2">
                     <div className="flex items-center justify-between">
                         <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">cURL</h4>
                         <Button variant="ghost" size="sm" type="button" className="h-6 text-xs gap-1" onClick={() => handleCopy(generateCurlCommand(tool.name, getCurrentData()), 'curl')}>
                             {copiedCurl ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                             Copy
                         </Button>
                     </div>
                     <div className="rounded-md overflow-hidden border">
                         <SyntaxHighlighter
                            language="bash"
                            style={vscDarkPlus}
                            customStyle={{ margin: 0, padding: '1rem', fontSize: '12px' }}
                            wrapLines={true}
                            wrapLongLines={true}
                         >
                             {generateCurlCommand(tool.name, getCurrentData())}
                         </SyntaxHighlighter>
                     </div>
                 </div>

                 {/* Python Section */}
                 <div className="space-y-2">
                     <div className="flex items-center justify-between">
                         <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Python</h4>
                         <Button variant="ghost" size="sm" type="button" className="h-6 text-xs gap-1" onClick={() => handleCopy(generatePythonCommand(tool.name, getCurrentData()), 'python')}>
                             {copiedPython ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                             Copy
                         </Button>
                     </div>
                     <div className="rounded-md overflow-hidden border">
                         <SyntaxHighlighter
                            language="python"
                            style={vscDarkPlus}
                            customStyle={{ margin: 0, padding: '1rem', fontSize: '12px' }}
                            wrapLines={true}
                            wrapLongLines={true}
                         >
                             {generatePythonCommand(tool.name, getCurrentData())}
                         </SyntaxHighlighter>
                     </div>
                 </div>
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
