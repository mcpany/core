/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { MCPMessage } from "@/hooks/use-inspector-stream";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

interface MessageDetailProps {
    message: MCPMessage | null;
    onClose: () => void;
}

export function MessageDetail({ message, onClose }: MessageDetailProps) {
    if (!message) {
        return (
            <div className="h-full flex items-center justify-center text-muted-foreground bg-muted/5 p-4">
                Select a message to view details
            </div>
        );
    }

    return (
        <div className="h-full flex flex-col bg-background/50 backdrop-blur-sm">
            <div className="flex items-center justify-between p-4 border-b">
                <div className="flex items-center gap-2">
                    <CardTitle className="text-sm font-medium">Message Details</CardTitle>
                    <Badge variant="outline" className="font-mono text-xs">
                        {message.id}
                    </Badge>
                </div>
                <Button variant="ghost" size="icon" onClick={onClose} className="h-6 w-6">
                    <X className="h-4 w-4" />
                </Button>
            </div>

            <div className="flex-1 overflow-auto p-4 space-y-4">
                <Card>
                    <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                        <CardTitle className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Header</CardTitle>
                    </CardHeader>
                    <CardContent className="p-4 grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <span className="text-muted-foreground block text-xs mb-1">Timestamp</span>
                            <span className="font-mono">{new Date(message.timestamp).toLocaleString()}</span>
                        </div>
                        <div>
                            <span className="text-muted-foreground block text-xs mb-1">Direction</span>
                            <Badge variant={message.direction === 'inbound' ? 'secondary' : 'default'}>
                                {message.direction.toUpperCase()}
                            </Badge>
                        </div>
                        {message.method && (
                            <div className="col-span-2">
                                <span className="text-muted-foreground block text-xs mb-1">Method</span>
                                <span className="font-mono bg-muted px-2 py-1 rounded">{message.method}</span>
                            </div>
                        )}
                        {message.source && (
                             <div className="col-span-2">
                                <span className="text-muted-foreground block text-xs mb-1">Source</span>
                                <span className="font-mono text-xs">{message.source}</span>
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card className="flex-1 flex flex-col overflow-hidden">
                    <CardHeader className="py-3 px-4 bg-muted/20 border-b">
                        <CardTitle className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Payload</CardTitle>
                    </CardHeader>
                    <CardContent className="p-0 overflow-auto bg-muted">
                         <SyntaxHighlighter
                            language="json"
                            style={vscDarkPlus}
                            customStyle={{
                                margin: 0,
                                padding: '1rem',
                                fontSize: '12px',
                                lineHeight: '1.5',
                                backgroundColor: 'transparent'
                            }}
                            wrapLongLines={true}
                        >
                            {JSON.stringify(message.payload, null, 2)}
                        </SyntaxHighlighter>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
