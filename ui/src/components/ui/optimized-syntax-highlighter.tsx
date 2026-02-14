/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import SyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';

// ⚡ BOLT: Optimized imports to reduce bundle size.
// Only register the 'json' language for highlighting, avoiding the heavy full bundle.
SyntaxHighlighter.registerLanguage('json', json);

export default SyntaxHighlighter;
