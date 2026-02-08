/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useCallback } from "react";
import { ToolDefinition } from "@/lib/client";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Upload, Zap, FileText, Image as ImageIcon } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";

interface SmartFileDropProps {
    children: React.ReactNode;
    tools: ToolDefinition[];
    onToolSelect: (tool: ToolDefinition, file: File, field: string) => void;
}

interface ToolMatch {
    tool: ToolDefinition;
    field: string;
}

/**
 * SmartFileDrop component.
 * Wraps content with a drag-and-drop zone that suggests tools compatible with the dropped file.
 */
export function SmartFileDrop({ children, tools, onToolSelect }: SmartFileDropProps) {
    const [isDragging, setIsDragging] = useState(false);
    const [droppedFile, setDroppedFile] = useState<File | null>(null);
    const [matches, setMatches] = useState<ToolMatch[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const { toast } = useToast();

    const handleDragEnter = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(true);
    }, []);

    const handleDragOver = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        if (!isDragging) setIsDragging(true);
    }, [isDragging]);

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        // Check if leaving the window or just the element
        if (e.currentTarget.contains(e.relatedTarget as Node)) return;
        setIsDragging(false);
    }, []);

    const isCompatibleSchema = (schema: any, file: File): boolean => {
        if (schema.type !== 'string') return false;

        let isBase64 = false;
        if (schema.contentEncoding === 'base64' || schema.format === 'binary') {
            isBase64 = true;
        }

        if (schema.contentMediaType) {
            const accepted = schema.contentMediaType;
            const fileType = file.type;

            // Allow wildcard
            const typeMatch = accepted === fileType ||
                              (accepted.endsWith('/*') && fileType.startsWith(accepted.split('/')[0] + '/'));

            if (typeMatch) return true;

            // If contentMediaType is specified but doesn't match, return false
            return false;
        }

        // If no media type constraint, but isBase64, accept generic file
        return isBase64;
    };

    const findCompatibleTools = (file: File, toolList: ToolDefinition[]): ToolMatch[] => {
        const results: ToolMatch[] = [];

        for (const tool of toolList) {
            if (!tool.inputSchema || !tool.inputSchema.properties) continue;

            // Simple scan of top-level properties
            for (const [key, schema] of Object.entries(tool.inputSchema.properties) as [string, any][]) {
                if (isCompatibleSchema(schema, file)) {
                    results.push({ tool, field: key });
                    break; // Take first compatible field per tool
                }
            }
        }
        return results;
    };

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);

        const file = e.dataTransfer.files?.[0];
        if (!file) return;

        const foundMatches = findCompatibleTools(file, tools);

        if (foundMatches.length === 0) {
            toast({
                variant: "destructive",
                title: "No Compatible Tools",
                description: `No tools found that accept ${file.type || 'this file type'}.`
            });
            return;
        }

        if (foundMatches.length === 1) {
            // Auto-select if only one match
            const match = foundMatches[0];
            onToolSelect(match.tool, file, match.field);
            toast({
                title: "Tool Selected",
                description: `Configuring ${match.tool.name} with dropped file.`
            });
        } else {
            setDroppedFile(file);
            setMatches(foundMatches);
            setIsOpen(true);
        }

    }, [tools, onToolSelect, toast]);

    const handleSelectMatch = (match: ToolMatch) => {
        if (droppedFile) {
            onToolSelect(match.tool, droppedFile, match.field);
            setIsOpen(false);
        }
    };

    return (
        <div
            className="relative h-full w-full"
            data-testid="smart-drop-zone"
            onDragEnter={handleDragEnter}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
        >
            {children}

            {isDragging && (
                <div className="absolute inset-0 z-50 bg-background/80 backdrop-blur-sm border-2 border-dashed border-primary flex flex-col items-center justify-center animate-in fade-in duration-200 pointer-events-none">
                     {/* pointer-events-none allows the drop to fall through to the container div */}
                    <div className="p-4 bg-muted rounded-full mb-4">
                        <Upload className="h-10 w-10 text-primary animate-bounce" />
                    </div>
                    <h3 className="text-2xl font-bold tracking-tight text-foreground">Drop file to configure tool</h3>
                    <p className="text-muted-foreground mt-2">
                        We will automatically detect compatible tools.
                    </p>
                </div>
            )}

            <Dialog open={isOpen} onOpenChange={setIsOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Select Tool</DialogTitle>
                        <DialogDescription>
                            Multiple tools can handle <strong>{droppedFile?.name}</strong>. Choose one to continue.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="grid gap-2 mt-4 max-h-[60vh] overflow-y-auto">
                        {matches.map((match) => (
                            <Button
                                key={`${match.tool.name}-${match.field}`}
                                variant="outline"
                                className="justify-start h-auto py-3 px-4"
                                onClick={() => handleSelectMatch(match)}
                            >
                                <div className="p-2 bg-primary/10 rounded mr-3">
                                    <Zap className="h-5 w-5 text-primary" />
                                </div>
                                <div className="text-left">
                                    <div className="font-semibold">{match.tool.name}</div>
                                    <div className="text-xs text-muted-foreground">
                                        Input field: <code className="bg-muted px-1 rounded">{match.field}</code>
                                    </div>
                                </div>
                            </Button>
                        ))}
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
