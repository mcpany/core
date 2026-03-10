/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Eye, Loader2 } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ResourceContent } from "@/lib/client";

import { PrismLight } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import yaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml';
import markdown from 'react-syntax-highlighter/dist/esm/languages/prism/markdown';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

const SyntaxHighlighter = PrismLight as any;

SyntaxHighlighter.registerLanguage('json', json);
SyntaxHighlighter.registerLanguage('yaml', yaml);
SyntaxHighlighter.registerLanguage('xml', yaml);
SyntaxHighlighter.registerLanguage('markdown', markdown);
SyntaxHighlighter.registerLanguage('text', yaml);

interface ResourceViewerProps {
    content: ResourceContent | null;
    loading: boolean;
}

/**
 * ResourceViewer.
 *
 * @param loading - The loading.
 */
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
        let src = "";
        if (content.blob) {
            src = `data:${mimeType};base64,${content.blob}`;
        } else if (text && text.startsWith("data:image")) {
            src = text;
        }

        if (src) {
            return (
                <div className="flex items-center justify-center h-full bg-checkered p-4 overflow-hidden">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                        src={src}
                        alt={uri}
                        className="max-w-full max-h-full object-contain shadow-sm rounded-md border"
                    />
                </div>
            );
        }

        return (
            <div className="flex items-center justify-center h-full bg-checkered p-4">
                <div className="text-muted-foreground italic">Image content not available.</div>
            </div>
        );
    }

    if (mimeType.includes("json") || mimeType.includes("yaml") || mimeType.includes("xml")) {
         return (
            <ScrollArea className="h-full">
                <SyntaxHighlighter
                    language={mimeType.includes("json") ? "json" : "yaml"}
                    style={vscDarkPlus as any}
                    customStyle={{ margin: 0, borderRadius: 0, height: "100%", fontSize: '0.875rem' }}
                    showLineNumbers={true}
                >
                    {text || ""}
                </SyntaxHighlighter>
            </ScrollArea>
        );
    }

    // Markdown
     if (mimeType.includes("markdown") || uri?.endsWith(".md")) {
         return (
            <ScrollArea className="h-full p-6">
                <div className="prose dark:prose-invert max-w-none">
                     <SyntaxHighlighter
                        language="markdown"
                        style={vscDarkPlus as any}
                        customStyle={{ background: 'transparent', padding: 0 }}
                        wrapLines={true}
                    >
                        {text || ""}
                    </SyntaxHighlighter>
                </div>
            </ScrollArea>
        );
    }

    // Code / Plain Text
    return (
         <ScrollArea className="h-full">
             <SyntaxHighlighter
                language="text" // generic
                style={vscDarkPlus as any}
                customStyle={{ margin: 0, borderRadius: 0, height: "100%", fontSize: '0.875rem' }}
                showLineNumbers={true}
            >
                {text || ""}
            </SyntaxHighlighter>
        </ScrollArea>
    );
}
