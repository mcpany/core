/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



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
      <div className={cn("flex items-center group/node font-mono text-xs w-full", className)} style={{ paddingLeft: level > 0 ? '0' : undefined }}>
        <div className="flex-1 min-w-0">
            <PrimitiveValue value={data} />
        </div>
        <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 opacity-0 group-hover/node:opacity-100 transition-opacity shrink-0 ml-2 bg-muted/50 hover:bg-muted"
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
    <div className={cn("font-mono text-xs w-full", className)}>
      <div
        className="flex items-center gap-1 cursor-pointer hover:bg-muted/30 rounded py-1 px-1 -ml-1 select-none group/node w-full"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="text-muted-foreground w-4 flex justify-center shrink-0">
            {expanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
        </span>
        <span className="text-muted-foreground">{isArray ? "[" : "{"}</span>
        {!expanded && (
            <span className="text-muted-foreground opacity-50 mx-1 italic text-[10px] truncate max-w-[200px] md:max-w-md">{preview}</span>
        )}
        {!expanded && (
             <span className="text-muted-foreground">{isArray ? "]" : "}"}</span>
        )}
         <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5 opacity-0 group-hover/node:opacity-100 transition-opacity ml-auto shrink-0 bg-muted/50 hover:bg-muted"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
        </Button>
      </div>

      {expanded && (
        <div className="border-l-2 border-muted-foreground/20 ml-[7px] pl-3 flex flex-col gap-1 my-1 w-full">
          {entries.map(([key, value]) => (
            <div key={key} className="flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-4 w-full group/row hover:bg-muted/10 rounded px-1 -ml-1 py-0.5">
               {/* Key */}
               <div className="pt-[2px] shrink-0 text-primary/80 sm:w-[30%] sm:max-w-[250px] font-semibold truncate" title={key}>
                  {!isArray && (
                      <span>{key}</span>
                  )}
                  {isArray && (
                      <span className="text-muted-foreground/50 text-[10px]">{key}</span>
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
          <div className="pl-2 text-muted-foreground/60 font-semibold">
              {isArray ? "]" : "}"}
          </div>
      )}
    </div>
  );
}

function PrimitiveValue({ value }: { value: unknown }) {
  const [expandedText, setExpandedText] = useState(false);

  if (typeof value === 'string') {
    if (value.startsWith('data:image/') && value.length > 50) {
        return (
            <div className="mt-1 mb-2">
                <span className="text-green-600 dark:text-green-400 break-all whitespace-pre-wrap opacity-50 text-[10px] block truncate max-w-[300px]" title="Click copy to get full string">"{value}"</span>
                <img src={value} alt="Base64 Image" className="max-w-[200px] h-auto rounded-md border bg-black/50 mt-1" />
            </div>
        );
    }

    const isLong = value.length > 100;

    return (
        <div className="flex flex-col items-start w-full">
            <span
                className={cn(
                    "text-green-700 dark:text-green-400 break-all whitespace-pre-wrap",
                    !expandedText && isLong && "line-clamp-3"
                )}
            >
                "{value}"
            </span>
            {isLong && (
                <button
                    onClick={(e) => { e.stopPropagation(); setExpandedText(!expandedText); }}
                    className="text-[10px] text-muted-foreground hover:text-primary mt-1 font-sans underline underline-offset-2"
                >
                    {expandedText ? "Show less" : "Show more"}
                </button>
            )}
        </div>
    );
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
