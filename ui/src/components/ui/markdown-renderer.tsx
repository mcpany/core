/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { cn } from '@/lib/utils';
import { Copy, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

/**
 * MarkdownRenderer component.
 * Renders Markdown content with syntax highlighting for code blocks.
 * @param props - The component props.
 * @returns The rendered Markdown content.
 */
export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
  return (
    <div className={cn("markdown-body space-y-3", className)}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h1: ({ node, ...props }) => <h1 className="text-xl font-bold mt-4 mb-2" {...props} />,
        h2: ({ node, ...props }) => <h2 className="text-lg font-bold mt-3 mb-2" {...props} />,
        h3: ({ node, ...props }) => <h3 className="text-base font-bold mt-2 mb-1" {...props} />,
        p: ({ node, ...props }) => <p className="leading-relaxed whitespace-pre-wrap break-words" {...props} />,
        ul: ({ node, ...props }) => <ul className="list-disc pl-5 space-y-1" {...props} />,
        ol: ({ node, ...props }) => <ol className="list-decimal pl-5 space-y-1" {...props} />,
        li: ({ node, ...props }) => <li className="" {...props} />,
        a: ({ node, ...props }) => <a className="text-primary underline hover:text-primary/80" target="_blank" rel="noopener noreferrer" {...props} />,
        blockquote: ({ node, ...props }) => <blockquote className="border-l-4 border-primary/30 pl-4 italic text-muted-foreground my-2" {...props} />,
        table: ({ node, ...props }) => <div className="overflow-x-auto my-4"><table className="min-w-full border-collapse border border-border" {...props} /></div>,
        thead: ({ node, ...props }) => <thead className="bg-muted" {...props} />,
        tbody: ({ node, ...props }) => <tbody className="bg-background" {...props} />,
        tr: ({ node, ...props }) => <tr className="border-b border-border" {...props} />,
        th: ({ node, ...props }) => <th className="px-4 py-2 text-left text-sm font-semibold border-r border-border last:border-0" {...props} />,
        td: ({ node, ...props }) => <td className="px-4 py-2 text-sm border-r border-border last:border-0" {...props} />,
        code({ node, inline, className, children, ...props }: any) {
          const match = /language-(\w+)/.exec(className || '');
          const [copied, setCopied] = React.useState(false);

          const handleCopy = () => {
             const text = String(children).replace(/\n$/, '');
             navigator.clipboard.writeText(text);
             setCopied(true);
             setTimeout(() => setCopied(false), 2000);
          };

          return !inline && match ? (
            <div className="relative group/code my-4 rounded-md overflow-hidden border border-border">
              <div className="flex items-center justify-between px-3 py-1.5 bg-muted/50 border-b border-border text-xs text-muted-foreground">
                 <span className="font-mono">{match[1]}</span>
                 <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 hover:bg-muted text-muted-foreground"
                    onClick={handleCopy}
                 >
                    {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                 </Button>
              </div>
              <SyntaxHighlighter
                style={vscDarkPlus}
                language={match[1]}
                PreTag="div"
                customStyle={{ margin: 0, padding: '1rem', fontSize: '13px', background: 'var(--code-bg, #1e1e1e)' }}
                {...props}
              >
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            </div>
          ) : (
            <code className={cn("bg-muted px-1.5 py-0.5 rounded text-sm font-mono text-primary", className)} {...props}>
              {children}
            </code>
          );
        }
      }}
    >
      {content}
    </ReactMarkdown>
    </div>
  );
}
