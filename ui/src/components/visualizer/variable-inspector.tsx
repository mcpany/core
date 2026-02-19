/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { JsonView } from "@/components/ui/json-view";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Node } from "@xyflow/react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface VariableInspectorProps {
  selectedNode: Node | null;
  onClose: () => void;
}

/**
 * VariableInspector displays details and state of the selected node.
 * @param props - The component props.
 * @param props.selectedNode - The currently selected node.
 * @param props.onClose - Callback to close the inspector.
 * @returns The VariableInspector component.
 */
export function VariableInspector({ selectedNode, onClose }: VariableInspectorProps) {
  if (!selectedNode) return null;

  return (
    <Card className="w-80 border-l shadow-xl h-full rounded-none rounded-l-lg absolute right-0 top-0 bottom-0 z-20 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">
          Inspector: {selectedNode.data.label as string}
        </CardTitle>
        <Button variant="ghost" size="icon" onClick={onClose} className="h-8 w-8">
            <X className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent className="overflow-y-auto h-[calc(100%-4rem)]">
        <div className="space-y-4">
            <div>
                <h4 className="text-xs font-semibold mb-2 text-muted-foreground">Node Properties</h4>
                <div className="text-xs space-y-1">
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">ID:</span>
                        <span className="font-mono">{selectedNode.id}</span>
                    </div>
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Type:</span>
                        <span className="font-mono">{selectedNode.type}</span>
                    </div>
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Position:</span>
                        <span className="font-mono">
                            {Math.round(selectedNode.position.x)}, {Math.round(selectedNode.position.y)}
                        </span>
                    </div>
                </div>
            </div>

            <div>
                <h4 className="text-xs font-semibold mb-2 text-muted-foreground">Data (Variables)</h4>
                <JsonView data={selectedNode.data} />
            </div>
        </div>
      </CardContent>
    </Card>
  );
}
