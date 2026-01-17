/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { ToolDefinition, apiClient } from "@/lib/client";
import { ScrollArea } from "@/components/ui/scroll-area";
import { PlayCircle, Loader2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SchemaViewer } from "./schema-viewer";

import { Switch } from "@/components/ui/switch";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * ToolInspector.
 *
 * @param onOpenChange - The onOpenChange.
 */
export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  const [input, setInput] = useState("{}");
  const [output, setOutput] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [isDryRun, setIsDryRun] = useState(false);

  if (!tool) return null;

  const handleExecute = async () => {
    setLoading(true);
    setOutput(null);
    try {
      const args = JSON.parse(input);
      const res = await apiClient.executeTool({
          toolName: tool.name,
          arguments: args
      }, isDryRun);
      setOutput(JSON.stringify(res, null, 2));
    } catch (e: any) {
      setOutput(`Error: ${e.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
              {tool.name}
              <Badge variant="outline">{tool.serviceId}</Badge>
          </DialogTitle>
          <DialogDescription>
            {tool.description}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
            <div className="grid gap-2">
                <Label>Schema</Label>
                <Tabs defaultValue="visual" className="w-full">
                  <TabsList className="grid w-[200px] grid-cols-2 h-8">
                    <TabsTrigger value="visual" className="text-xs">Visual</TabsTrigger>
                    <TabsTrigger value="json" className="text-xs">JSON</TabsTrigger>
                  </TabsList>
                  <TabsContent value="visual" className="mt-2">
                     <ScrollArea className="h-[200px] w-full rounded-md border p-4 bg-muted/20">
                        <SchemaViewer schema={tool.inputSchema as any} />
                     </ScrollArea>
                  </TabsContent>
                  <TabsContent value="json" className="mt-2">
                    <ScrollArea className="h-[200px] w-full rounded-md border p-4 bg-muted/50">
                        <pre className="text-xs">{JSON.stringify(tool.inputSchema, null, 2)}</pre>
                    </ScrollArea>
                  </TabsContent>
                </Tabs>
            </div>

            <div className="grid gap-2">
                <Label htmlFor="args">Arguments (JSON)</Label>
                <Textarea
                    id="args"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    className="font-mono text-sm"
                    rows={5}
                />
            </div>

            {output && (
                 <div className="grid gap-2">
                    <Label>Result</Label>
                    <ScrollArea className="h-[150px] w-full rounded-md border p-4 bg-muted/50">
                        <pre className="text-xs text-green-600 dark:text-green-400 font-mono">{output}</pre>
                    </ScrollArea>
                </div>
            )}
        </div>

        <DialogFooter className="flex justify-between items-center sm:justify-between">
          <div className="flex items-center space-x-2">
              <Switch id="dry-run" checked={isDryRun} onCheckedChange={setIsDryRun} />
              <Label htmlFor="dry-run">Dry Run</Label>
          </div>
          <div className="flex gap-2">
              <Button variant="secondary" onClick={() => onOpenChange(false)}>Close</Button>
              <Button onClick={handleExecute} disabled={loading}>
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
                Execute
              </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
