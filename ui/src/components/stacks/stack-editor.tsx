/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, Save, X, Eye, Code, LayoutTemplate } from "lucide-react";
import yaml from "js-yaml";
import { ConfigEditor } from "./config-editor";
import { StackVisualizer } from "./stack-visualizer";
import { ServicePalette } from "./service-palette";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/components/ui/tabs";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";

interface StackEditorProps {
  initialContent?: string;
  onSave: (content: string) => Promise<void>;
  onCancel: () => void;
}

/**
 * StackEditor component for editing stack configurations (YAML).
 * @param props The component props.
 * @returns The rendered component.
 */
export function StackEditor({ initialContent = "", onSave, onCancel }: StackEditorProps) {
  const [content, setContent] = useState(initialContent);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  // Default to "editor" which now maps to the Split View to satisfy E2E tests expecting visualizer visibility in Editor tab
  const [activeTab, setActiveTab] = useState("editor");

  const handleSave = async () => {
    setError(null);
    setSaving(true);
    try {
        // Basic YAML validation
        try {
            yaml.load(content);
        } catch (e: any) {
            throw new Error(`Invalid YAML: ${e.message}`);
        }

        await onSave(content);
    } catch (e: any) {
        setError(e.message || "Failed to save stack");
    } finally {
        setSaving(false);
    }
  };

  const handleTemplateSelect = (snippet: string) => {
      // Simple append for now
      setContent(prev => {
          // Ensure we append on a new line with separation
          const prefix = prev.endsWith("\n") ? "\n" : "\n\n";
          return prev + prefix + snippet;
      });
  };

  return (
    <div className="flex flex-col h-full space-y-4">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex-1 min-h-[600px] border rounded-md overflow-hidden bg-background">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
              <div className="flex items-center justify-between px-4 border-b bg-muted/20">
                  <TabsList className="bg-transparent p-0 gap-2">
                      <TabsTrigger value="editor" className="data-[state=active]:bg-background data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-10 px-4">
                          <LayoutTemplate className="mr-2 h-4 w-4" /> Editor
                      </TabsTrigger>
                      <TabsTrigger value="code" className="data-[state=active]:bg-background data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-10 px-4">
                          <Code className="mr-2 h-4 w-4" /> Code
                      </TabsTrigger>
                      <TabsTrigger value="visualizer" className="data-[state=active]:bg-background data-[state=active]:shadow-none border-b-2 border-transparent data-[state=active]:border-primary rounded-none h-10 px-4">
                          <Eye className="mr-2 h-4 w-4" /> Visualizer
                      </TabsTrigger>
                  </TabsList>
                  <div className="text-xs text-muted-foreground">
                      config.yaml
                  </div>
              </div>

              <div className="flex-1 overflow-hidden relative">
                  {/* Editor Tab now uses Split View to satisfy tests */}
                  <TabsContent value="editor" className="h-full m-0 data-[state=inactive]:hidden">
                      <ResizablePanelGroup direction="horizontal">
                          <ResizablePanel defaultSize={20} minSize={15} maxSize={30}>
                               <ServicePalette onTemplateSelect={handleTemplateSelect} />
                          </ResizablePanel>
                          <ResizableHandle />
                          <ResizablePanel defaultSize={40}>
                              <ConfigEditor
                                  value={content}
                                  onChange={(value) => setContent(value || "")}
                                  language="yaml"
                              />
                          </ResizablePanel>
                          <ResizableHandle />
                          <ResizablePanel defaultSize={40}>
                              <StackVisualizer yamlContent={content} />
                          </ResizablePanel>
                      </ResizablePanelGroup>
                  </TabsContent>

                  <TabsContent value="code" className="h-full m-0 data-[state=inactive]:hidden">
                      <div className="flex h-full">
                          <ServicePalette onTemplateSelect={handleTemplateSelect} />
                          <div className="flex-1 h-full">
                              <ConfigEditor
                                  value={content}
                                  onChange={(value) => setContent(value || "")}
                                  language="yaml"
                              />
                          </div>
                      </div>
                  </TabsContent>

                  <TabsContent value="visualizer" className="h-full m-0 data-[state=inactive]:hidden">
                        <StackVisualizer yamlContent={content} />
                  </TabsContent>
              </div>
          </Tabs>
      </div>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={saving}>
          <X className="mr-2 h-4 w-4" /> Cancel
        </Button>
        <Button onClick={handleSave} disabled={saving}>
          <Save className="mr-2 h-4 w-4" /> {saving ? "Saving..." : "Deploy Stack"}
        </Button>
      </div>
    </div>
  );
}
