/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { ChevronRight, ChevronDown, Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface JsonTreeProps {
  data: unknown;
  level?: number;
  defaultExpandedLevel?: number;
  className?: string;
}

/**
 * JsonTree component.
 * Renders a recursive tree view of JSON data.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.level - The current nesting level (default: 0).
 * @param props.defaultExpandedLevel - The level up to which nodes are expanded by default (default: 1).
 * @param props.className - The className.
 * @returns The rendered component.
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
      <div className={cn("flex items-center gap-2 group/node font-mono text-xs hover:bg-white/5 rounded px-1 -ml-1", className)} style={{ paddingLeft: level > 0 ? '0' : undefined }}>
        <PrimitiveValue value={data} />
        <Button
            variant="ghost"
            size="icon"
            className="h-4 w-4 opacity-0 group-hover/node:opacity-100 transition-opacity ml-auto"
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
        <div className={cn("font-mono text-xs text-muted-foreground", className)}>
            {isArray ? "[]" : "{}"}
        </div>
     );
  }

  const entries = Object.entries(data as object);
  const preview = isArray
    ? `Array(${entries.length})`
    : `{ ${entries.slice(0, 3).map(([k]) => k).join(", ")}${entries.length > 3 ? ", ..." : ""} }`;

  return (
    <div className={cn("font-mono text-xs", className)}>
      <div
        className="flex items-center gap-1 cursor-pointer hover:bg-white/5 rounded px-1 -ml-1 select-none group/node"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="text-muted-foreground w-4 flex justify-center shrink-0">
            {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
        </span>
        <span className="text-muted-foreground">{isArray ? "[" : "{"}</span>
        {!expanded && (
            <span className="text-muted-foreground opacity-50 mx-1 italic text-[10px]">{preview}</span>
        )}
        {!expanded && (
             <span className="text-muted-foreground">{isArray ? "]" : "}"}</span>
        )}
         <Button
            variant="ghost"
            size="icon"
            className="h-4 w-4 opacity-0 group-hover/node:opacity-100 transition-opacity ml-auto"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
        </Button>
      </div>

      {expanded && (
        <div className="border-l border-white/10 ml-2 pl-2 flex flex-col">
          {entries.map(([key, value], idx) => (
            <div key={key} className="flex items-start gap-1">
               {/* Key */}
               <div className="pt-[2px] shrink-0 text-purple-400">
                  {!isArray && (
                      <span className="mr-1 opacity-80">
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
               {/* Comma if needed (optional purely visual preference, syntax highlighter usually omits in tree view but keeps structure) */}
            </div>
          ))}
        </div>
      )}
      {expanded && (
          <div className="pl-6 text-muted-foreground">
              {isArray ? "]" : "}"}
          </div>
      )}
    </div>
  );
}

function PrimitiveValue({ value }: { value: unknown }) {
  if (typeof value === 'string') {
    return <span className="text-green-400 break-all whitespace-pre-wrap">"{value}"</span>;
  }
  if (typeof value === 'number') {
    return <span className="text-blue-400">{value}</span>;
  }
  if (typeof value === 'boolean') {
    return <span className="text-orange-400">{value ? 'true' : 'false'}</span>;
  }
  if (value === null) {
    return <span className="text-gray-500 italic">null</span>;
  }
  if (value === undefined) {
    return <span className="text-gray-500 italic">undefined</span>;
  }
  return <span>{String(value)}</span>;
}
