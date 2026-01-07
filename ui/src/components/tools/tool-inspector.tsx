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

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  const [input, setInput] = useState("{}");
  const [output, setOutput] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  if (!tool) return null;

  const handleExecute = async () => {
    setLoading(true);
    setOutput(null);
    try {
      const args = JSON.parse(input);
      const res = await apiClient.executeTool({
          toolName: tool.name,
          arguments: args
      });
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
                <ScrollArea className="h-[150px] w-full rounded-md border p-4 bg-muted/50">
                    <pre className="text-xs">{JSON.stringify(tool.inputSchema, null, 2)}</pre>
                </ScrollArea>
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

        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>Close</Button>
          <Button onClick={handleExecute} disabled={loading}>
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlayCircle className="mr-2 h-4 w-4" />}
            Execute
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
