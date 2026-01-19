/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import * as React from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Copy, Download, RefreshCw, X } from "lucide-react";
import { apiClient, ResourceContent, ResourceDefinition } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { ResourceViewer } from "./resource-viewer";
import { Badge } from "@/components/ui/badge";

interface ResourcePreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  resource: ResourceDefinition | null;
  initialContent?: ResourceContent | null;
}

/**
 * ResourcePreviewModal component.
 * @param props - The component props.
 * @param props.isOpen - Whether the component is open.
 * @param props.onClose - The onClose property.
 * @param props.resource - The resource property.
 * @param props.initialContent - The initialContent property.
 * @returns The rendered component.
 */
export function ResourcePreviewModal({
  isOpen,
  onClose,
  resource,
  initialContent,
}: ResourcePreviewModalProps) {
  const [content, setContent] = React.useState<ResourceContent | null>(null);
  const [loading, setLoading] = React.useState(false);
  const { toast } = useToast();

  React.useEffect(() => {
    if (isOpen && resource) {
      if (initialContent && initialContent.uri === resource.uri) {
        setContent(initialContent);
      } else {
        fetchContent(resource.uri);
      }
    } else {
      setContent(null);
    }
  }, [isOpen, resource, initialContent]);

  const fetchContent = async (uri: string) => {
    setLoading(true);
    try {
      const res = await apiClient.readResource(uri);
      if (res.contents && res.contents.length > 0) {
        setContent(res.contents[0]);
      } else {
        setContent(null);
      }
    } catch (e) {
      console.error("Failed to read resource", e);
      toast({
        title: "Error",
        description: "Failed to read resource content.",
        variant: "destructive",
      });
      setContent(null);
    } finally {
      setLoading(false);
    }
  };

  const handleCopyContent = () => {
    if (content?.text) {
      navigator.clipboard.writeText(content.text);
      toast({ title: "Copied", description: "Content copied to clipboard." });
    }
  };

  const handleDownload = () => {
    if (content && resource) {
      let blob: Blob;
      if (content.blob) {
        // Convert base64 to blob
        try {
          const byteCharacters = atob(content.blob);
          const byteNumbers = new Array(byteCharacters.length);
          for (let i = 0; i < byteCharacters.length; i++) {
            byteNumbers[i] = byteCharacters.charCodeAt(i);
          }
          const byteArray = new Uint8Array(byteNumbers);
          blob = new Blob([byteArray], { type: content.mimeType });
        } catch (e) {
          console.error("Failed to convert base64 to blob", e);
          // Fallback to text if possible or empty
          blob = new Blob([content.text || ""], { type: content.mimeType });
        }
      } else {
        blob = new Blob([content.text || ""], { type: content.mimeType });
      }

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = resource.name || "resource";
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  if (!resource) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-4xl h-[85vh] flex flex-col p-0 gap-0">
        <DialogHeader className="p-4 border-b flex flex-row items-center justify-between space-y-0">
            <div className="flex flex-col gap-1 overflow-hidden">
                <div className="flex items-center gap-3">
                    <DialogTitle className="truncate" title={resource.name}>{resource.name}</DialogTitle>
                    <Badge variant="outline" className="text-xs font-normal whitespace-nowrap">
                        {content?.mimeType || resource.mimeType}
                    </Badge>
                </div>
                <DialogDescription className="sr-only">
                    Preview of resource {resource.name}
                </DialogDescription>
            </div>

            <div className="flex items-center gap-1 pl-2">
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => fetchContent(resource.uri)} title="Refresh">
                    <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                </Button>
                <div className="h-4 w-px bg-border mx-1" />
                 <Button variant="ghost" size="sm" className="h-8 px-2 text-xs" onClick={handleCopyContent} disabled={!content?.text}>
                    <Copy className="h-3 w-3 mr-1" /> Copy
                </Button>
                <Button variant="ghost" size="sm" className="h-8 px-2 text-xs" onClick={handleDownload} disabled={!content}>
                    <Download className="h-3 w-3 mr-1" /> Download
                </Button>
                <div className="h-4 w-px bg-border mx-1" />
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose} title="Close">
                    <X className="h-4 w-4" />
                </Button>
            </div>
        </DialogHeader>

        <div className="flex-1 overflow-hidden relative bg-muted/5">
             <ResourceViewer content={content} loading={loading} />
        </div>

        <div className="p-2 border-t bg-muted/20 text-xs text-muted-foreground px-4 flex items-center gap-2">
            <span className="font-mono bg-muted px-1.5 py-0.5 rounded select-all truncate max-w-2xl">{resource.uri}</span>
        </div>
      </DialogContent>
    </Dialog>
  );
}
