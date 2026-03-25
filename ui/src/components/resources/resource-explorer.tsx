/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React, { useState, useEffect, useMemo, useCallback } from "react";
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
    Expand,
    ChevronLeft,
    SearchCode
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

const getIcon = (mimeType?: string) => {
    if (!mimeType) return File;
    if (mimeType.includes("json")) return FileJson;
    if (mimeType.includes("image")) return ImageIcon;
    if (mimeType.includes("text")) return FileText;
    if (mimeType.includes("sql") || mimeType.includes("database")) return Database;
    return File;
};

// ⚡ BOLT: [Render Optimization] Extract Resource List/Grid items into React.memo components.
// Randomized Selection from Top 5 High-Impact Targets (React/View - Render waste).
// This prevents re-rendering hundreds of unchanged resource rows every time selectedUri changes.
/**
 * Summary: MemoizedResourceListItem component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.res: The res property.
 *   - props.isSelected: The isSelected property.
 *   - props.onSelect: The onSelect property.
 *   - props.onDragStart: The onDragStart property.
 *   - props.onPreview: The onPreview property.
 *   - props.onCopyUri: The onCopyUri property.
 *   - props.onCopyName: The name of the onCopy.
 *   - props.onDownload: The onDownload property.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const MemoizedResourceListItem = React.memo(({ res, isSelected, onSelect, onDragStart, onPreview, onCopyUri, onCopyName, onDownload }: any) => {
/**
 * Summary: Icon component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
    const Icon = getIcon(res.mimeType);
    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <div
                    className={cn(
                        "flex items-center gap-3 p-3 px-4 cursor-pointer hover:bg-accent/50 transition-colors text-sm group",
                        isSelected ? "bg-accent text-accent-foreground border-l-4 border-l-primary pl-3" : "border-l-4 border-l-transparent"
                    )}
                    onClick={() => onSelect(res.uri)}
                    draggable
                    onDragStart={(e) => onDragStart(e, res)}
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
                <ContextMenuItem onClick={() => onSelect(res.uri)}>
                    <Eye className="mr-2 h-4 w-4" /> View Details
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onPreview(res)}>
                    <Expand className="mr-2 h-4 w-4" /> Preview in Modal
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => onCopyUri(res.uri)}>
                    <Copy className="mr-2 h-4 w-4" /> Copy URI
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onCopyName(res.name)}>
                    <FileText className="mr-2 h-4 w-4" /> Copy Name
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => onDownload(res.uri)} disabled={!isSelected}>
                    <Download className="mr-2 h-4 w-4" /> Download
                </ContextMenuItem>
            </ContextMenuContent>
        </ContextMenu>
    );
});
MemoizedResourceListItem.displayName = "MemoizedResourceListItem";

/**
 * Summary: MemoizedResourceGridItem component.
 *
 * Parameters:
 *   - props (Object): The component props.
 *   - props.res: The res property.
 *   - props.isSelected: The isSelected property.
 *   - props.onSelect: The onSelect property.
 *   - props.onPreview: The onPreview property.
 *   - props.onCopyUri: The onCopyUri property.
 *   - props.onCopyName: The name of the onCopy.
 *   - props.onDownload: The onDownload property.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
const MemoizedResourceGridItem = React.memo(({ res, isSelected, onSelect, onPreview, onCopyUri, onCopyName, onDownload }: any) => {
/**
 * Summary: Icon component.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.ReactNode: The rendered component.
 *
 * Throws/Errors:
 *   - None.
 */
    const Icon = getIcon(res.mimeType);
    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <Card
                    className={cn(
                        "cursor-pointer hover:border-primary/50 transition-all",
                        isSelected ? "border-primary ring-1 ring-primary" : ""
                    )}
                    onClick={() => onSelect(res.uri)}
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
                <ContextMenuItem onClick={() => onSelect(res.uri)}>
                    <Eye className="mr-2 h-4 w-4" /> View Details
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onPreview(res)}>
                    <Expand className="mr-2 h-4 w-4" /> Preview in Modal
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => onCopyUri(res.uri)}>
                    <Copy className="mr-2 h-4 w-4" /> Copy URI
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onCopyName(res.name)}>
                    <FileText className="mr-2 h-4 w-4" /> Copy Name
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={() => onDownload(res.uri)} disabled={!isSelected}>
                    <Download className="mr-2 h-4 w-4" /> Download
                </ContextMenuItem>
            </ContextMenuContent>
        </ContextMenu>
    );
});
MemoizedResourceGridItem.displayName = "MemoizedResourceGridItem";

