"use client";

import { ResourceDefinition, apiClient } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Eye, Loader2 } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";


interface ServiceResourcesProps {
    resources: ResourceDefinition[];
}

/**
 * ServiceResources lists the resources exposed by an upstream service.
 * It allows users to view resource details and read their content.
 */
export function ServiceResources({ resources }: ServiceResourcesProps) {
    const [selectedResource, setSelectedResource] = useState<ResourceDefinition | null>(null);
    const [content, setContent] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleRead = async (resource: ResourceDefinition) => {
        setSelectedResource(resource);
        setLoading(true);
        setContent(null);
        try {
            const res = await apiClient.readResource(resource.uri);
            if (res.contents && res.contents.length > 0) {
                // Prefer text, then blob (if blob, maybe show placeholder or try to decode?)
                const item = res.contents[0];
                if (item.text) {
                    setContent(item.text);
                } else if (item.blob) {
                     setContent(`[Binary Content] Base64: ${item.blob.substring(0, 50)}...`);
                } else {
                    setContent("[Empty Resource]");
                }
            } else {
                 setContent("[No Content Returned]");
            }
        } catch (e) {
            console.error("Failed to read resource", e);
            toast({
                title: "Failed to read resource",
                description: String(e),
                variant: "destructive"
            });
            setContent(null);
        } finally {
            setLoading(false);
        }
    };

    if (!resources || resources.length === 0) {
        return <div className="text-center py-10 text-muted-foreground">No resources discovered for this service.</div>;
    }

    return (
        <>
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Name</TableHead>
                            <TableHead>URI</TableHead>
                            <TableHead>MIME Type</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {resources.map((resource) => (
                            <TableRow key={resource.uri}>
                                <TableCell className="font-medium">{resource.name}</TableCell>
                                <TableCell className="font-mono text-xs">{resource.uri}</TableCell>
                                <TableCell>{resource.mimeType}</TableCell>
                                <TableCell className="text-right">
                                    <Button variant="ghost" size="sm" onClick={() => handleRead(resource)}>
                                        <Eye className="mr-2 h-4 w-4" /> Read
                                    </Button>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>

            <Dialog open={!!selectedResource} onOpenChange={(open) => !open && setSelectedResource(null)}>
                <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle>{selectedResource?.name}</DialogTitle>
                        <DialogDescription className="font-mono text-xs break-all">
                            {selectedResource?.uri}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex-1 overflow-auto rounded-md border bg-muted/20 p-4 relative min-h-[200px]">
                        {loading ? (
                            <div className="absolute inset-0 flex items-center justify-center">
                                <Loader2 className="h-8 w-8 animate-spin" />
                            </div>
                        ) : (
                            <div className="text-sm font-mono whitespace-pre-wrap break-all">
                                {content}
                            </div>
                        )}
                    </div>
                </DialogContent>
            </Dialog>
        </>
    );
}
