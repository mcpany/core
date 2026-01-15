/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import {
    FileText,
    Database,
    Image as ImageIcon,
    FileJson,
    Search,
    RefreshCw,
    Download,
    Copy,
    Eye,
    ChevronRight,
    File,
    Maximize2,
    Minimize2,
    LayoutGrid,
    List as ListIcon,
    Expand
} from "lucide-react";

import { apiClient, ResourceDefinition, ResourceContent } from "@/lib/client";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuSeparator,
    ContextMenuTrigger,
} from "@/components/ui/context-menu";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";
import { ResourceViewer } from "./resource-viewer";
import { ResourcePreviewModal } from "./resource-preview-modal";


interface ResourceExplorerProps {
    initialResources?: ResourceDefinition[];
}

export function ResourceExplorer({ initialResources = [] }: ResourceExplorerProps) {
    const [resources, setResources] = useState<ResourceDefinition[]>(initialResources);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [viewMode, setViewMode] = useState<"list" | "grid">("list");
    const [selectedUri, setSelectedUri] = useState<string | null>(null);
    const [resourceContent, setResourceContent] = useState<ResourceContent | null>(null);
    const [contentLoading, setContentLoading] = useState(false);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [previewResource, setPreviewResource] = useState<ResourceDefinition | null>(null);

    const { toast } = useToast();

    useEffect(() => {
        if (initialResources.length === 0) {
            loadResources();
        }
    }, []);

    useEffect(() => {
        if (selectedUri) {
            loadResourceContent(selectedUri);
        } else {
            setResourceContent(null);
        }
    }, [selectedUri]);

    const loadResources = async () => {
        setLoading(true);
        try {
            const res = await apiClient.listResources();
            setResources(res.resources || []);
        } catch (e) {
            console.error("Failed to load resources", e);
            toast({
                title: "Error",
                description: "Failed to load resources.",
                variant: "destructive"
            });
        } finally {
            setLoading(false);
        }
    };

    const loadResourceContent = async (uri: string) => {
        setContentLoading(true);
        try {
            const res = await apiClient.readResource(uri);
            if (res.contents && res.contents.length > 0) {
                setResourceContent(res.contents[0]);
            } else {
                setResourceContent(null);
            }
        } catch (e) {
            console.error("Failed to read resource", e);
            toast({
                title: "Error",
                description: "Failed to read resource content.",
                variant: "destructive"
            });
            setResourceContent(null);
        } finally {
            setContentLoading(false);
        }
    };

    const filteredResources = useMemo(() => {
        return resources.filter(r =>
            r.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            r.uri.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [resources, searchQuery]);

    const getIcon = (mimeType?: string) => {
        if (!mimeType) return File;
        if (mimeType.includes("json")) return FileJson;
        if (mimeType.includes("image")) return ImageIcon;
        if (mimeType.includes("text")) return FileText;
        if (mimeType.includes("sql") || mimeType.includes("database")) return Database;
        return File;
    };

    const handleCopyContent = () => {
        if (resourceContent?.text) {
            navigator.clipboard.writeText(resourceContent.text);
            toast({ title: "Copied", description: "Content copied to clipboard." });
        }
    };

    const handleDownload = (uri?: string) => {
        // If called from context menu with URI, we might not have content loaded yet.
        // For this demo, we only support download if content is already loaded (selected)
        // or we could fetch it. Ideally we should fetch.
        // For now, let's fallback to current selection if match.

        // If uri is provided and matches selectedUri, use resourceContent.
        // Otherwise we can't easily download without fetching first.
        const targetUri = uri || selectedUri;
        if (!targetUri) return;

        if (targetUri === selectedUri && resourceContent) {
             const blob = new Blob([resourceContent.text || ""], { type: resourceContent.mimeType });
             const url = URL.createObjectURL(blob);
             const a = document.createElement("a");
             a.href = url;
             const selectedRes = resources.find(r => r.uri === targetUri);
             a.download = selectedRes?.name || "resource";
             a.click();
        } else {
             // TODO: Fetch and download
             toast({ title: "Info", description: "Select the resource first to download." });
        }
    };

    const handleCopyUri = (uri: string) => {
        navigator.clipboard.writeText(uri);
        toast({ title: "Copied", description: "Resource URI copied to clipboard." });
    };

    const handleCopyName = (name: string) => {
        navigator.clipboard.writeText(name);
        toast({ title: "Copied", description: "Resource name copied to clipboard." });
    };

    return (
        <div className={cn("flex flex-col h-full bg-background", isFullscreen ? "fixed inset-0 z-50" : "rounded-lg border shadow-sm")}>
            {/* Header Toolbar */}
            <div className="flex items-center justify-between p-2 px-4 border-b bg-muted/20 h-14 shrink-0">
                <div className="flex items-center gap-2 flex-1 max-w-md">
                     <div className="relative w-full">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search resources..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="pl-8 h-9 text-xs"
                        />
                    </div>
                </div>

                <div className="flex items-center gap-2">
                    <div className="flex items-center bg-muted rounded-md p-1 gap-1">
                        <Button
                            variant={viewMode === "list" ? "secondary" : "ghost"}
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => setViewMode("list")}
                            title="List View"
                        >
                            <ListIcon className="h-4 w-4" />
                        </Button>
                         <Button
                            variant={viewMode === "grid" ? "secondary" : "ghost"}
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => setViewMode("grid")}
                            title="Grid View"
                        >
                            <LayoutGrid className="h-4 w-4" />
                        </Button>
                    </div>
                    <div className="h-4 w-px bg-border mx-1" />
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={loadResources} title="Refresh">
                        <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setIsFullscreen(!isFullscreen)} title="Fullscreen">
                        {isFullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
                    </Button>
                </div>
            </div>

            <ResizablePanelGroup direction="horizontal" className="flex-1">
                <ResizablePanel defaultSize={30} minSize={20} maxSize={50} className="flex flex-col bg-muted/5">
                    <ScrollArea className="flex-1">
                        {filteredResources.length === 0 ? (
                            <div className="p-8 text-center text-muted-foreground text-sm">
                                {loading ? "Loading..." : "No resources found."}
                            </div>
                        ) : viewMode === "list" ? (
                            <div className="divide-y">
                                {filteredResources.map(res => {
                                    const Icon = getIcon(res.mimeType);
                                    const isSelected = selectedUri === res.uri;
                                    return (
                                        <ContextMenu key={res.uri}>
                                            <ContextMenuTrigger asChild>
                                                <div
                                                    className={cn(
                                                        "flex items-center gap-3 p-3 px-4 cursor-pointer hover:bg-accent/50 transition-colors text-sm group",
                                                        isSelected ? "bg-accent text-accent-foreground border-l-4 border-l-primary pl-3" : "border-l-4 border-l-transparent"
                                                    )}
                                                    onClick={() => setSelectedUri(res.uri)}
                                                >
                                                    <Icon className={cn("h-4 w-4 text-muted-foreground group-hover:text-primary", isSelected && "text-primary")} />
                                                    <div className="flex-1 min-w-0">
                                                        <div className="font-medium truncate">{res.name}</div>
                                                        <div className="text-[10px] text-muted-foreground truncate opacity-70" title={res.uri}>{res.uri}</div>
                                                    </div>
                                                    {isSelected && <ChevronRight className="h-3 w-3 text-muted-foreground" />}
                                                </div>
                                            </ContextMenuTrigger>
                                            <ContextMenuContent>
                                                <ContextMenuItem onClick={() => setSelectedUri(res.uri)}>
                                                    <Eye className="mr-2 h-4 w-4" /> View Details
                                                </ContextMenuItem>
                                                <ContextMenuItem onClick={() => setPreviewResource(res)}>
                                                    <Expand className="mr-2 h-4 w-4" /> Preview in Modal
                                                </ContextMenuItem>
                                                <ContextMenuSeparator />
                                                <ContextMenuItem onClick={() => handleCopyUri(res.uri)}>
                                                    <Copy className="mr-2 h-4 w-4" /> Copy URI
                                                </ContextMenuItem>
                                                <ContextMenuItem onClick={() => handleCopyName(res.name)}>
                                                    <FileText className="mr-2 h-4 w-4" /> Copy Name
                                                </ContextMenuItem>
                                                <ContextMenuSeparator />
                                                <ContextMenuItem onClick={() => handleDownload(res.uri)} disabled={!isSelected}>
                                                    <Download className="mr-2 h-4 w-4" /> Download
                                                </ContextMenuItem>
                                            </ContextMenuContent>
                                        </ContextMenu>
                                    );
                                })}
                            </div>
                        ) : (
                            <div className="grid grid-cols-2 gap-2 p-3">
                                {filteredResources.map(res => {
                                    const Icon = getIcon(res.mimeType);
                                    const isSelected = selectedUri === res.uri;
                                    return (
                                        <ContextMenu key={res.uri}>
                                            <ContextMenuTrigger asChild>
                                                <Card
                                                    className={cn(
                                                        "cursor-pointer hover:border-primary/50 transition-all",
                                                        isSelected ? "border-primary ring-1 ring-primary" : ""
                                                    )}
                                                    onClick={() => setSelectedUri(res.uri)}
                                                >
                                                    <CardContent className="p-3 flex flex-col items-center text-center gap-2">
                                                        <div className="p-2 bg-muted rounded-full">
                                                            <Icon className="h-6 w-6 text-muted-foreground" />
                                                        </div>
                                                        <div className="w-full">
                                                            <div className="font-medium text-xs truncate" title={res.name}>{res.name}</div>
                                                            <div className="text-[10px] text-muted-foreground truncate mt-0.5">{res.mimeType || "unknown"}</div>
                                                        </div>
                                                    </CardContent>
                                                </Card>
                                            </ContextMenuTrigger>
                                            <ContextMenuContent>
                                                <ContextMenuItem onClick={() => setSelectedUri(res.uri)}>
                                                    <Eye className="mr-2 h-4 w-4" /> View Details
                                                </ContextMenuItem>
                                                <ContextMenuItem onClick={() => setPreviewResource(res)}>
                                                    <Expand className="mr-2 h-4 w-4" /> Preview in Modal
                                                </ContextMenuItem>
                                                <ContextMenuSeparator />
                                                <ContextMenuItem onClick={() => handleCopyUri(res.uri)}>
                                                    <Copy className="mr-2 h-4 w-4" /> Copy URI
                                                </ContextMenuItem>
                                                <ContextMenuItem onClick={() => handleCopyName(res.name)}>
                                                    <FileText className="mr-2 h-4 w-4" /> Copy Name
                                                </ContextMenuItem>
                                                <ContextMenuSeparator />
                                                <ContextMenuItem onClick={() => handleDownload(res.uri)} disabled={!isSelected}>
                                                    <Download className="mr-2 h-4 w-4" /> Download
                                                </ContextMenuItem>
                                            </ContextMenuContent>
                                        </ContextMenu>
                                    );
                                })}
                            </div>
                        )}
                    </ScrollArea>
                    <div className="p-2 border-t bg-muted/10 text-[10px] text-muted-foreground text-center">
                        {filteredResources.length} items
                    </div>
                </ResizablePanel>

                <ResizableHandle />

                <ResizablePanel defaultSize={70} className="bg-background flex flex-col min-w-0">
                    {selectedUri ? (
                        <>
                            <div className="flex items-center justify-between p-3 border-b bg-muted/5 h-12 shrink-0">
                                <div className="flex items-center gap-2 overflow-hidden">
                                     <div className="font-mono text-xs text-muted-foreground truncate max-w-md bg-muted px-2 py-1 rounded select-all">
                                        {selectedUri}
                                     </div>
                                     <Badge variant="outline" className="text-[10px] font-normal h-5">{resourceContent?.mimeType || "loading..."}</Badge>
                                </div>
                                <div className="flex items-center gap-1">
                                    <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={handleCopyContent} disabled={!resourceContent}>
                                        <Copy className="h-3 w-3 mr-1" /> Copy
                                    </Button>
                                    <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={() => handleDownload()} disabled={!resourceContent}>
                                        <Download className="h-3 w-3 mr-1" /> Download
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-7 w-7"
                                        onClick={() => {
                                            const res = resources.find(r => r.uri === selectedUri);
                                            if (res) setPreviewResource(res);
                                        }}
                                        title="Maximize"
                                    >
                                        <Expand className="h-3 w-3" />
                                    </Button>
                                </div>
                            </div>
                            <div className="flex-1 overflow-hidden relative">
                                <ResourceViewer content={resourceContent} loading={contentLoading} />
                            </div>
                        </>
                    ) : (
                         <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-4">
                            <div className="bg-muted/30 p-8 rounded-full">
                                <Search className="h-16 w-16 opacity-20" />
                            </div>
                            <div className="text-center">
                                <h3 className="text-lg font-medium">No Resource Selected</h3>
                                <p className="text-sm opacity-70">Select an item from the list to view its contents.</p>
                            </div>
                        </div>
                    )}
                </ResizablePanel>
            </ResizablePanelGroup>

            <ResourcePreviewModal
                isOpen={!!previewResource}
                onClose={() => setPreviewResource(null)}
                resource={previewResource}
                initialContent={previewResource?.uri === selectedUri ? resourceContent : undefined}
            />
        </div>
    );
}
