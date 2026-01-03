/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ToolDefinition } from "@/lib/client";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import Link from "next/link";

interface ToolInspectorProps {
  tool: ToolDefinition | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ToolInspector({ tool, open, onOpenChange }: ToolInspectorProps) {
  if (!tool) return null;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[500px] sm:w-[600px]">
        <SheetHeader>
          <SheetTitle className="flex items-center space-x-2">
            <span>{tool.name}</span>
            <Badge variant={tool.enabled ? "default" : "secondary"}>
              {tool.enabled ? "Enabled" : "Disabled"}
            </Badge>
          </SheetTitle>
          <SheetDescription>{tool.description}</SheetDescription>
          <div className="mt-2">
              <Link href={`/playground?tool=${encodeURIComponent(tool.name)}`}>
                  <Button size="sm" variant="secondary">
                      Run in Playground
                  </Button>
              </Link>
          </div>
        </SheetHeader>

        <div className="mt-6 space-y-6">
          <div>
            <h4 className="text-sm font-medium mb-2">Schema</h4>
            <div className="rounded-md border bg-muted/50 p-4">
              <ScrollArea className="h-[300px]">
                <pre className="text-xs font-mono">
                  {JSON.stringify(tool.schema, null, 2)}
                </pre>
              </ScrollArea>
            </div>
          </div>

          <Separator />

          <div className="grid grid-cols-2 gap-4">
             <div>
                <h4 className="text-sm font-medium text-muted-foreground">Service</h4>
                <p className="text-sm">{tool.serviceName}</p>
             </div>
             <div>
                <h4 className="text-sm font-medium text-muted-foreground">Source</h4>
                <p className="text-sm">{tool.source || "Unknown"}</p>
             </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
