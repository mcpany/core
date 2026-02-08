"use client";

import { ToolDefinition } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Play, FileJson } from "lucide-react";
import Link from "next/link";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';


interface ServiceToolsProps {
    tools: ToolDefinition[];
}

/**
 * ServiceTools lists the tools exposed by an upstream service.
 * It provides actions to try out tools in the playground and view their schema definitions.
 */
export function ServiceTools({ tools }: ServiceToolsProps) {
    if (!tools || tools.length === 0) {
        return <div className="text-center py-10 text-muted-foreground">No tools discovered for this service.</div>;
    }

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {tools.map((tool) => (
                <Card key={tool.name} className="flex flex-col">
                    <CardHeader className="pb-3">
                        <CardTitle className="flex items-center justify-between text-base">
                            <span className="truncate" title={tool.name}>{tool.name}</span>
                        </CardTitle>
                        <CardDescription className="line-clamp-2 min-h-[40px]">
                            {tool.description || "No description provided."}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1 flex items-end gap-2 mt-auto pt-0">
                         <div className="flex gap-2 w-full">
                            <Link href={`/playground?tool=${tool.name}`} className="flex-1">
                                <Button variant="outline" className="w-full">
                                    <Play className="mr-2 h-4 w-4" /> Try
                                </Button>
                            </Link>

                            <Dialog>
                                <DialogTrigger asChild>
                                    <Button variant="ghost" size="icon" title="View Schema">
                                        <FileJson className="h-4 w-4" />
                                    </Button>
                                </DialogTrigger>
                                <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                                    <DialogHeader>
                                        <DialogTitle>{tool.name}</DialogTitle>
                                        <DialogDescription>
                                            Input schema definition.
                                        </DialogDescription>
                                    </DialogHeader>
                                    <div className="rounded-md border overflow-hidden">
                                         <SyntaxHighlighter
                                            language="json"
                                            style={vscDarkPlus}
                                            customStyle={{ margin: 0, fontSize: '12px' }}
                                        >
                                            {JSON.stringify(tool.inputSchema || {}, null, 2)}
                                        </SyntaxHighlighter>
                                    </div>
                                </DialogContent>
                            </Dialog>
                         </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
