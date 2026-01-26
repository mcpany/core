/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

/**
 * Props for the JsonViewer component.
 */
interface JsonViewerProps {
  /** The JSON data to display. Can be any valid JSON value. */
  data: unknown;
}

/**
 * A component that renders JSON data with syntax highlighting.
 *
 * @param props - The component props.
 * @returns A syntax-highlighted JSON view.
 */
export default function JsonViewer({ data }: JsonViewerProps) {
  return (
    <SyntaxHighlighter
      language="json"
      style={vscDarkPlus}
      customStyle={{
        margin: 0,
        padding: '1rem',
        borderRadius: '0.5rem',
        backgroundColor: '#1e1e1e', // Dark background
        fontSize: '12px',
        lineHeight: '1.5'
      }}
      wrapLongLines={true}
    >
      {JSON.stringify(data, null, 2)}
    </SyntaxHighlighter>
  );
}
