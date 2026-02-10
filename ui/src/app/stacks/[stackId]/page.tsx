/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2 } from "lucide-react";
import * as yaml from "js-yaml";

const DEFAULT_TEMPLATE = `name: my-new-stack
version: 1.0.0
services:
  - name: example-service
    commandLineService:
      command: "echo 'Hello World'"
`;

/**
 * StackDetailPage component.
 * Displays the Stack Editor for creating or editing a stack.
 * @returns The rendered page.
 */
export default function StackDetailPage() {
  const params = useParams();
  const stackId = Array.isArray(params?.stackId) ? params.stackId[0] : params?.stackId;

  const router = useRouter();
  const { toast } = useToast();
  const [content, setContent] = useState<string>("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!stackId) return;

    if (stackId === "new") {
      setContent(DEFAULT_TEMPLATE);
      setLoading(false);
      return;
    }

    // Fetch existing
    apiClient.getStackYaml(stackId)
      .then(setContent)
      .catch((e) => {
        console.error(e);
        toast({ title: "Error", description: "Failed to load stack configuration", variant: "destructive" });
      })
      .finally(() => setLoading(false));
  }, [stackId, toast]);

  const handleSave = async (newContent: string) => {
    if (!stackId) return;

    try {
      // If new, validate and get name
      let targetId = stackId;
      if (stackId === "new") {
         console.log("Parsing YAML...");
         let parsed: any;
         try {
             // Handle both default export and named export nuances
             const loadFn = (yaml as any).load || yaml.load;
             if (!loadFn) {
                 throw new Error("YAML parser not found");
             }
             parsed = loadFn(newContent);
         } catch (e: any) {
             throw new Error(`Invalid YAML: ${e.message}`);
         }

         if (!parsed || !parsed.name) {
             throw new Error("Stack name is required in YAML");
         }
         targetId = parsed.name;
         console.log("Target ID:", targetId);
      }

      await apiClient.saveStackYaml(targetId, newContent);
      toast({ title: "Stack Saved", description: "Configuration applied successfully." });

      if (stackId === "new") {
          console.log("Redirecting to", `/stacks/${targetId}`);
          router.push(`/stacks/${targetId}`);
      }
    } catch (e: any) {
        console.error("Save error:", e);
        toast({ title: "Save Failed", description: e.message, variant: "destructive" });
    }
  };

  if (loading) {
      return <div className="h-full flex items-center justify-center"><Loader2 className="animate-spin h-8 w-8" /></div>;
  }

  return (
    <div className="h-[calc(100vh-4rem)] p-4 md:p-8 pt-6 flex flex-col">
        <h1 className="text-2xl font-bold tracking-tight mb-4">
            {stackId === "new" ? "New Stack" : `Edit Stack: ${stackId}`}
        </h1>
        <div className="flex-1 min-h-0 border rounded-lg shadow-sm">
            <StackEditor
                initialValue={content}
                onSave={handleSave}
                onCancel={() => router.push("/stacks")}
            />
        </div>
    </div>
  );
}
