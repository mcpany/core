/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Eye, Edit, Variable } from "lucide-react";

interface PromptEditorProps {
  value: string;
  onChange: (value: string) => void;
}

/**
 * PromptEditor component.
 * Provides a dual-pane editor for authoring prompts with variable interpolation preview.
 *
 * @param props - The component props.
 * @param props.value - The current prompt text (markdown).
 * @param props.onChange - Callback when the prompt text changes.
 * @returns The rendered component.
 */
export function PromptEditor({ value, onChange }: PromptEditorProps) {
  const [activeTab, setActiveTab] = React.useState("edit");
  const [variables, setVariables] = React.useState<string[]>([]);
  const [testValues, setTestValues] = React.useState<Record<string, string>>({});

  // Parse variables on value change
  React.useEffect(() => {
    const matches = value.match(/\{\{([^}]+)\}\}/g);
    if (matches) {
      const vars = Array.from(new Set(matches.map((m) => m.slice(2, -2).trim())));
      setVariables(vars);
    } else {
      setVariables([]);
    }
  }, [value]);

  const previewText = React.useMemo(() => {
    let text = value;
    variables.forEach((v) => {
      const val = testValues[v] || `{{${v}}}`; // Keep placeholder if empty? Or use bold placeholder?
      // Simple string replace (global)
      // Escaping special regex chars in variable name not strictly handled here but generally variable names are simple
      text = text.split(`{{${v}}}`).join(val === `{{${v}}}` ? `**${val}**` : val);
    });
    return text;
  }, [value, variables, testValues]);

  return (
    <div className="flex flex-col h-full min-h-[400px] border rounded-md overflow-hidden bg-background">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex items-center justify-between px-4 border-b bg-muted/40">
          <TabsList className="bg-transparent p-0 h-10">
            <TabsTrigger
              value="edit"
              className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-4"
            >
              <Edit className="w-4 h-4 mr-2" /> Edit
            </TabsTrigger>
            <TabsTrigger
              value="preview"
              className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-full px-4"
            >
              <Eye className="w-4 h-4 mr-2" /> Preview
            </TabsTrigger>
          </TabsList>
          {variables.length > 0 && (
            <div className="text-xs text-muted-foreground flex items-center gap-1">
              <Variable className="w-3 h-3" />
              {variables.length} variables detected
            </div>
          )}
        </div>

        <TabsContent value="edit" className="flex-1 p-0 m-0 relative group">
          <Textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="w-full h-full resize-none border-0 focus-visible:ring-0 p-4 font-mono text-sm leading-relaxed bg-transparent z-10 relative"
            placeholder="Enter instructions here. Use {{variable}} for dynamic inputs."
          />
          {/* Simple overlay hints could go here if we had absolute positioning and text measuring, but for now just the textarea is enough */}
        </TabsContent>

        <TabsContent value="preview" className="flex-1 flex flex-col md:flex-row p-0 m-0">
          {/* Variable Inputs Sidebar */}
          {variables.length > 0 && (
            <div className="w-full md:w-64 border-b md:border-b-0 md:border-r bg-muted/10 p-4 overflow-y-auto">
              <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-4">
                Test Variables
              </h4>
              <div className="space-y-4">
                {variables.map((v) => (
                  <div key={v} className="space-y-1.5">
                    <Label htmlFor={`var-${v}`} className="text-xs font-mono">
                      {v}
                    </Label>
                    <Input
                      id={`var-${v}`}
                      value={testValues[v] || ""}
                      onChange={(e) =>
                        setTestValues({ ...testValues, [v]: e.target.value })
                      }
                      placeholder={`Value for ${v}`}
                      className="h-8 text-xs"
                    />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Rendered Preview */}
          <div className="flex-1 p-6 overflow-y-auto bg-white/50 dark:bg-black/20">
            {value ? (
              <div className="prose dark:prose-invert max-w-none text-sm">
                {/*
                    Using simple whitespace-pre-wrap for now to preserve structure.
                    Ideally use ReactMarkdown but need to ensure it handles the substituted bolding correctly.
                 */}
                 <p className="whitespace-pre-wrap font-sans">{previewText}</p>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-muted-foreground opacity-50">
                <Edit className="w-12 h-12 mb-2" />
                <p>Start editing to see preview</p>
              </div>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
