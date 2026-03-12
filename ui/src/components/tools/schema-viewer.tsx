/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { ChevronRight, ChevronDown, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

/**
 * Schema represents a JSON Schema object used for defining tool input parameters.
 */
export interface Schema {
  type?: string | string[];
  description?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  anyOf?: Schema[];
  oneOf?: Schema[];
  allOf?: Schema[];
  enum?: any[];
  default?: any;
  format?: string;
  [key: string]: any;
}

interface SchemaViewerProps {
  schema: Schema;
  name?: string;
  required?: boolean;
  depth?: number;
  isLast?: boolean;
}

const getTypeColor = (type?: string | string[]) => {
  const t = Array.isArray(type) ? type[0] : type;
  switch (t) {
    case "string": return "text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800";
    case "number":
    case "integer": return "text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800";
    case "boolean": return "text-purple-600 dark:text-purple-400 bg-purple-50 dark:bg-purple-900/20 border-purple-200 dark:border-purple-800";
    case "object": return "text-slate-600 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/20 border-slate-200 dark:border-slate-800";
    case "array": return "text-orange-600 dark:text-orange-400 bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800";
    case "null": return "text-gray-500";
    default: return "text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700";
  }
};

/**
 * TypeBadge component.
 * @param props - The component props.
 * @param props.type - The type definition.
 * @param props.format - The format property.
 * @returns The rendered component.
 */
const TypeBadge = ({ type, format }: { type?: string | string[], format?: string }) => {
  if (!type) return null;
  const label = Array.isArray(type) ? type.join(" | ") : type;
  const displayLabel = format ? `${label} (${format})` : label;

  return (
    <span className={cn("text-[10px] px-1.5 py-0.5 rounded border font-mono uppercase tracking-wider select-none", getTypeColor(type))}>
      {displayLabel}
    </span>
  );
};

/**
 * SchemaViewer component.
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.name - The name.
 * @param props.required - Whether the field is required.
 * @param props.depth - The nesting depth.
 * @returns The rendered component.
 */
export function SchemaViewer({ schema, name, required = false, depth = 0 }: SchemaViewerProps) {
  const [isOpen, setIsOpen] = useState(true);

  if (!schema) return <div className="text-muted-foreground italic text-xs">No schema defined</div>;

  const isObject = schema.type === "object" || !!schema.properties;
  const isArray = schema.type === "array" || !!schema.items;
  const hasChildren = isObject || isArray;

  // Handle recursion for objects
  const properties = schema.properties ? Object.entries(schema.properties) : [];

  // Handle recursion for arrays
  const items = schema.items;

  return (
    <div className={cn("font-mono text-sm", depth > 0 && "ml-3 border-l-2 pl-3 border-border/20 dark:border-border/40")}>
      <div className="flex items-start py-1.5 group">
        {hasChildren ? (
           <Collapsible open={isOpen} onOpenChange={setIsOpen} className="w-full">
             <div className="flex items-center gap-2 select-none group-hover:bg-muted/20 p-1 rounded-md transition-colors -ml-1">
               <CollapsibleTrigger className="p-0.5 hover:bg-muted rounded transition-colors focus:outline-none focus:ring-1 focus:ring-ring">
                 {isOpen ? <ChevronDown className="h-3 w-3 text-muted-foreground/70" /> : <ChevronRight className="h-3 w-3 text-muted-foreground/70" />}
               </CollapsibleTrigger>
               {name && <span className="font-semibold text-foreground/90 tracking-tight">{name}</span>}
               {required && <span className="text-red-500/80 text-xs font-bold" title="Required">*</span>}
               <TypeBadge type={schema.type} format={schema.format} />
               {schema.description && (
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger asChild>
                        <Info className="h-3.5 w-3.5 text-muted-foreground/40 hover:text-foreground/80 transition-colors cursor-help" />
                      </TooltipTrigger>
                      <TooltipContent className="max-w-[300px] text-xs backdrop-blur-md bg-background/80 shadow-lg border-border/50">
                        <p className="leading-relaxed">{schema.description}</p>
                      </TooltipContent>
                    </Tooltip>
               )}
             </div>

             <CollapsibleContent>
               <div className="pt-1">
                 {isObject && properties.map(([key, propSchema], idx) => (
                   <SchemaViewer
                     key={key}
                     schema={propSchema}
                     name={key}
                     required={schema.required?.includes(key)}
                     depth={depth + 1}
                     isLast={idx === properties.length - 1}
                   />
                 ))}

                 {isArray && items && (
                   <div className="mt-2 relative bg-muted/5 dark:bg-muted/10 p-2 rounded-md border border-border/30">
                     <span className="text-[10px] uppercase font-bold tracking-[0.2em] text-muted-foreground/40 mb-2 block pl-1">ITEMS</span>
                     <SchemaViewer
                       schema={items}
                       depth={depth + 1}
                     />
                   </div>
                 )}
               </div>
             </CollapsibleContent>
           </Collapsible>
        ) : (
          <div className="flex items-center gap-2 relative group-hover:bg-muted/20 p-1 rounded-md transition-colors -ml-1 w-full">
             <span className="w-4 flex items-center justify-center shrink-0">
                 <span className="h-[2px] w-3 bg-border/20 dark:bg-border/40 rounded-full" />
             </span>
             {name && <span className="font-semibold text-foreground/80 tracking-tight shrink-0">{name}</span>}
             {required && <span className="text-red-500/80 text-xs font-bold shrink-0" title="Required">*</span>}
             <TypeBadge type={schema.type} format={schema.format} />
             {schema.enum && (
                <span className="text-[10px] font-mono text-muted-foreground/70 bg-muted/40 px-1.5 py-0.5 rounded border border-border/50 whitespace-nowrap overflow-hidden text-ellipsis max-w-[200px]" title={schema.enum.join(", ")}>
                  enum: [{schema.enum.join(", ")}]
                </span>
             )}
              {schema.description && (
                    <Tooltip delayDuration={300}>
                      <TooltipTrigger asChild>
                        <Info className="h-3.5 w-3.5 text-muted-foreground/40 hover:text-foreground/80 transition-colors cursor-help shrink-0" />
                      </TooltipTrigger>
                      <TooltipContent className="max-w-[300px] text-xs backdrop-blur-md bg-background/80 shadow-lg border-border/50">
                        <p className="leading-relaxed">{schema.description}</p>
                      </TooltipContent>
                    </Tooltip>
               )}
          </div>
        )}
      </div>
    </div>
  );
}
