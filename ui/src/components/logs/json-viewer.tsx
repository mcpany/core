/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { JsonView } from "@/components/ui/json-view";

interface JsonViewerProps {
  data: unknown;
}

/**
 * JsonViewer is a component that renders JSON data with syntax highlighting.
 * @deprecated Use JsonView from "@/components/ui/json-view" instead.
 *
 * @param props - The component props.
 * @param props.data - The JSON data to display.
 * @returns A syntax-highlighted JSON view.
 */
export default function JsonViewer({ data }: JsonViewerProps) {
  return (
    <JsonView data={data} />
  );
}
