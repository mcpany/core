/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useMemo, useEffect } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Eye, Edit3, Variable } from "lucide-react";

interface PromptEditorProps {
  value: string;
  onChange: (value: string) => void;
}

/**
 * PromptEditor component.
 * provides an editor for prompts with variable detection and preview.
 *
 * @param props - The component props.
 * @param props.value - The current prompt text.
 * @param props.onChange - Callback when text changes.
 * @returns The rendered component.
 */
export function PromptEditor({ value, onChange }: PromptEditorProps) {
  const [activeTab, setActiveTab] = useState("edit");
  const [previewValues, setPreviewValues] = useState<Record<string, string>>({});

  // Detect variables like {{var_name}}
  const variables = useMemo(() => {
    const regex = /\{\{([^}]+)\}\}/g;
    const vars = new Set<string>();
    let match;
    while ((match = regex.exec(value)) !== null) {
      vars.add(match[1].trim());
    }
    return Array.from(vars);
  }, [value]);

  const renderedPreview = useMemo(() => {
    return value.replace(/\{\{([^}]+)\}\}/g, (match, p1) => {
        const varName = p1.trim();
        return previewValues[varName] !== undefined ? previewValues[varName] : match;
    });
  }, [value, previewValues]);

  return (
    <div className="flex flex-col h-full border rounded-md overflow-hidden bg-background">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex items-center justify-between px-4 border-b bg-muted/20">
            <TabsList className="bg-transparent border-b-0 h-10 p-0">
                <TabsTrigger value="edit" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4">
                    <Edit3 className="mr-2 h-4 w-4" /> Edit
                </TabsTrigger>
                <TabsTrigger value="preview" className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4">
                    <Eye className="mr-2 h-4 w-4" /> Preview
                </TabsTrigger>
            </TabsList>
            {variables.length > 0 && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground mr-2">
                    <Variable className="h-3 w-3" />
                    <span>{variables.length} variable{variables.length !== 1 ? 's' : ''} detected</span>
                </div>
            )}
        </div>

        <TabsContent value="edit" className="flex-1 mt-0 relative">
          <Textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="h-full w-full resize-none rounded-none border-0 focus-visible:ring-0 p-4 font-mono text-sm leading-relaxed"
            placeholder="Enter instructions here. Use {{variable}} for dynamic inputs."
          />
        </TabsContent>

        <TabsContent value="preview" className="flex-1 mt-0 overflow-y-auto p-4 bg-muted/5">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 h-full">
                {/* Inputs Column */}
                <div className="md:col-span-1 space-y-4 border-r pr-4">
                    <h4 className="text-sm font-medium mb-2">Test Variables</h4>
                    {variables.length === 0 ? (
                        <div className="text-sm text-muted-foreground italic">
                            No variables detected. Add <code>{"{{name}}"}</code> to your instructions.
                        </div>
                    ) : (
                        variables.map((v) => (
                            <div key={v} className="space-y-1.5">
                                <Label htmlFor={`var-${v}`} className="text-xs font-mono">{v}</Label>
                                <Input
                                    id={`var-${v}`}
                                    placeholder={`Value for ${v}`}
                                    value={previewValues[v] || ""}
                                    onChange={(e) => setPreviewValues({...previewValues, [v]: e.target.value})}
                                    className="h-8 text-sm"
                                />
                            </div>
                        ))
                    )}
                </div>

                {/* Preview Column */}
                <div className="md:col-span-2 flex flex-col">
                    <h4 className="text-sm font-medium mb-2">Rendered Output</h4>
                    <Card className="flex-1 bg-card">
                        <CardContent className="p-4 h-full">
                            <div className="prose dark:prose-invert max-w-none text-sm whitespace-pre-wrap font-mono">
                                {renderedPreview}
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
