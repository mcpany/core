/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ResourceDefinition } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { FileText, Eye, Download, File, FileCode, FileImage } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface ServiceResourcesProps {
    resources: ResourceDefinition[];
}

/**
 * ServiceResources component.
 * Displays a list of resources provided by the service.
 * @param props - The component props.
 * @param props.resources - The list of resources.
 * @returns The rendered component.
 */
export function ServiceResources({ resources }: ServiceResourcesProps) {
    const [selectedResource, setSelectedResource] = useState<ResourceDefinition | null>(null);
    const [resourceContent, setResourceContent] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleRead = async (resource: ResourceDefinition) => {
        setSelectedResource(resource);
        setLoading(true);
        setResourceContent(null);
        try {
            const res = await apiClient.readResource(resource.uri);
            if (res.contents && res.contents.length > 0) {
                const content = res.contents[0];
                setResourceContent(content.text || (content.blob ? "[Binary Data]" : "[Empty]"));
            } else {
                setResourceContent("[No content returned]");
            }
        } catch (e) {
            console.error("Failed to read resource", e);
            toast({
                title: "Error",
                description: "Failed to read resource content.",
                variant: "destructive"
            });
            setResourceContent("Error loading content.");
        } finally {
            setLoading(false);
        }
    };

    const getIcon = (mimeType: string) => {
        if (mimeType.startsWith("image/")) return <FileImage className="h-4 w-4" />;
        if (mimeType.includes("json") || mimeType.includes("xml") || mimeType.includes("javascript")) return <FileCode className="h-4 w-4" />;
        if (mimeType.startsWith("text/")) return <FileText className="h-4 w-4" />;
        return <File className="h-4 w-4" />;
    };

    if (!resources || resources.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground bg-muted/20 rounded-lg border-2 border-dashed">
                <FileText className="h-10 w-10 mb-4 opacity-50" />
                <h3 className="text-lg font-medium">No resources found</h3>
                <p className="text-sm">This service does not expose any resources.</p>
            </div>
        );
    }

    return (
        <>
            <Card>
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[30px]"></TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>URI</TableHead>
                            <TableHead>MIME Type</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {resources.map((resource) => (
                            <TableRow key={resource.uri}>
                                <TableCell>
                                    {getIcon(resource.mimeType || "application/octet-stream")}
                                </TableCell>
                                <TableCell className="font-medium">
                                    {resource.name || "Untitled Resource"}
                                    {resource.description && (
                                        <p className="text-xs text-muted-foreground line-clamp-1">{resource.description}</p>
                                    )}
                                </TableCell>
                                <TableCell className="font-mono text-xs text-muted-foreground max-w-[200px] truncate" title={resource.uri}>
                                    {resource.uri}
                                </TableCell>
                                <TableCell>
                                    <Badge variant="secondary" className="font-mono text-[10px]">
                                        {resource.mimeType || "unknown"}
                                    </Badge>
                                </TableCell>
                                <TableCell className="text-right">
                                    <Button variant="ghost" size="sm" onClick={() => handleRead(resource)}>
                                        <Eye className="h-4 w-4 mr-2" /> Read
                                    </Button>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </Card>

            <Dialog open={!!selectedResource} onOpenChange={(open) => !open && setSelectedResource(null)}>
                <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            {selectedResource?.name || "Resource Content"}
                        </DialogTitle>
                        <DialogDescription className="font-mono text-xs break-all">
                            {selectedResource?.uri}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex-1 overflow-auto bg-muted/50 p-4 rounded-md border mt-2">
                         {loading ? (
                             <div className="flex items-center justify-center h-full text-muted-foreground">
                                 Loading...
                             </div>
                         ) : (
                             <pre className="text-xs font-mono whitespace-pre-wrap break-all">
                                 {resourceContent}
                             </pre>
                         )}
                    </div>
                </DialogContent>
            </Dialog>
        </>
    );
}
