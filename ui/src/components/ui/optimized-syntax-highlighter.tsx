/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { PrismLight as SyntaxHighlighter } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import yaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml';
import protobuf from 'react-syntax-highlighter/dist/esm/languages/prism/protobuf';
import markdown from 'react-syntax-highlighter/dist/esm/languages/prism/markdown';

// ⚡ BOLT: Optimized imports to reduce bundle size.
// Only register the necessary languages for highlighting, avoiding the heavy full bundle.
SyntaxHighlighter.registerLanguage('json', json);
SyntaxHighlighter.registerLanguage('yaml', yaml);
SyntaxHighlighter.registerLanguage('protobuf', protobuf);
SyntaxHighlighter.registerLanguage('markdown', markdown);

export default SyntaxHighlighter;
