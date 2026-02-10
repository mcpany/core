/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/client";
import { StackEditor } from "@/components/stacks/stack-editor";
import { useToast } from "@/hooks/use-toast";

interface StackDetailPageProps {
  params: Promise<{ stackId: string }>;
}

/**
 * StackDetailPage component.
 * @returns The rendered component.
 */
export default function StackDetailPage({ params }: StackDetailPageProps) {
  const [stackId, setStackId] = useState<string | null>(null);
  const router = useRouter();
  const { toast } = useToast();
  const [yaml, setYaml] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    params.then(p => setStackId(p.stackId));
  }, [params]);

  useEffect(() => {
    if (!stackId) return;

    async function load() {
      if (stackId === "new") {
        setLoading(false);
        return;
      }

      try {
        const config = await apiClient.getStackYaml(stackId!);
        setYaml(config);
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
    }
    load();
  }, [stackId, toast]);

  const handleSave = async (content: string) => {
    if (!stackId) return;

    try {
      await apiClient.saveStackYaml(stackId === "new" ? "new" : stackId, content);
      toast({
          title: "Success",
          description: "Stack configuration saved.",
      });
      if (stackId === "new") {
          router.push("/stacks");
      }
    } catch (e) {
      console.error("Failed to save stack", e);
      toast({
          title: "Error",
          description: String(e),
          variant: "destructive"
      });
      throw e;
    }
  };

  const handleCancel = () => {
    router.push("/stacks");
  };

  if (!stackId || loading) {
      return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-8 pt-6 space-y-4">
        <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold tracking-tight">
                {stackId === "new" ? "New Stack" : `Edit Stack: ${stackId}`}
            </h1>
        </div>
        <div className="flex-1 min-h-0 border rounded-lg bg-card text-card-foreground shadow-sm p-0 overflow-hidden">
             <StackEditor
                key={stackId || "loading"}
                initialYaml={yaml}
                onSave={handleSave}
                onCancel={handleCancel}
                loading={loading}
             />
        </div>
    </div>
  );
}
