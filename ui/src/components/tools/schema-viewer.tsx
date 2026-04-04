/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React, { useState } from "react";
import { ChevronRight, ChevronDown, Info } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * Intent: Document Schema
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
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

function SchemaRow({ schema, name, required = false, depth = 0 }: SchemaViewerProps) {
    const [isOpen, setIsOpen] = useState(true);

    if (!schema) return null;

    const isObject = schema.type === "object" || !!schema.properties;
    const isArray = schema.type === "array" || !!schema.items;
    const hasChildren = isObject || isArray;

    const properties = schema.properties ? Object.entries(schema.properties) : [];
    const items = schema.items;

    return (
        <>
            <tr className={cn("border-b border-border/40 hover:bg-muted/50 transition-colors", depth === 0 ? "bg-background" : "bg-muted/10")}>
                <td className="py-2.5 px-4 align-top">
                    <div className="flex items-center gap-2" style={{ paddingLeft: `${depth * 1.5}rem` }}>
                        {hasChildren ? (
                            <button
                                onClick={() => setIsOpen(!isOpen)}
                                className="p-0.5 hover:bg-muted rounded text-muted-foreground transition-colors"
                            >
                                {isOpen ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
                            </button>
                        ) : (
                            <span className="w-5" /> // spacer
                        )}
                        <span className="font-mono text-sm font-semibold text-foreground">
                            {name || (depth === 0 ? "root" : "items")}
                        </span>
                        {required && <span className="text-red-500 text-xs font-bold" title="Required">*</span>}
                    </div>
                </td>
                <td className="py-2.5 px-4 align-top w-[150px]">
                    <TypeBadge type={schema.type} format={schema.format} />
                </td>
                <td className="py-2.5 px-4 align-top text-sm text-muted-foreground max-w-[300px]">
                    <div className="flex flex-col gap-1">
                        {schema.description && <span>{schema.description}</span>}
                        {schema.enum && (
                            <span className="text-xs opacity-80 mt-1">
                                <strong className="font-medium">Enum:</strong> [{schema.enum.join(", ")}]
                            </span>
                        )}
                         {schema.default !== undefined && (
                            <span className="text-xs opacity-80 mt-1">
                                <strong className="font-medium">Default:</strong> {JSON.stringify(schema.default)}
                            </span>
                        )}
                    </div>
                </td>
            </tr>
            {isOpen && hasChildren && (
                <>
                    {isObject && properties.map(([key, propSchema], idx) => (
                        <SchemaRow
                            key={key}
                            schema={propSchema}
                            name={key}
                            required={schema.required?.includes(key)}
                            depth={depth + 1}
                            isLast={idx === properties.length - 1}
                        />
                    ))}
                    {isArray && items && (
                        <SchemaRow
                            schema={items}
                            depth={depth + 1}
                        />
                    )}
                </>
            )}
        </>
    );
}

/**
 * Intent: Document SchemaViewer
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
 * SchemaViewer component.
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @param props.name - The name.
 * @param props.required - Whether the field is required.
 * @param props.depth - The nesting depth.
 * @returns The rendered component.
 */
export function SchemaViewer({ schema, name, required = false, depth = 0 }: SchemaViewerProps) {
  if (!schema) return <div className="text-muted-foreground italic text-xs p-4">No schema defined</div>;

  return (
    <div className="w-full overflow-x-auto rounded-md border border-border bg-card shadow-sm">
        <table className="w-full text-left border-collapse">
            <thead>
                <tr className="bg-muted/50 border-b border-border/60 text-xs uppercase tracking-wider text-muted-foreground">
                    <th className="py-2 px-4 font-semibold w-1/3 min-w-[200px]">Property</th>
                    <th className="py-2 px-4 font-semibold w-[150px]">Type</th>
                    <th className="py-2 px-4 font-semibold">Description</th>
                </tr>
            </thead>
            <tbody>
                <SchemaRow schema={schema} name={name} required={required} depth={depth} />
            </tbody>
        </table>
    </div>
  );
}
