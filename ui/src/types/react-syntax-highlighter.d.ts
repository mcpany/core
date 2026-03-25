/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Type declarations for react-syntax-highlighter ESM sub-path imports
declare module 'react-syntax-highlighter/dist/esm/prism-light' {
  import { ComponentType } from 'react';
  export interface SyntaxHighlighterProps {
    language?: string;
    style?: Record<string, unknown>;
    children?: string;
    customStyle?: Record<string, unknown>;
    wrapLines?: boolean;
    wrapLongLines?: boolean;
    showLineNumbers?: boolean;
    lineNumberStyle?: Record<string, unknown>;
    [key: string]: unknown;
  }
  const PrismLight: ComponentType<SyntaxHighlighterProps> & {
    registerLanguage: (name: string, language: unknown) => void;
  };
  export default PrismLight;
}

declare module 'react-syntax-highlighter/dist/esm/light' {
  import { ComponentType } from 'react';
  export interface SyntaxHighlighterProps {
    language?: string;
    style?: Record<string, unknown>;
    children?: string;
    customStyle?: Record<string, unknown>;
    wrapLines?: boolean;
    wrapLongLines?: boolean;
    showLineNumbers?: boolean;
    [key: string]: unknown;
  }
  const Light: ComponentType<SyntaxHighlighterProps> & {
    registerLanguage: (name: string, language: unknown) => void;
  };
  export default Light;
}

declare module 'react-syntax-highlighter/dist/esm/languages/prism/json' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/yaml' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/javascript' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/python' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/bash' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/typescript' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/jsx' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/tsx' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/prism/markdown' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/json' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/yaml' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/xml' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/markdown' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/javascript' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/python' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/bash' { const v: unknown; export default v; }
declare module 'react-syntax-highlighter/dist/esm/languages/hljs/plaintext' { const v: unknown; export default v; }

declare module 'react-syntax-highlighter/dist/esm/styles/prism' {
  export const vscDarkPlus: Record<string, unknown>;
  export const oneDark: Record<string, unknown>;
  const styles: Record<string, Record<string, unknown>>;
  export default styles;
}

declare module 'react-syntax-highlighter/dist/esm/styles/hljs' {
  export const docco: Record<string, unknown>;
  export const dark: Record<string, unknown>;
  export const vs2015: Record<string, unknown>;
  const styles: Record<string, Record<string, unknown>>;
  export default styles;
}

declare module 'react-syntax-highlighter/dist/esm/styles/hljs/vs2015' {
  const vs2015: Record<string, unknown>;
  export default vs2015;
}
