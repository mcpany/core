/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Copy, Download } from "lucide-react";
import { ResourceDefinition, ResourceContent, apiClient } from "@/lib/client";
import { ResourceViewer } from "./resource-viewer";
import { useToast } from "@/hooks/use-toast";

interface ResourcePreviewModalProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  resource: ResourceDefinition | null;
  initialContent?: ResourceContent | null;
}

export function ResourcePreviewModal({
  isOpen,
  onOpenChange,
  resource,
  initialContent,
}: ResourcePreviewModalProps) {
  const [content, setContent] = useState<ResourceContent | null>(initialContent || null);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (isOpen && resource) {
      if (initialContent && initialContent.uri === resource.uri) {
        setContent(initialContent);
      } else {
        loadContent(resource.uri);
      }
    } else if (!isOpen) {
        setContent(null);
    }
  }, [isOpen, resource, initialContent]);

  const loadContent = async (uri: string) => {
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
    if (content && content.text) {
         const blob = new Blob([content.text], { type: content.mimeType });
         const url = URL.createObjectURL(blob);
         const a = document.createElement("a");
         a.href = url;
         a.download = resource?.name || "resource";
         a.click();
         URL.revokeObjectURL(url);
    }
  };

  if (!resource) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col p-0 gap-0">
        <DialogHeader className="p-4 border-b shrink-0">
          <div className="flex items-center justify-between pr-8">
            <div className="flex flex-col gap-1 min-w-0">
                <DialogTitle className="flex items-center gap-2 truncate">
                    {resource.name}
                </DialogTitle>
                <DialogDescription className="font-mono text-xs text-muted-foreground truncate">
                    {resource.uri}
                </DialogDescription>
            </div>
            <div className="flex items-center gap-1 shrink-0">
               <Button variant="ghost" size="sm" onClick={handleCopyContent} disabled={!content}>
                    <Copy className="h-4 w-4 mr-2" /> Copy
               </Button>
               <Button variant="ghost" size="sm" onClick={handleDownload} disabled={!content}>
                    <Download className="h-4 w-4 mr-2" /> Download
               </Button>
            </div>
          </div>
        </DialogHeader>
        <div className="flex-1 overflow-hidden p-0 relative">
             <ResourceViewer content={content} loading={loading} />
        </div>
      </DialogContent>
    </Dialog>
  );
}
