"use client";

import * as React from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import ReactSyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import markdown from 'react-syntax-highlighter/dist/esm/languages/hljs/markdown';
import { vs2015 } from 'react-syntax-highlighter/dist/esm/styles/hljs';

ReactSyntaxHighlighter.registerLanguage('markdown', markdown);

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function MarkdownEditor({ value, onChange, placeholder, className }: MarkdownEditorProps) {
  const [activeTab, setActiveTab] = React.useState("edit");

  return (
    <div className={cn("flex flex-col border rounded-md overflow-hidden", className)}>
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex items-center justify-between px-2 bg-muted/50 border-b">
          <TabsList className="h-8 bg-transparent p-0">
            <TabsTrigger
              value="edit"
              className="h-8 rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 text-xs"
            >
              Edit
            </TabsTrigger>
            <TabsTrigger
              value="preview"
              className="h-8 rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent px-4 text-xs"
            >
              Preview
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="edit" className="flex-1 mt-0 relative">
          <Textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            className="min-h-[300px] h-full rounded-none border-0 focus-visible:ring-0 resize-none font-mono text-sm p-4"
          />
        </TabsContent>

        <TabsContent value="preview" className="flex-1 mt-0 overflow-auto bg-card">
          <div className="min-h-[300px] h-full p-4 prose dark:prose-invert max-w-none text-sm">
            {value ? (
                <ReactSyntaxHighlighter
                    language="markdown"
                    style={vs2015}
                    customStyle={{ background: 'transparent', padding: 0 }}
                    wrapLines={true}
                >
                    {value}
                </ReactSyntaxHighlighter>
            ) : (
                <div className="text-muted-foreground italic">Nothing to preview.</div>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
