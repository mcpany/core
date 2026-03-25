/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */




import { useMemo, memo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UpstreamServiceConfig } from "@/lib/types";
import { File } from "lucide-react";
import yaml from 'js-yaml';
import { JsonView } from '@/components/ui/json-view';
import SyntaxHighlighter from 'react-syntax-highlighter/dist/esm/light';
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json';
import yamlLang from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml';
import vs2015 from 'react-syntax-highlighter/dist/esm/styles/hljs/vs2015';
import { ScrollArea } from "./ui/scroll-area";

SyntaxHighlighter.registerLanguage('json', json);
SyntaxHighlighter.registerLanguage('yaml', yamlLang);

/**
 * Converts an object to a TextProto formatted string.
 * @param obj The object to convert.
 * @param indent The current indentation level.
 * @returns The TextProto string representation.
 */
function objectToTextProto(obj: any, indent = 0): string {
  const indentStr = '  '.repeat(indent);
  let protoStr = '';

  for (const key in obj) {
    if (obj[key] === null || obj[key] === undefined || (Array.isArray(obj[key]) && obj[key].length === 0)) {
      continue;
    }

    const value = obj[key];

    if (Array.isArray(value)) {
      for (const item of value) {
        if (typeof item === 'object' && item !== null) {
          protoStr += `${indentStr}${key}: {\n`;
          protoStr += objectToTextProto(item, indent + 1);
          protoStr += `${indentStr}}\n`;
        } else {
          protoStr += `${indentStr}${key}: ${JSON.stringify(item)}\n`;
        }
      }
    } else if (typeof value === 'object') {
      protoStr += `${indentStr}${key}: {\n`;
      protoStr += objectToTextProto(value, indent + 1);
      protoStr += `${indentStr}}\n`;
    } else {
      protoStr += `${indentStr}${key}: ${typeof value === 'string' ? `"${value}"` : value}\n`;
    }
  }

  return protoStr;
}


/**
 * Displays a code block with syntax highlighting.
 * @param props.language The language for syntax highlighting (e.g., 'json', 'yaml').
 * @param props.code The code content to display.
 */
function CodeBlock({ language, code }: { language: string; code: string }) {
    return (
        <ScrollArea className="h-72 w-full rounded-md border bg-background/50">
             <SyntaxHighlighter language={language} style={vs2015} showLineNumbers customStyle={{ background: 'transparent', margin: 0, padding: '1rem' }}>
                {code}
            </SyntaxHighlighter>
        </ScrollArea>
    )
}

/**
 * Displays the configuration of a service in multiple formats (YAML, JSON, TextProto).
 *
 * @param props - The component props.
 * @param props.service - The service configuration to display.
 * @returns {JSX.Element} The rendered config card.
 */
export const FileConfigCard = memo(function FileConfigCard({ service }: { service: UpstreamServiceConfig }) {
    const { jsonConfig, yamlConfig, textProtoConfig } = useMemo(() => {
        const tempService = structuredClone(service);
        delete (tempService as { id?: string }).id;

        return {
            jsonConfig: JSON.stringify(tempService, null, 2),
            yamlConfig: yaml.dump(tempService),
            textProtoConfig: objectToTextProto(tempService)
        };
    }, [service]);

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-xl flex items-center gap-2"><File /> File Config</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="mb-4">
                    <p className="text-sm text-muted-foreground">
                        Filesystem services allow defining specific allowed and denied paths.
                        You can edit these in the configuration JSON below.
                    </p>
                </div>
                <Tabs defaultValue="yaml">
                    <TabsList>
                        <TabsTrigger value="yaml">YAML</TabsTrigger>
                        <TabsTrigger value="json">JSON</TabsTrigger>
                        <TabsTrigger value="textproto">Text Proto</TabsTrigger>
                    </TabsList>
                    <TabsContent value="yaml" className="mt-4">
                        <CodeBlock language="yaml" code={yamlConfig} />
                    </TabsContent>
                    <TabsContent value="json" className="mt-4">
                        <JsonView data={service} maxHeight={500} />
                    </TabsContent>
                    <TabsContent value="textproto" className="mt-4">
                        <CodeBlock language="protobuf" code={textProtoConfig} />
                    </TabsContent>
                </Tabs>
            </CardContent>
        </Card>
    )
});
