/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React, { useState } from "react";
import { ChevronRight, ChevronDown, Copy, Check, FileJson } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface JsonTreeProps {
  data: unknown;
  level?: number;
  defaultExpandedLevel?: number;
  className?: string;
  name?: string;
}

/**
 * Intent: Document JsonTree
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * JsonTree component.
 * Renders a recursive, polished "Apple Design Standard" tree view of JSON data.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.level - The current nesting level (default: 0).
 * @param props.defaultExpandedLevel - The level up to which nodes are expanded by default (default: 1).
 * @param props.className - The className.
 * @param props.name - Optional property name.
 * @returns The rendered component.
 */
export function JsonTree({ data, level = 0, defaultExpandedLevel = 1, className, name }: JsonTreeProps) {
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

  const wrapperClass = level === 0
    ? cn("font-mono text-xs rounded-md border bg-muted/10 backdrop-blur-sm p-2 w-full", className)
    : cn("font-mono text-xs w-full", className);

  if (!isObject) {
    return (
      <div className={cn("flex items-center gap-4 group/node hover:bg-accent/50 rounded-sm px-2 py-1 -mx-2 transition-colors w-full", wrapperClass)}>
        {name && (
            <span className="font-semibold opacity-90 min-w-[120px] text-foreground shrink-0">{name}</span>
        )}
        <div className="flex-1 flex items-center gap-3 min-w-0">
            <PrimitiveValue value={data} />
            <PrimitiveBadge value={data} />
        </div>
        <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 opacity-0 group-hover/node:opacity-100 transition-opacity ml-auto shrink-0"
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
        <div className={cn("flex items-center gap-4 group/node hover:bg-accent/50 rounded-sm px-2 py-1 -mx-2 transition-colors w-full text-muted-foreground", wrapperClass)}>
            {name && (
                <span className="font-semibold opacity-90 min-w-[120px] text-foreground shrink-0">{name}</span>
            )}
            <span className="flex-1">{isArray ? "[]" : "{}"}</span>
        </div>
     );
  }

  const entries = Object.entries(data as object);
  const preview = isArray
    ? `Array(${entries.length})`
    : `{ ${entries.slice(0, 3).map(([k]) => k).join(", ")}${entries.length > 3 ? ", ..." : ""} }`;

  return (
    <div className={wrapperClass}>
      <div
        className="flex items-center gap-2 cursor-pointer hover:bg-accent/50 rounded-sm px-2 py-1 -mx-2 transition-colors select-none group/node w-full"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="text-muted-foreground w-4 flex justify-center shrink-0">
            {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
        </span>
        {name && (
             <span className="font-semibold opacity-90 min-w-[100px] text-foreground shrink-0">{name}</span>
        )}

        <div className="flex-1 flex items-center gap-2 text-muted-foreground truncate min-w-0">
            {level === 0 && !name && <FileJson className="h-3 w-3" />}
            {isArray ? "[" : "{"}
            {!expanded && (
                <span className="opacity-50 italic text-[10px] truncate">{preview}</span>
            )}
            {!expanded && (
                 <span>{isArray ? "]" : "}"}</span>
            )}
        </div>

        {expanded && (
            <Badge variant="outline" className="h-4 px-1 text-[9px] font-normal tracking-wide text-muted-foreground opacity-50 ml-auto shrink-0 uppercase">
                {isArray ? `Array[${entries.length}]` : 'Object'}
            </Badge>
        )}

         <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 opacity-0 group-hover/node:opacity-100 transition-opacity ml-2 shrink-0"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
        </Button>
      </div>

      {expanded && (
        <div className="border-l-2 border-border/50 ml-[9px] pl-4 mt-1 flex flex-col gap-0.5 relative">
          {/* Subtle line glow effect */}
          <div className="absolute left-[-2px] top-0 bottom-0 w-[2px] bg-gradient-to-b from-transparent via-primary/20 to-transparent opacity-0 hover:opacity-100 transition-opacity pointer-events-none" />

          {entries.map(([key, value]) => (
             <div key={key} className="flex flex-col w-full">
                <JsonTree
                  data={value}
                  level={level + 1}
                  defaultExpandedLevel={defaultExpandedLevel}
                  name={isArray ? undefined : key}
                />
             </div>
          ))}
        </div>
      )}
      {expanded && (
          <div className="pl-6 text-muted-foreground mt-1 py-1">
              {isArray ? "]" : "}"}
          </div>
      )}
    </div>
  );
}

/**
 * PrimitiveValue component.
 * @param props - The component props.
 * @param props.value - The current value.
 * @returns The rendered component.
 */
function PrimitiveValue({ value }: { value: unknown }) {
  if (typeof value === 'string') {
    if (value.startsWith('data:image/') && value.length > 50) {
        return (
            <div className="mt-1 mb-2 max-w-full">
                <span className="text-green-500 dark:text-green-400 break-all whitespace-pre-wrap opacity-50 text-[10px] block truncate max-w-[300px]" title="Click copy to get full string">"{value}"</span>
                <img src={value} alt="Base64 Image" className="max-w-[200px] h-auto rounded-md border bg-black/50 mt-1 shadow-sm" />
            </div>
        );
    }
    return <span className="text-green-600 dark:text-green-400 break-all whitespace-pre-wrap">"{value}"</span>;
  }
  if (typeof value === 'number') {
    return <span className="text-blue-600 dark:text-blue-400 font-medium">{value}</span>;
  }
  if (typeof value === 'boolean') {
    return <span className="text-orange-600 dark:text-orange-400 font-medium">{value ? 'true' : 'false'}</span>;
  }
  if (value === null) {
    return <span className="text-muted-foreground italic opacity-70">null</span>;
  }
  if (value === undefined) {
    return <span className="text-muted-foreground italic opacity-70">undefined</span>;
  }
  return <span className="text-foreground">{String(value)}</span>;
}

/**
 * PrimitiveBadge component returns a distinct type badge.
 * @param props - The component props.
 * @param props.value - The current value.
 * @returns The rendered component.
 */
function PrimitiveBadge({ value }: { value: unknown }) {
    let type = typeof value;
    if (value === null) type = 'null';

    return (
        <Badge variant="outline" className="h-[18px] px-1.5 text-[9px] font-normal uppercase tracking-wider text-muted-foreground opacity-50 border-muted-foreground/20 bg-muted/5">
            {type}
        </Badge>
    );
}
