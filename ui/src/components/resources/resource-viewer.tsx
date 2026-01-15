/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Eye, Loader2 } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ResourceContent } from "@/lib/client";

import ReactSyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';
import xml from 'react-syntax-highlighter/dist/esm/languages/hljs/xml';
import markdown from 'react-syntax-highlighter/dist/esm/languages/hljs/markdown';
import plaintext from 'react-syntax-highlighter/dist/esm/languages/hljs/plaintext';
import { vs2015 } from 'react-syntax-highlighter/dist/esm/styles/hljs';

ReactSyntaxHighlighter.registerLanguage('json', json);
ReactSyntaxHighlighter.registerLanguage('yaml', yaml);
ReactSyntaxHighlighter.registerLanguage('xml', xml);
ReactSyntaxHighlighter.registerLanguage('markdown', markdown);
ReactSyntaxHighlighter.registerLanguage('text', plaintext);

interface ResourceViewerProps {
    content: ResourceContent | null;
    loading: boolean;
}

export function ResourceViewer({ content, loading }: ResourceViewerProps) {
    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
                <p>Loading content...</p>
            </div>
        );
    }

    if (!content) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2 p-8 text-center">
                <Eye className="h-12 w-12 opacity-20" />
                <p>Select a resource to view its content.</p>
            </div>
        );
    }

    const { mimeType, text, uri } = content;

    if (mimeType.startsWith("image/")) {
        // Since we mocked text content, we can't really show an image unless it's a blob URL.
        // But let's assume we might handle base64 in future.
        return (
             <div className="flex items-center justify-center h-full bg-checkered p-4">
                 <div className="text-muted-foreground italic">Image preview not supported in this demo.</div>
             </div>
        );
    }

    if (mimeType.includes("json") || mimeType.includes("yaml") || mimeType.includes("xml")) {
         return (
            <ScrollArea className="h-full">
                <ReactSyntaxHighlighter
                    language={mimeType.includes("json") ? "json" : "yaml"}
                    style={vs2015}
                    customStyle={{ margin: 0, borderRadius: 0, height: "100%", fontSize: '0.875rem' }}
                    showLineNumbers={true}
                >
                    {text || ""}
                </ReactSyntaxHighlighter>
            </ScrollArea>
        );
    }

    // Markdown
     if (mimeType.includes("markdown") || uri?.endsWith(".md")) {
         return (
            <ScrollArea className="h-full p-6">
                <div className="prose dark:prose-invert max-w-none">
                     <ReactSyntaxHighlighter
                        language="markdown"
                        style={vs2015}
                        customStyle={{ background: 'transparent', padding: 0 }}
                        wrapLines={true}
                    >
                        {text || ""}
                    </ReactSyntaxHighlighter>
                </div>
            </ScrollArea>
        );
    }

    // Code / Plain Text
    return (
         <ScrollArea className="h-full">
             <ReactSyntaxHighlighter
                language="text" // generic
                style={vs2015}
                customStyle={{ margin: 0, borderRadius: 0, height: "100%", fontSize: '0.875rem' }}
                showLineNumbers={true}
            >
                {text || ""}
            </ReactSyntaxHighlighter>
        </ScrollArea>
    );
}
