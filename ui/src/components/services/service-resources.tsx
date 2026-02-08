/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig, ResourceDefinition, apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { FileText, Eye } from "lucide-react";
import { useState, useEffect } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

interface ServiceResourcesProps {
    service: UpstreamServiceConfig;
}

export function ServiceResources({ service }: ServiceResourcesProps) {
    const [resources, setResources] = useState<ResourceDefinition[]>([]);
    const [selectedResource, setSelectedResource] = useState<ResourceDefinition | null>(null);
    const [resourceContent, setResourceContent] = useState<string | null>(null);
    const [loadingContent, setLoadingContent] = useState(false);

    useEffect(() => {
        // Extract resources from config
        let extracted: ResourceDefinition[] = [];
        if (service.httpService?.resources) extracted = service.httpService.resources;
        else if (service.grpcService?.resources) extracted = service.grpcService.resources;
        else if (service.commandLineService?.resources) extracted = service.commandLineService.resources;
        else if (service.mcpService?.resources) extracted = service.mcpService.resources;
        else if (service.openapiService?.resources) extracted = service.openapiService.resources;

        setResources(extracted);
    }, [service]);

    const handleRead = async (res: ResourceDefinition) => {
        setSelectedResource(res);
        setLoadingContent(true);
        setResourceContent(null);
        try {
            const response = await apiClient.readResource(res.uri);
            if (response.contents && response.contents.length > 0) {
                setResourceContent(response.contents[0].text || "(Binary Content)");
            } else {
                setResourceContent("(Empty)");
            }
        } catch (e: any) {
            setResourceContent(`Failed to read resource: ${e.message}`);
        } finally {
            setLoadingContent(false);
        }
    };

    if (resources.length === 0) {
        return <div className="text-muted-foreground text-sm">No resources found for this service.</div>;
    }

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {resources.map((res) => (
                <Card key={res.uri} className="flex flex-col">
                    <CardHeader className="pb-3">
                        <div className="flex items-start justify-between">
                             <div className="flex items-center gap-2">
                                <FileText className="h-4 w-4 text-blue-500" />
                                <CardTitle className="text-sm font-bold truncate" title={res.name}>
                                    {res.name}
                                </CardTitle>
                             </div>
                             <Button size="sm" variant="ghost" className="h-6 w-6 p-0" onClick={() => handleRead(res)}>
                                <Eye className="h-4 w-4" />
                            </Button>
                        </div>
                        <CardDescription className="text-xs truncate" title={res.uri}>
                            {res.uri}
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="text-xs text-muted-foreground">
                            {res.mimeType || "application/octet-stream"}
                        </div>
                    </CardContent>
                </Card>
            ))}

            <Dialog open={!!selectedResource} onOpenChange={(open) => !open && setSelectedResource(null)}>
                <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle>{selectedResource?.name}</DialogTitle>
                        <DialogDescription className="truncate">{selectedResource?.uri}</DialogDescription>
                    </DialogHeader>
                    <div className="flex-1 overflow-auto bg-muted/50 p-4 rounded-md font-mono text-xs whitespace-pre-wrap border">
                        {loadingContent ? "Reading..." : resourceContent}
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
