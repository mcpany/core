/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ServiceCollection } from "@/lib/marketplace-service";
import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Terminal, Globe, Play } from "lucide-react";

interface CollectionDetailsDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    collection: ServiceCollection | undefined;
    onInstantiateService: (service: UpstreamServiceConfig) => void;
}

/**
 * CollectionDetailsDialog component.
 * Displays details about a specific service collection in a dialog.
 *
 * @param props - The component props.
 * @param props.open - Whether the dialog is open.
 * @param props.onOpenChange - Callback function when the dialog open state changes.
 * @param props.collection - The service collection to display details for.
 * @param props.onInstantiateService - Callback function to instantiate a service from the collection.
 * @returns The rendered CollectionDetailsDialog component.
 */
export function CollectionDetailsDialog({
    open,
    onOpenChange,
    collection,
    onInstantiateService
}: CollectionDetailsDialogProps) {
    if (!collection) return null;

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
                <DialogHeader>
                    <DialogTitle className="text-2xl">{collection.name}</DialogTitle>
                    <DialogDescription>
                        {collection.description}
                    </DialogDescription>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mt-2">
                        <span>By {collection.author}</span>
                        <span>•</span>
                        <span>v{collection.version}</span>
                    </div>
                </DialogHeader>

                <div className="flex-1 overflow-y-auto py-4 pr-2">
                    <h3 className="font-semibold mb-3">Included Services ({collection.services.length})</h3>
                    <div className="grid gap-4">
                        {collection.services.map((service, idx) => (
                            <Card key={idx} className="bg-muted/30">
                                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                    <div className="flex items-center gap-2">
                                        {service.commandLineService ? (
                                            <Terminal className="h-4 w-4 text-muted-foreground" />
                                        ) : (
                                            <Globe className="h-4 w-4 text-muted-foreground" />
                                        )}
                                        <CardTitle className="text-base font-medium">
                                            {service.name || service.sanitizedName}
                                        </CardTitle>
                                    </div>
                                    <Button size="sm" onClick={() => onInstantiateService(service)}>
                                        <Play className="mr-2 h-3 w-3" />
                                        Instantiate
                                    </Button>
                                </CardHeader>
                                <CardContent>
                                    <CardDescription>
                                        {service.commandLineService
                                            ? `Command: ${service.commandLineService.command}`
                                            : service.httpService
                                                ? `URL: ${service.httpService.url}`
                                                : "Configuration Template"}
                                    </CardDescription>
                                    {service.tools && service.tools.length > 0 && (
                                        <div className="mt-2 text-xs text-muted-foreground">
                                            Tools: {service.tools.map(t => t.name).join(", ")}
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        ))}
                        {collection.services.length === 0 && (
                            <div className="text-center p-4 text-muted-foreground text-sm border border-dashed rounded-md">
                                This collection has no services listed.
                            </div>
                        )}
                    </div>
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        Close
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
