/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { ChevronRight, ChevronDown, Code, Box, Braces } from "lucide-react";
import { cn } from "@/lib/utils";
import { JsonView } from "@/components/ui/json-view";

interface SmartPayloadViewerProps {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    data: any;
    title?: string;
    icon?: React.ReactNode;
}

/**
 * SmartPayloadViewer displays JSON payload in an Apple-style, high-contrast, structured UI.
 *
 * @param props - The component props.
 * @param props.data - The data payload to display.
 * @param props.title - Optional title.
 * @param props.icon - Optional icon.
 * @returns The rendered component.
 */
export function SmartPayloadViewer({ data, title = "Payload", icon = <Code className="h-4 w-4" /> }: SmartPayloadViewerProps) {
    const [isExpanded, setIsExpanded] = useState(true);

    if (data === undefined || data === null) {
        return (
            <div className="w-full rounded-xl border border-muted/30 bg-background/50 backdrop-blur-md overflow-hidden shadow-sm">
                <div className="flex items-center gap-2 p-3 bg-muted/10 border-b border-muted/20">
                    <span className="text-muted-foreground">{icon}</span>
                    <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
                </div>
                <div className="p-4 text-xs italic text-muted-foreground">null</div>
            </div>
        );
    }

    // Heuristics for JSON-RPC MCP structures
    const isRpcRequest = data.jsonrpc === "2.0" && data.method !== undefined;
    const isRpcResponse = data.jsonrpc === "2.0" && (data.result !== undefined || data.error !== undefined);
    const isMcpArgs = data.arguments !== undefined && typeof data.arguments === 'object';

    // Top-level meta extracting
    const extractMeta = () => {
        if (isRpcRequest) return { Method: data.method, ID: data.id };
        if (isRpcResponse && data.error) return { ErrorCode: data.error.code, ErrorMessage: data.error.message, ID: data.id };
        if (isRpcResponse && data.result) return { ID: data.id, Success: true };
        return null;
    };

    const meta = extractMeta();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const renderKeyValue = (key: string, value: any, depth = 0) => {
        const isObj = typeof value === 'object' && value !== null;
        return (
            <div key={key} className={cn("flex flex-col py-1.5", depth > 0 && "pl-4 border-l border-muted/20")}>
                 <div className="flex items-start gap-2">
                     <span className="text-xs font-medium text-primary/80 shrink-0 min-w-[100px] font-mono">{key}</span>
                     {!isObj ? (
                          <span className={cn(
                              "text-xs break-all",
                              typeof value === 'string' ? "text-green-600 dark:text-green-400" :
                              typeof value === 'number' ? "text-blue-600 dark:text-blue-400" :
                              typeof value === 'boolean' ? "text-purple-600 dark:text-purple-400 font-semibold" :
                              "text-muted-foreground"
                          )}>
                              {typeof value === 'string' ? `"${value}"` : String(value)}
                          </span>
                     ) : null}
                 </div>
                 {isObj && (
                     <div className="mt-1">
                         {Array.isArray(value) ? (
                             value.length > 0 ? (
                                 value.map((item, idx) => renderKeyValue(`[${idx}]`, item, depth + 1))
                             ) : <span className="text-xs text-muted-foreground ml-4">[]</span>
                         ) : (
                             Object.keys(value).length > 0 ? (
                                 Object.entries(value).map(([k, v]) => renderKeyValue(k, v, depth + 1))
                             ) : <span className="text-xs text-muted-foreground ml-4">{}</span>
                         )}
                     </div>
                 )}
            </div>
        );
    };

    return (
        <div className="w-full rounded-xl border border-muted/30 bg-background/50 backdrop-blur-md overflow-hidden shadow-sm transition-all duration-200">
             <button
                className="w-full flex items-center justify-between p-3 bg-muted/10 hover:bg-muted/20 border-b border-muted/20 transition-colors cursor-pointer"
                onClick={() => setIsExpanded(!isExpanded)}
             >
                  <div className="flex items-center gap-2">
                      <div className="p-1 rounded bg-background/80 shadow-sm text-primary">
                          {icon}
                      </div>
                      <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
                  </div>
                  <div className="flex items-center gap-3">
                      {meta && Object.entries(meta).map(([k, v]) => (
                           v !== undefined && (
                               <div key={k} className="hidden sm:flex items-center gap-1.5 text-[10px] bg-background/50 px-2 py-0.5 rounded-full border border-muted/30 shadow-sm">
                                   <span className="text-muted-foreground uppercase font-bold tracking-wider">{k}</span>
                                   <span className={cn("font-mono", String(v).includes('Error') || k === 'Error' ? "text-red-500 font-semibold" : "")}>{String(v)}</span>
                               </div>
                           )
                      ))}
                      {isExpanded ? <ChevronDown className="h-4 w-4 text-muted-foreground" /> : <ChevronRight className="h-4 w-4 text-muted-foreground" />}
                  </div>
             </button>

             {isExpanded && (
                  <div className="flex flex-col animate-in fade-in slide-in-from-top-2 duration-200">
                      {isRpcRequest && isMcpArgs && (
                           <div className="p-4 border-b border-muted/20 bg-muted/5">
                                <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-3 flex items-center gap-1.5">
                                    <Box className="h-3 w-3" /> Arguments
                                </h4>
                                <div className="space-y-1 bg-background/50 rounded-lg p-3 border border-muted/20 shadow-sm">
                                    {Object.entries(data.arguments).map(([k, v]) => renderKeyValue(k, v))}
                                    {Object.keys(data.arguments).length === 0 && (
                                         <span className="text-xs text-muted-foreground italic">No arguments provided.</span>
                                    )}
                                </div>
                           </div>
                      )}

                      {(isRpcResponse && data.error) && (
                           <div className="p-4 border-b border-red-200 dark:border-red-900/30 bg-red-50 dark:bg-red-900/10">
                                <h4 className="text-xs font-bold uppercase tracking-wider text-red-600 dark:text-red-400 mb-3 flex items-center gap-1.5">
                                    <Braces className="h-3 w-3" /> Error Details
                                </h4>
                                <div className="space-y-1 bg-background/80 rounded-lg p-3 border border-red-100 dark:border-red-900/50 shadow-sm text-red-900 dark:text-red-200">
                                    {Object.entries(data.error).map(([k, v]) => renderKeyValue(k, v))}
                                </div>
                           </div>
                      )}

                      <div className="p-4">
                           <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-3 flex items-center gap-1.5">
                                <Code className="h-3 w-3" /> Raw JSON
                           </h4>
                           <div className="rounded-lg border border-muted/30 overflow-hidden shadow-sm">
                               <JsonView data={data} maxHeight={400} />
                           </div>
                      </div>
                  </div>
             )}
        </div>
    );
}
