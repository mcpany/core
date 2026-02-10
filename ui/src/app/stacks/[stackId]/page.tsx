/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, use } from "react";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

/**
 * StackDetailPage component.
 * @param props - The component props.
 * @param props.params - The route parameters.
 * @returns The rendered component.
 */
export default function StackDetailPage(props: { params: Promise<{ stackId: string }> }) {
  const params = use(props.params);
  const { stackId } = params;
  const isNew = stackId === "new";

  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(!isNew);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isNew) return;

    async function load() {
      try {
        const yaml = await apiClient.getStackYaml(stackId);
        setContent(yaml);
      } catch (e: any) {
        console.error("Failed to load stack", e);
        setError(e.message || "Failed to load stack configuration");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [stackId, isNew]);

  if (loading) {
      return (
          <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  if (error) {
      return (
          <div className="flex items-center justify-center h-full">
              <div className="text-center">
                  <h3 className="text-lg font-medium text-destructive">Error</h3>
                  <p className="text-muted-foreground">{error}</p>
              </div>
          </div>
      );
  }

  return (
    <div className="h-[calc(100vh-4rem)] p-4 md:p-8 pt-6">
        <StackEditor
            initialContent={content}
            stackId={stackId}
            onSave={apiClient.saveStackYaml}
            isNew={isNew}
        />
    </div>
  );
}
