/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

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
interface SchemaRow {
  key: string;
  name: string;
  type?: string | string[];
  description?: string;
  required: boolean;
  depth: number;
  format?: string;
  enum?: any[];
}

function flattenSchema(schema: Schema, keyPath: string = "", name: string = "", depth: number = 0, isRequired: boolean = false): SchemaRow[] {
  let rows: SchemaRow[] = [];
  if (!schema) return rows;

  const isObject = schema.type === "object" || !!schema.properties;
  const isArray = schema.type === "array" || !!schema.items;

  if (name) {
    rows.push({
      key: keyPath,
      name: name,
      type: schema.type,
      description: schema.description,
      required: isRequired,
      depth,
      format: schema.format,
      enum: schema.enum
    });
  }

  if (isObject && schema.properties) {
    Object.entries(schema.properties).forEach(([k, propSchema]) => {
      const newKeyPath = keyPath ? `${keyPath}.${k}` : k;
      const propRequired = schema.required?.includes(k) || false;
      const nextDepth = name ? depth + 1 : depth;
      rows = rows.concat(flattenSchema(propSchema, newKeyPath, k, nextDepth, propRequired));
    });
  }

  if (isArray && schema.items) {
    const newKeyPath = keyPath ? `${keyPath}[]` : "[]";
    const nextDepth = name ? depth + 1 : depth;
    rows = rows.concat(flattenSchema(schema.items, newKeyPath, "[] (items)", nextDepth, false));
  }

  return rows;
}

/**
 * SchemaViewer component.
 * @param props - The component props.
 * @param props.schema - The schema definition.
 * @returns The rendered component.
 */
export function SchemaViewer({ schema }: { schema: Schema }) {
  if (!schema) return <div className="text-muted-foreground italic text-xs p-4">No schema defined</div>;

  const rows = flattenSchema(schema);

  if (rows.length === 0) {
    return <div className="text-muted-foreground italic text-xs p-4">Empty schema</div>;
  }

  return (
    <div className="rounded-md border overflow-hidden">
      <Table>
        <TableHeader className="bg-muted/50">
          <TableRow>
            <TableHead className="w-[30%]">Property</TableHead>
            <TableHead className="w-[20%]">Type</TableHead>
            <TableHead className="w-[10%]">Required</TableHead>
            <TableHead className="w-[40%]">Description</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key} className="group hover:bg-muted/50 transition-colors">
              <TableCell className="font-mono text-sm py-2">
                <div className="flex items-center" style={{ paddingLeft: `${row.depth * 1.5}rem` }}>
                  {row.depth > 0 && <span className="text-muted-foreground/40 mr-2">↳</span>}
                  <span className="font-semibold">{row.name}</span>
                </div>
              </TableCell>
              <TableCell className="py-2">
                <TypeBadge type={row.type} format={row.format} />
              </TableCell>
              <TableCell className="py-2">
                {row.required ? (
                  <span className="text-red-500 font-bold text-xs uppercase tracking-wider">Yes</span>
                ) : (
                  <span className="text-muted-foreground text-xs uppercase tracking-wider">No</span>
                )}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground py-2">
                <div className="flex flex-col gap-1">
                  {row.description && <span>{row.description}</span>}
                  {row.enum && (
                    <span className="text-xs font-mono bg-muted/50 px-1.5 py-0.5 rounded border w-fit">
                      Enum: [{row.enum.join(", ")}]
                    </span>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
