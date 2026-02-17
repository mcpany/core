/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { ConfigEditor } from "./config-editor";
import { StackVisualizer } from "./stack-visualizer";
import { ServicePalette } from "./service-palette";
import { Button } from "@/components/ui/button";
import { Loader2, Save, X, PanelLeftClose, PanelLeftOpen, Columns, Maximize2 } from "lucide-react";
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from "@/components/ui/resizable";
import * as yaml from "js-yaml";

interface StackEditorProps {
  initialValue: string;
  onSave: (value: string) => Promise<void>;
  onCancel: () => void;
}

/**
 * StackEditor component for editing stack configurations with visual feedback.
 *
*
 * Summary: StackEditor component for editing stack configurations with visual feedback.
 *
 * @param props - Component properties
 * @param props.initialValue - Initial YAML content
 * @param props.onSave - Callback when saving
 * @param props.onCancel - Callback when cancelling
 * @returns The rendered StackEditor
 */
export function StackEditor({ initialValue, onSave, onCancel }: StackEditorProps) {
  const [value, setValue] = useState(initialValue);
  const [saving, setSaving] = useState(false);
  const [showPalette, setShowPalette] = useState(true);
  const [showVisualizer, setShowVisualizer] = useState(true);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(value);
    } finally {
      setSaving(false);
    }
  };

  const handleTemplateSelect = (snippet: string) => {
      try {
          const doc = yaml.load(value) as any || {};
          const template = yaml.load(snippet) as any;

          // Template is usually an array of one item: [{name: ...}]
          const newService = Array.isArray(template) ? template[0] : template;

          if (!doc.services) {
              doc.services = [];
          }

          // Check if services is array or map
          if (Array.isArray(doc.services)) {
              doc.services.push(newService);
          } else if (typeof doc.services === 'object') {
              // It's a map, we need to convert newService to map entry
              // newService is {name: "foo", ...}
              const name = newService.name || `service-${Object.keys(doc.services).length + 1}`;
              const { name: _, ...rest } = newService;
              doc.services[name] = rest;
          }

          const newYaml = yaml.dump(doc, { indent: 2, lineWidth: -1 });
          setValue(newYaml);
      } catch (e) {
          console.error("Failed to smartly insert template, falling back to append", e);
          const newValue = value + "\n" + snippet;
          setValue(newValue);
      }
  };

  return (
    <div className="flex flex-col h-full bg-background">
      <div className="flex items-center justify-between p-2 border-b bg-muted/40 shrink-0">
        <div className="flex items-center gap-2">
             <Button variant="ghost" size="icon" onClick={() => setShowPalette(!showPalette)} title="Toggle Palette">
                {showPalette ? <PanelLeftClose className="h-4 w-4" /> : <PanelLeftOpen className="h-4 w-4" />}
            </Button>
            <span className="text-sm font-medium text-muted-foreground ml-2">Stack Composer</span>
        </div>
        <div className="flex items-center gap-2">
           <Button variant="ghost" size="icon" onClick={() => setShowVisualizer(!showVisualizer)} title="Toggle Visualizer">
                {showVisualizer ? <Maximize2 className="h-4 w-4" /> : <Columns className="h-4 w-4" />}
            </Button>
          <div className="w-px h-4 bg-border mx-2" />
          <Button variant="ghost" size="sm" onClick={onCancel}>
            <X className="mr-2 h-4 w-4" /> Cancel
          </Button>
          <Button size="sm" onClick={handleSave} disabled={saving}>
            {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
            Save & Deploy
          </Button>
        </div>
      </div>

      <div className="flex-1 min-h-0">
          <ResizablePanelGroup direction="horizontal">
              {showPalette && (
                  <>
                    <ResizablePanel defaultSize={20} minSize={15} maxSize={30}>
                        <ServicePalette onTemplateSelect={handleTemplateSelect} />
                    </ResizablePanel>
                    <ResizableHandle />
                  </>
              )}

              <ResizablePanel defaultSize={showVisualizer ? 50 : 80}>
                  <ConfigEditor value={value} onChange={(v) => setValue(v || "")} />
              </ResizablePanel>

              {showVisualizer && (
                  <>
                    <ResizableHandle />
                    <ResizablePanel defaultSize={30} minSize={20}>
                        <StackVisualizer yamlContent={value} />
                    </ResizablePanel>
                  </>
              )}
          </ResizablePanelGroup>
      </div>
    </div>
  );
}
