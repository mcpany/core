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
    Expand,
    ChevronLeft,
    SearchCode,
    Folder,
    Server,
    PanelLeftClose,
    PanelLeftOpen
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
import { buildResourceTree, TreeNode, flattenTree } from "@/lib/resource-tree";
import { ResourceTree } from "./resource-tree";
import { ResourceBreadcrumb } from "./resource-breadcrumb";


interface ResourceExplorerProps {
    initialResources?: ResourceDefinition[];
}

/**
 * ResourceExplorer.
 * Refactored to support 3-pane layout with Sidebar Tree.
 *
 * @param { initialResources = [] - The { initialResources = [].
 */
export function ResourceExplorer({ initialResources = [] }: ResourceExplorerProps) {
    const [resources, setResources] = useState<ResourceDefinition[]>(initialResources);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [viewMode, setViewMode] = useState<"list" | "grid">("list");
    const [isDeepSearch, setIsDeepSearch] = useState(false);

    // Navigation State
    const [currentFolder, setCurrentFolder] = useState<TreeNode | null>(null); // null means root
    const [selectedUri, setSelectedUri] = useState<string | null>(null); // For preview pane
    const [sidebarOpen, setSidebarOpen] = useState(true);

    const [resourceContent, setResourceContent] = useState<ResourceContent | null>(null);
    const [contentLoading, setContentLoading] = useState(false);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [previewResource, setPreviewResource] = useState<ResourceDefinition | null>(null);

    const { toast } = useToast();

    // Derived Tree
    const treeData = useMemo(() => buildResourceTree(resources), [resources]);

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
    };

    const loadResourceContent = async (uri: string) => {
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
    };

    // Calculate breadcrumb path
    const breadcrumbPath = useMemo(() => {
        const path: TreeNode[] = [];
        // Since we don't have parent pointers in TreeNode (simplicity),
        // we can reconstruct path if we know currentFolder.
        // Wait, currentFolder is a node reference. We just need to find it in tree?
        // Actually, for `file://` we split by path, so we can't easily walk up without parent links or full re-traverse.
        // Alternatively, we can store "path stack" in state instead of just currentFolder.
        // Or, since we only have `buildResourceTree` output, let's just use `fullPath` matching if unique?
        // Re-traversing from root to find path to currentFolder.id:

        if (!currentFolder) return [];

        const findPath = (nodes: TreeNode[], targetId: string): TreeNode[] | null => {
            for (const node of nodes) {
                if (node.id === targetId) return [node];
                if (node.children) {
                    const found = findPath(node.children, targetId);
                    if (found) return [node, ...found];
                }
            }
            return null;
        };

        return findPath(treeData, currentFolder.id) || [];
    }, [currentFolder, treeData]);


    // Determine visible items in Main Pane
    const visibleItems = useMemo(() => {
        // If searching, search EVERYTHING (flat)
        if (searchQuery) {
            const allNodes = flattenTree(treeData);
            return allNodes.filter(node => {
                const matchesName = node.name.toLowerCase().includes(searchQuery.toLowerCase());
                // Only show files in search results? Or folders too? Usually files are what users want.
                // Let's show matching files and folders.
                return matchesName;
            });
        }

        // Otherwise show children of current folder (or root)
        return currentFolder ? (currentFolder.children || []) : treeData;
    }, [currentFolder, treeData, searchQuery]);

    const getIcon = (node: TreeNode) => {
        if (node.type === "folder") {
            if (node.name.includes("://")) return Server;
            if (node.name === "db" || node.fullPath.includes("postgres")) return Database;
            return Folder;
        }

        const mimeType = node.resource?.mimeType;
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

    const handleDownload = async (uri?: string) => {
        const targetUri = uri || selectedUri;
        if (!targetUri) return;

        const targetRes = resources.find(r => r.uri === targetUri);
        if (!targetRes) {
            toast({ title: "Error", description: "Resource definition not found." });
            return;
        }

        try {
            toast({ title: "Downloading...", description: "Fetching resource content." });
            const res = await apiClient.readResource(targetUri);
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
    };

    const handleCopyUri = (uri: string) => {
        navigator.clipboard.writeText(uri);
        toast({ title: "Copied", description: "Resource URI copied to clipboard." });
    };

    const handleCopyName = (name: string) => {
        navigator.clipboard.writeText(name);
        toast({ title: "Copied", description: "Resource name copied to clipboard." });
    };

    const handleDragStart = (e: React.DragEvent, res: ResourceDefinition) => {
        e.dataTransfer.setData("text/plain", res.uri);
        e.dataTransfer.setData("text/uri-list", res.uri);
        const token = localStorage.getItem('mcp_auth_token');
        const downloadUrl = `${window.location.origin}/api/resources/download?uri=${encodeURIComponent(res.uri)}&name=${encodeURIComponent(res.name)}&token=${token || ''}`;
        const downloadData = `${res.mimeType || 'application/octet-stream'}:${res.name}:${downloadUrl}`;
        e.dataTransfer.setData("DownloadURL", downloadData);
        e.dataTransfer.effectAllowed = "copy";
    };

    const handleSidebarSelect = (node: TreeNode) => {
        if (node.type === "folder") {
            setCurrentFolder(node);
            setSearchQuery(""); // Clear search when navigating
        } else {
            // If file selected in sidebar, preview it
            setSelectedUri(node.fullPath);
        }
    };

    const handleMainItemClick = (node: TreeNode) => {
        if (node.type === "folder") {
            setCurrentFolder(node);
            setSearchQuery("");
        } else {
            setSelectedUri(node.fullPath);
        }
    };

    return (
        <div className={cn("flex flex-col h-full bg-background", isFullscreen ? "fixed inset-0 z-50" : "rounded-lg border shadow-sm")}>
            {/* Header Toolbar */}
            <div className="flex items-center justify-between p-2 px-4 border-b bg-muted/20 h-14 shrink-0">
                 <div className="flex items-center gap-2">
                    <Button variant="ghost" size="icon" onClick={() => setSidebarOpen(!sidebarOpen)}>
                        {sidebarOpen ? <PanelLeftClose className="h-4 w-4" /> : <PanelLeftOpen className="h-4 w-4" />}
                    </Button>
                    <div className="relative w-64 md:w-80">
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
                {/* Sidebar Pane */}
                {sidebarOpen && (
                    <>
                    <ResizablePanel defaultSize={20} minSize={15} maxSize={30} className="flex flex-col bg-muted/5 border-r">
                         <div className="p-2 text-xs font-semibold text-muted-foreground border-b uppercase tracking-wider">
                             Explorer
                         </div>
                         <ScrollArea className="flex-1 p-2">
                             <ResourceTree
                                data={treeData}
                                onSelect={handleSidebarSelect}
                                selectedId={currentFolder?.id || selectedUri || undefined}
                             />
                         </ScrollArea>
                    </ResizablePanel>
                    <ResizableHandle />
                    </>
                )}

                {/* Main Content Pane */}
                <ResizablePanel defaultSize={40} minSize={30}>
                    <div className="flex flex-col h-full bg-background">
                         {/* Breadcrumb Bar */}
                         <div className="px-4 py-2 border-b flex items-center bg-background shrink-0 h-10">
                            <ResourceBreadcrumb
                                path={breadcrumbPath}
                                onNavigate={(node) => {
                                    setCurrentFolder(node);
                                    setSearchQuery("");
                                }}
                            />
                         </div>

                        <ScrollArea className="flex-1">
                            {visibleItems.length === 0 ? (
                                <div className="p-8 text-center text-muted-foreground text-sm">
                                    {loading ? "Loading..." : "No items found."}
                                </div>
                            ) : viewMode === "list" ? (
                                <div className="divide-y">
                                    {visibleItems.map(node => {
                                        const Icon = getIcon(node);
                                        const isSelected = selectedUri === node.fullPath;
                                        return (
                                            <div
                                                key={node.id}
                                                className={cn(
                                                    "flex items-center gap-3 p-3 px-4 cursor-pointer hover:bg-accent/50 transition-colors text-sm group",
                                                    isSelected && node.type === 'file' ? "bg-accent text-accent-foreground border-l-4 border-l-primary pl-3" : "border-l-4 border-l-transparent"
                                                )}
                                                onClick={() => handleMainItemClick(node)}
                                                onDoubleClick={() => node.type === 'folder' && handleMainItemClick(node)}
                                            >
                                                <Icon className={cn("h-4 w-4 text-muted-foreground group-hover:text-primary", isSelected && node.type === 'file' && "text-primary")} />
                                                <div className="flex-1 min-w-0">
                                                    <div className="font-medium truncate">{node.name}</div>
                                                    {node.type === 'file' && (
                                                        <div className="text-[10px] text-muted-foreground truncate opacity-70" title={node.fullPath}>{node.fullPath}</div>
                                                    )}
                                                </div>
                                                {node.type === 'folder' && <ChevronRight className="h-3 w-3 text-muted-foreground" />}
                                            </div>
                                        );
                                    })}
                                </div>
                            ) : (
                                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 p-3">
                                    {visibleItems.map(node => {
                                        const Icon = getIcon(node);
                                        const isSelected = selectedUri === node.fullPath;
                                        return (
                                            <Card
                                                key={node.id}
                                                className={cn(
                                                    "cursor-pointer hover:border-primary/50 transition-all shadow-sm",
                                                    isSelected && node.type === 'file' ? "border-primary ring-1 ring-primary" : ""
                                                )}
                                                onClick={() => handleMainItemClick(node)}
                                                onDoubleClick={() => node.type === 'folder' && handleMainItemClick(node)}
                                            >
                                                <CardContent className="p-3 flex flex-col items-center text-center gap-2">
                                                    <div className="p-2 bg-muted rounded-full">
                                                        <Icon className={cn("h-6 w-6", node.type === 'folder' ? "text-blue-500/70" : "text-muted-foreground")} />
                                                    </div>
                                                    <div className="w-full">
                                                        <div className="font-medium text-xs truncate" title={node.name}>{node.name}</div>
                                                        <div className="text-[10px] text-muted-foreground truncate mt-0.5">
                                                            {node.type === 'folder' ? "Folder" : (node.resource?.mimeType || "File")}
                                                        </div>
                                                    </div>
                                                </CardContent>
                                            </Card>
                                        );
                                    })}
                                </div>
                            )}
                        </ScrollArea>
                        <div className="p-2 border-t bg-muted/10 text-[10px] text-muted-foreground text-center">
                            {visibleItems.length} items
                        </div>
                    </div>
                </ResizablePanel>

                <ResizableHandle />

                {/* Preview Pane */}
                <ResizablePanel defaultSize={40} className="bg-background flex flex-col min-w-0">
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
                                    <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={() => handleDownload()} disabled={!selectedUri}>
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
                                <p className="text-sm opacity-70">Select a file to view its contents.</p>
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
