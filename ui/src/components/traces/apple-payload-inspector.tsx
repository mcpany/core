/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Copy, ChevronRight, ChevronDown, Braces, Brackets, Hash, Type, ToggleLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";

/**
 * Props for the ApplePayloadInspector component.
 */
interface ApplePayloadInspectorProps {
  /**
   * The JSON payload to inspect.
   */
  payload: unknown;
}

/**
 * A highly polished, Apple/Unifi-inspired structured JSON inspector.
 * Replaces raw JSON dumps with a beautiful, expandable, type-aware UI.
 *
 * @param props - The component props.
 * @param props.payload - The JSON data.
 * @returns The rendered component.
 */
export function ApplePayloadInspector({ payload }: ApplePayloadInspectorProps) {
  const { toast } = useToast();

  const handleCopy = () => {
    navigator.clipboard.writeText(JSON.stringify(payload, null, 2));
    toast({
      title: "Copied",
      description: "Payload copied to clipboard."
    });
  };

  if (payload === undefined || payload === null) {
      return (
          <div className="flex items-center justify-center p-8 bg-muted/20 border border-dashed rounded-xl">
             <span className="text-muted-foreground font-mono text-sm opacity-50">null</span>
          </div>
      );
  }

  // If it's a primitive at the root level
  if (typeof payload !== 'object') {
      return (
          <div className="bg-card border rounded-xl shadow-sm overflow-hidden">
             <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/40 backdrop-blur-sm">
                 <div className="flex items-center gap-2">
                     <TypeIcon value={payload} />
                     <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Value</span>
                 </div>
                 <Button variant="ghost" size="icon" onClick={handleCopy} className="h-6 w-6">
                    <Copy className="h-3 w-3" />
                 </Button>
             </div>
             <div className="p-4 font-mono text-sm">
                 {String(payload)}
             </div>
          </div>
      );
  }

  return (
    <div className="bg-card border rounded-xl shadow-sm overflow-hidden text-sm">
      <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/40 backdrop-blur-sm">
        <div className="flex items-center gap-2">
          {Array.isArray(payload) ? <Brackets className="h-4 w-4 text-indigo-500" /> : <Braces className="h-4 w-4 text-emerald-500" />}
          <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
             {Array.isArray(payload) ? "Array" : "Object"}
          </span>
          <Badge variant="secondary" className="text-[10px] h-5 px-1.5 ml-1">
              {Array.isArray(payload) ? `${payload.length} items` : `${Object.keys(payload).length} keys`}
          </Badge>
        </div>
        <Button variant="ghost" size="icon" onClick={handleCopy} className="h-6 w-6" title="Copy Raw JSON">
          <Copy className="h-3 w-3" />
        </Button>
      </div>
      <div className="p-1">
        <div className="w-full">
            <JsonNode data={payload} name="root" isRoot={true} />
        </div>
      </div>
    </div>
  );
}

function TypeIcon({ value }: { value: unknown }) {
    if (value === null) return <span className="text-[10px] font-bold text-muted-foreground w-4 text-center">∅</span>;
    if (Array.isArray(value)) return <Brackets className="h-3.5 w-3.5 text-indigo-500" />;
    switch (typeof value) {
        case 'object': return <Braces className="h-3.5 w-3.5 text-emerald-500" />;
        case 'string': return <Type className="h-3.5 w-3.5 text-amber-500" />;
        case 'number': return <Hash className="h-3.5 w-3.5 text-blue-500" />;
        case 'boolean': return <ToggleLeft className="h-3.5 w-3.5 text-rose-500" />;
        default: return <Type className="h-3.5 w-3.5 text-muted-foreground" />;
    }
}

function JsonNode({ data, name, isRoot = false, depth = 0 }: { data: unknown, name: string, isRoot?: boolean, depth?: number }) {
    const [isExpanded, setIsExpanded] = useState(isRoot || depth < 2); // Auto-expand up to depth 2

    const isComplex = data !== null && typeof data === 'object';
    const isArray = Array.isArray(data);
    const keys = isComplex ? Object.keys(data) : [];

    if (!isComplex) {
        // Primitive rendering
        return (
             <div className="flex items-start py-1.5 px-2 hover:bg-muted/50 rounded-md transition-colors group">
                 <div className="w-5 shrink-0" /> {/* Spacer for alignment */}
                 <div className="flex items-center gap-2 w-1/3 min-w-[120px] shrink-0">
                     <TypeIcon value={data} />
                     <span className="font-medium text-muted-foreground">{name}</span>
                 </div>
                 <div className="flex-1 font-mono text-foreground break-all pl-4 border-l border-border/50">
                     {data === null ? (
                         <span className="text-muted-foreground/50 italic">null</span>
                     ) : typeof data === 'string' ? (
                         <span className="text-amber-600 dark:text-amber-400">"{data}"</span>
                     ) : typeof data === 'number' ? (
                         <span className="text-blue-600 dark:text-blue-400">{data}</span>
                     ) : typeof data === 'boolean' ? (
                         <span className="text-rose-600 dark:text-rose-400">{data ? 'true' : 'false'}</span>
                     ) : (
                         String(data)
                     )}
                 </div>
             </div>
        );
    }

    // Complex Object / Array rendering
    return (
        <div className="flex flex-col w-full">
            {!isRoot && (
                <div
                    className="flex items-center py-1.5 px-2 hover:bg-muted/50 rounded-md transition-colors cursor-pointer group select-none"
                    onClick={() => setIsExpanded(!isExpanded)}
                >
                    <div className="w-5 shrink-0 flex items-center justify-center text-muted-foreground hover:text-foreground">
                        {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                    </div>
                    <div className="flex items-center gap-2 w-1/3 min-w-[120px] shrink-0">
                        <TypeIcon value={data} />
                        <span className="font-medium text-foreground">{name}</span>
                    </div>
                    <div className="flex-1 pl-4 border-l border-border/50 flex items-center gap-2">
                        <span className="text-xs text-muted-foreground font-mono">
                            {isArray ? `Array(${keys.length})` : `Object{${keys.length}}`}
                        </span>
                        {!isExpanded && keys.length > 0 && (
                            <span className="text-xs text-muted-foreground/50 truncate max-w-[200px]">
                                {isArray ? '[...]' : '{...}'}
                            </span>
                        )}
                    </div>
                </div>
            )}

            {isExpanded && (
                <div className={cn("flex flex-col w-full", !isRoot && "pl-5 border-l border-border/30 ml-2.5 mt-1 mb-1")}>
                     {keys.length === 0 ? (
                         <div className="flex items-center py-1 px-2 pl-6 text-muted-foreground/50 text-xs italic">
                             Empty
                         </div>
                     ) : (
                         keys.map(key => (
                             <JsonNode
                                 key={key}
                                 name={key}
                                 data={data[key as keyof typeof data]}
                                 depth={depth + 1}
                             />
                         ))
                     )}
                </div>
            )}
        </div>
    );
}
