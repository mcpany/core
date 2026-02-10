/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { ChevronLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

/**
 * StackDetailPage component for viewing and editing a specific stack.
 * @param props The component props.
 * @returns The rendered component.
 */
export default function StackDetailPage(props: { params: Promise<{ stackId: string }> }) {
  // Handle Next.js 15 params as Promise without React.use (since we are on React 18)
  const [stackId, setStackId] = useState<string | null>(null);
  const router = useRouter();
  const { toast } = useToast();
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    props.params.then((p) => {
      if (mounted) {
        setStackId(p.stackId);
      }
    });
    return () => { mounted = false; };
  }, [props.params]);

  useEffect(() => {
    if (!stackId) return;

    const isNew = stackId === "new";

    if (isNew) {
      setContent(`# New Stack Configuration
name: new-stack
services:
  - name: example-service
    # Use snake_case for service types as per backend schema
    command_line_service:
      # Command must be a single executable in PATH. Arguments are not supported in this field directly?
      # Actually, server validation treats "command" as the binary.
      # For args, we might need a different service type or shell wrapper if supported.
      # For this test, we use a simple command that exists.
      command: ls
      working_directory: /tmp
`);
      setLoading(false);
      return;
    }

    async function loadStack() {
      try {
        const yaml = await apiClient.getStackYaml(stackId!);
        setContent(yaml);
      } catch (error) {
        console.error("Failed to load stack", error);
        toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load stack configuration."
        });
      } finally {
        setLoading(false);
      }
    }
    loadStack();
  }, [stackId, toast]);

  const handleSave = async (newContent: string) => {
    if (!stackId) return;
    try {
      // Simple regex to find name
      const nameMatch = newContent.match(/^name:\s*(.+)$/m);
      if (!nameMatch) {
          throw new Error("YAML must contain a 'name' field");
      }
      const name = nameMatch[1].trim();

      await apiClient.saveStackYaml(name, newContent);

      toast({
        title: "Stack Saved",
        description: `Stack ${name} has been deployed.`
      });

      if (stackId === "new" || name !== stackId) {
        router.push(`/stacks/${name}`);
      }
    } catch (error: any) {
      console.error("Failed to save stack", error);
      throw error; // Re-throw to show in editor error alert
    }
  };

  if (loading || !stackId) {
    return <div className="flex items-center justify-center h-full"><Loader2 className="animate-spin" /></div>;
  }

  const isNew = stackId === "new";

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-8 pt-6 space-y-4">
      <div className="flex items-center gap-4">
        <Link href="/stacks">
          <Button variant="ghost" size="icon">
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </Link>
        <h1 className="text-2xl font-bold tracking-tight">
          {isNew ? "New Stack" : stackId}
        </h1>
      </div>

      <div className="flex-1 min-h-0">
        <StackEditor
          initialContent={content}
          onSave={handleSave}
          onCancel={() => router.back()}
        />
      </div>
    </div>
  );
}
