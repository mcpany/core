/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * StackDetailPage component.
 * @returns The rendered component.
 */
export default function StackDetailPage() {
  const params = useParams();
  const stackId = params?.stackId as string;
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (stackId) {
      loadStack();
    }
  }, [stackId]);

  const loadStack = async () => {
    try {
      setLoading(true);
      const yaml = await apiClient.getStackYaml(stackId);
      setContent(yaml);
    } catch (e) {
      console.error("Failed to load stack", e);
      toast({
          title: "Error",
          description: "Failed to load stack configuration.",
          variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (newContent: string) => {
      await apiClient.saveStackYaml(stackId, newContent);
      toast({
          title: "Stack Saved",
          description: "Configuration updated successfully."
      });
      // Optionally reload to ensure sync
      loadStack();
  };

  if (loading) {
      return (
          <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  return (
    <div className="h-[calc(100vh-4rem)] p-6">
        <StackEditor
            initialContent={content}
            onSave={handleSave}
            onCancel={() => router.push("/stacks")}
        />
    </div>
  );
}
