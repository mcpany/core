/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { ChevronDown, Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

/**
 * Props for the JsonTree component
 */
export interface JsonTreeProps {
  data: unknown;
  level?: number;
  defaultExpandedLevel?: number;
  className?: string;
}

/**
 * JsonTree component.
 * Renders a recursive tree view of JSON data with high contrast,
 * subtle borders, and meaningful animations (Unifi/Apple aesthetic).
 */
export function JsonTree({ data, level = 0, defaultExpandedLevel = 1, className }: JsonTreeProps) {
  const isObject = typeof data === 'object' && data !== null;
  const isArray = Array.isArray(data);
  const isEmpty = isObject && Object.keys(data as object).length === 0;

  const [expanded, setExpanded] = useState(level < defaultExpandedLevel);
  const [copied, setCopied] = useState(false);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    const text = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).catch(err => console.error("Clipboard error", err));
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!isObject) {
    return (
      <div className={cn("flex items-center gap-2 group/node font-mono text-sm leading-relaxed hover:bg-muted/50 rounded-md px-2 py-0.5 -ml-2 transition-colors", className)} style={{ paddingLeft: level > 0 ? '0' : undefined }}>
        <PrimitiveValue value={data} />
        <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 opacity-0 group-hover/node:opacity-100 transition-all duration-200 ml-auto bg-background/50 backdrop-blur border border-border/50 shadow-sm"
            onClick={handleCopy}
            title="Copy value"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
        </Button>
      </div>
    );
  }

  if (isEmpty) {
     return (
        <div className={cn("font-mono text-sm text-muted-foreground/50", className)}>
            {isArray ? "[]" : "{}"}
        </div>
     );
  }

  const entries = Object.entries(data as object);
  const keysCount = entries.length;

  return (
    <div className={cn("font-mono text-sm leading-relaxed", className)}>
      <div
        className="flex items-center gap-1.5 cursor-pointer hover:bg-muted/50 rounded-md px-2 py-1 -ml-2 select-none group/node transition-colors"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="text-muted-foreground/70 w-4 flex justify-center shrink-0 transition-transform duration-200" style={{ transform: expanded ? 'rotate(0deg)' : 'rotate(-90deg)' }}>
            <ChevronDown className="h-4 w-4" />
        </span>
        <span className="text-foreground/90 font-medium">{isArray ? "[" : "{"}</span>
        {!expanded && (
            <span className="text-muted-foreground/60 mx-1.5 text-xs px-1.5 py-0.5 bg-muted rounded border border-border/50 shadow-sm">
                {isArray ? `${keysCount} items` : `${keysCount} keys`}
            </span>
        )}
        {!expanded && (
             <span className="text-foreground/90 font-medium">{isArray ? "]" : "}"}</span>
        )}
         <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 opacity-0 group-hover/node:opacity-100 transition-all duration-200 ml-auto bg-background/50 backdrop-blur border border-border/50 shadow-sm"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
        </Button>
      </div>

      {expanded && (
        <div className="border-l-[1.5px] border-border/40 ml-[6px] pl-4 flex flex-col my-1 relative before:absolute before:inset-y-0 before:-left-[1.5px] before:w-[1.5px] before:bg-gradient-to-b before:from-transparent before:via-border/40 before:to-transparent">
          {entries.map(([key, value]) => (
            <div key={key} className="flex items-start gap-2 py-0.5">
               {/* Key */}
               <div className="pt-[2px] shrink-0 text-purple-600 dark:text-purple-400 font-medium">
                  {!isArray && (
                      <span className="mr-0.5">
                        "{key}":
                      </span>
                  )}
               </div>

               {/* Value */}
               <div className="flex-1 min-w-0">
                  <JsonTree
                    data={value}
                    level={level + 1}
                    defaultExpandedLevel={defaultExpandedLevel}
                  />
               </div>
            </div>
          ))}
        </div>
      )}
      {expanded && (
          <div className="pl-6 text-foreground/90 font-medium">
              {isArray ? "]" : "}"}
          </div>
      )}
    </div>
  );
}

function PrimitiveValue({ value }: { value: unknown }) {
  if (typeof value === 'string') {
    if (value.startsWith('data:image/') && value.length > 50) {
        return (
            <div className="mt-1 mb-2">
                <span className="text-green-600 dark:text-green-400 break-all whitespace-pre-wrap opacity-50 text-[10px] block truncate max-w-[300px]" title="Click copy to get full string">"{value}"</span>
                <img src={value} alt="Base64 Image" className="max-w-[200px] h-auto rounded-lg border border-border/50 shadow-sm bg-black/5 mt-2" />
            </div>
        );
    }
    return <span className="text-amber-600 dark:text-amber-400 break-all whitespace-pre-wrap font-medium">"{value}"</span>;
  }
  if (typeof value === 'number') {
    return <span className="text-blue-600 dark:text-blue-400 font-medium">{value}</span>;
  }
  if (typeof value === 'boolean') {
    return <span className="text-orange-600 dark:text-orange-400 font-medium">{value ? 'true' : 'false'}</span>;
  }
  if (value === null) {
    return <span className="text-muted-foreground/60 italic font-medium">null</span>;
  }
  if (value === undefined) {
    return <span className="text-muted-foreground/60 italic font-medium">undefined</span>;
  }
  return <span className="font-medium">{String(value)}</span>;
}
