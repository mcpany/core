/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Bug, Play, X, AlertTriangle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface InterceptorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  toolName: string;
  payload: Record<string, unknown>;
  onExecute: (modifiedPayload: Record<string, unknown>) => void;
  onCancel: () => void;
}

/**
 * InterceptorDialog component.
 * Allows inspecting and modifying the tool execution payload before sending.
 */
export function InterceptorDialog({
  open,
  onOpenChange,
  toolName,
  payload,
  onExecute,
  onCancel
}: InterceptorDialogProps) {
  const [jsonContent, setJsonContent] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setJsonContent(JSON.stringify(payload, null, 2));
      setError(null);
    }
  }, [open, payload]);

  const handleExecute = () => {
    try {
      const modified = JSON.parse(jsonContent);
      onExecute(modified);
      onOpenChange(false);
    } catch (e: any) {
      setError(e.message || "Invalid JSON");
    }
  };

  const handleCancel = () => {
    onCancel();
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] border-amber-500/50">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="bg-amber-500/10 text-amber-500 border-amber-500/20 gap-1">
                <Bug className="h-3 w-3" /> Intercepted
            </Badge>
            <DialogTitle>Breakpoint Hit: {toolName}</DialogTitle>
          </div>
          <DialogDescription>
            Execution paused. You can modify the arguments before sending to the server.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-2">
            <Textarea
              value={jsonContent}
              onChange={(e) => {
                  setJsonContent(e.target.value);
                  setError(null);
              }}
              className="font-mono text-xs h-[300px] bg-muted/30"
              spellCheck={false}
            />
          </div>

          {error && (
            <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>JSON Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={handleCancel}>
            <X className="mr-2 h-4 w-4" /> Cancel
          </Button>
          <Button onClick={handleExecute} className="bg-amber-600 hover:bg-amber-700 text-white">
            <Play className="mr-2 h-4 w-4" /> Resume Execution
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
