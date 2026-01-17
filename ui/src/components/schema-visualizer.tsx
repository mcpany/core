/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ChevronDown, ChevronRight } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";

interface SchemaNodeProps {
  name?: string;
  schema: any;
  required?: boolean;
  level?: number;
}

const TypeBadge = ({ type }: { type: string }) => {
  const colors: Record<string, string> = {
    string: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300",
    number: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300",
    integer: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300",
    boolean: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300",
    array: "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300",
    object: "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300",
    null: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300",
  };

  const colorClass = colors[type] || "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300";

  return (
    <span className={cn("px-2 py-0.5 rounded-full text-xs font-mono font-medium", colorClass)}>
      {type}
    </span>
  );
};

const SchemaNode = ({ name, schema, required, level = 0 }: SchemaNodeProps) => {
  const [expanded, setExpanded] = useState(true);

  if (!schema) return null;

  const isObject = schema.type === "object" && schema.properties;
  const isArray = schema.type === "array" && schema.items;
  const hasChildren = isObject || isArray;

  const type = Array.isArray(schema.type) ? schema.type.join(" | ") : schema.type;

  return (
    <>
      <TableRow className="hover:bg-muted/50 transition-colors">
        <TableCell className="font-mono text-sm align-top py-2" style={{ paddingLeft: `${level * 1.5 + 0.5}rem` }}>
          <div className="flex items-center gap-2">
            {hasChildren && (
              <button
                onClick={() => setExpanded(!expanded)}
                className="p-0.5 hover:bg-muted rounded-sm transition-colors"
              >
                {expanded ? (
                  <ChevronDown className="h-3 w-3 text-muted-foreground" />
                ) : (
                  <ChevronRight className="h-3 w-3 text-muted-foreground" />
                )}
              </button>
            )}
            {!hasChildren && <span className="w-4" />} {/* Spacer for indentation */}
            <span className={cn(required && "font-bold text-primary")}>
              {name || "root"}
            </span>
            {required && <span className="text-destructive text-xs ml-1">*</span>}
          </div>
        </TableCell>
        <TableCell className="align-top py-2">
          <TypeBadge type={type || "any"} />
        </TableCell>
        <TableCell className="text-muted-foreground text-sm align-top py-2">
          <div className="flex flex-col gap-1">
             <span>{schema.description}</span>
             {schema.enum && (
                 <div className="text-xs mt-1">
                    <span className="font-semibold text-muted-foreground/80">Allowed values: </span>
                    <code className="bg-muted px-1 rounded text-xs break-all">{JSON.stringify(schema.enum)}</code>
                 </div>
             )}
              {schema.default !== undefined && (
                  <div className="text-xs text-muted-foreground/80">
                      Default: <code className="bg-muted px-1 rounded text-xs">{JSON.stringify(schema.default)}</code>
                  </div>
              )}
          </div>
        </TableCell>
      </TableRow>

      {expanded && isObject && schema.properties && (
        <>
          {Object.entries(schema.properties).map(([propName, propSchema]: [string, any]) => (
            <SchemaNode
              key={propName}
              name={propName}
              schema={propSchema}
              required={schema.required?.includes(propName)}
              level={level + 1}
            />
          ))}
        </>
      )}

      {expanded && isArray && (
        <SchemaNode
          name="items"
          schema={schema.items}
          level={level + 1}
        />
      )}
    </>
  );
};

/**
 * SchemaVisualizer.
 *
 * @param { schema - The { schema.
 */
export function SchemaVisualizer({ schema }: { schema: any }) {
  if (!schema || Object.keys(schema).length === 0) {
    return (
      <div className="text-muted-foreground text-sm italic p-4 text-center border rounded-md bg-muted/20">
        No input schema defined.
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[250px]">Property</TableHead>
            <TableHead className="w-[100px]">Type</TableHead>
            <TableHead>Description</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
            {/* If the root schema is an object, we usually want to show its properties directly, not the root object itself as a row.
                However, if it's not an object (e.g. array), we show it.
            */}
            {schema.type === 'object' && schema.properties ? (
                 Object.entries(schema.properties).map(([propName, propSchema]: [string, any]) => (
                    <SchemaNode
                      key={propName}
                      name={propName}
                      schema={propSchema}
                      required={schema.required?.includes(propName)}
                      level={0}
                    />
                  ))
            ) : (
                 <SchemaNode name="input" schema={schema} />
            )}
        </TableBody>
      </Table>
    </div>
  );
}