/**
 * ResourceExplorer.
 *
 * @param { initialResources = [] - The { initialResources = [].
 */
export function ResourceExplorer({ initialResources = [] }: ResourceExplorerProps) {
    const [resources, setResources] = useState<ResourceDefinition[]>(initialResources);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [viewMode, setViewMode] = useState<"list" | "grid">("list");
    const [isDeepSearch, setIsDeepSearch] = useState(false);
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

    const loadResources = useCallback(async () => {
        setLoading(true);
        try {
            const res = await apiClient.listResources();
            if (!res) {
                setResources([]);
                return;
            }
            if (Array.isArray(res)) {
                setResources(res);
            } else if (res && Array.isArray(res.resources)) {
                setResources(res.resources);
            } else {
                setResources([]);
            }
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
    }, [toast]);

    const loadResourceContent = useCallback(async (uri: string) => {
        setContentLoading(true);
        try {
            const res = await apiClient.readResource(uri);
            if (res?.contents && res.contents.length > 0) {
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
    }, [toast]);

    const filteredResources = useMemo(() => {
        return resources.filter(r => {
            const matchesBasic = r.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                               r.uri.toLowerCase().includes(searchQuery.toLowerCase());

            if (isDeepSearch && searchQuery.length > 2) {
                // If we have content cached for this resource, search it too
                // Note: This is an optimistic client-side deep search.
                return matchesBasic || (r.uri === selectedUri && resourceContent?.text?.toLowerCase().includes(searchQuery.toLowerCase()));
            }
            return matchesBasic;
        });
    }, [resources, searchQuery, isDeepSearch, selectedUri, resourceContent]);

    const handleCopyContent = useCallback(() => {
        if (resourceContent?.text) {
            navigator.clipboard.writeText(resourceContent.text);
            toast({ title: "Copied", description: "Content copied to clipboard." });
        }
    }, [resourceContent, toast]);

    const handleDownload = useCallback(async (uri: string) => {
        if (!uri) return;

        const targetRes = resources.find(r => r.uri === uri);
        if (!targetRes) {
            toast({ title: "Error", description: "Resource definition not found." });
            return;
        }

        try {
            toast({ title: "Downloading...", description: "Fetching resource content." });
            const res = await apiClient.readResource(uri);
            if (!res.contents || res.contents.length === 0) {
                 toast({ title: "Error", description: "No content found for resource.", variant: "destructive" });
                 return;
            }

            const content = res.contents[0];
            let blob: Blob;

            if (content.blob) {
                // Decode base64 to blob
                const byteCharacters = atob(content.blob);
                const byteNumbers = new Array(byteCharacters.length);
                for (let i = 0; i < byteCharacters.length; i++) {
                    byteNumbers[i] = byteCharacters.charCodeAt(i);
                }
                const byteArray = new Uint8Array(byteNumbers);
                blob = new Blob([byteArray], { type: content.mimeType });
            } else {
                blob = new Blob([content.text || ""], { type: content.mimeType });
            }

            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = targetRes.name;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

        } catch (e) {
            console.error("Failed to download resource", e);
            toast({ title: "Error", description: "Failed to download resource.", variant: "destructive" });
        }
    }, [resources, toast]);

    const handleDownloadCurrent = useCallback(() => {
        if (selectedUri) {
            handleDownload(selectedUri);
        }
    }, [handleDownload, selectedUri]);

    const handleCopyUri = useCallback((uri: string) => {
        navigator.clipboard.writeText(uri);
        toast({ title: "Copied", description: "Resource URI copied to clipboard." });
    }, [toast]);

    const handleCopyName = useCallback((name: string) => {
        navigator.clipboard.writeText(name);
        toast({ title: "Copied", description: "Resource name copied to clipboard." });
    }, [toast]);

    const handleDragStart = useCallback((e: React.DragEvent, res: ResourceDefinition) => {
        // Sets the data to be dragged as the URI
        // This allows dragging to apps that accept text/uri-list
        e.dataTransfer.setData("text/plain", res.uri);
        e.dataTransfer.setData("text/uri-list", res.uri);

        // Add DownloadURL support for drag-and-drop to desktop
        const token = localStorage.getItem('mcp_auth_token');
        // Construct absolute URL
        const downloadUrl = `${window.location.origin}/api/v1/resources/download?uri=${encodeURIComponent(res.uri)}&name=${encodeURIComponent(res.name)}&token=${token || ''}`;
        // Format: mimeType:fileName:url
        const downloadData = `${res.mimeType || 'application/octet-stream'}:${res.name}:${downloadUrl}`;
        e.dataTransfer.setData("DownloadURL", downloadData);

        e.dataTransfer.effectAllowed = "copy";
    }, []);

    const navigateSibling = useCallback((direction: 'next' | 'prev') => {
        const currentIndex = filteredResources.findIndex(r => r.uri === selectedUri);
        if (currentIndex === -1) return;

        let nextIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1;
        if (nextIndex >= 0 && nextIndex < filteredResources.length) {
            setSelectedUri(filteredResources[nextIndex].uri);
        }
    }, [filteredResources, selectedUri]);

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
                    <Button
                        variant={isDeepSearch ? "secondary" : "ghost"}
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => setIsDeepSearch(!isDeepSearch)}
                        title="Search within content (cached only)"
                    >
                        <SearchCode className={cn("h-4 w-4", isDeepSearch && "text-primary")} />
                    </Button>
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
                                {filteredResources.map(res => (
                                    <MemoizedResourceListItem
                                        key={res.uri}
                                        res={res}
                                        isSelected={selectedUri === res.uri}
                                        onSelect={setSelectedUri}
                                        onDragStart={handleDragStart}
                                        onPreview={setPreviewResource}
                                        onCopyUri={handleCopyUri}
                                        onCopyName={handleCopyName}
                                        onDownload={handleDownload}
                                    />
                                ))}
                            </div>
                        ) : (
                            <div className="grid grid-cols-2 gap-2 p-3">
                                {filteredResources.map(res => (
                                    <MemoizedResourceGridItem
                                        key={res.uri}
                                        res={res}
                                        isSelected={selectedUri === res.uri}
                                        onSelect={setSelectedUri}
                                        onPreview={setPreviewResource}
                                        onCopyUri={handleCopyUri}
                                        onCopyName={handleCopyName}
                                        onDownload={handleDownload}
                                    />
                                ))}
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
                                    <div className="flex items-center gap-1 mr-2 px-1 bg-muted/50 rounded-md">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-7 w-7"
                                            onClick={() => navigateSibling('prev')}
                                            disabled={filteredResources.length <= 1 || filteredResources[0].uri === selectedUri}
                                            title="Previous"
                                        >
                                            <ChevronLeft className="h-4 w-4" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-7 w-7"
                                            onClick={() => navigateSibling('next')}
                                            disabled={filteredResources.length <= 1 || filteredResources[filteredResources.length - 1].uri === selectedUri}
                                            title="Next"
                                        >
                                            <ChevronRight className="h-4 w-4" />
                                        </Button>
                                    </div>
                                    <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={handleCopyContent} disabled={!resourceContent}>
                                        <Copy className="h-3 w-3 mr-1" /> Copy
                                    </Button>
                                    <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={handleDownloadCurrent} disabled={!selectedUri}>
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
