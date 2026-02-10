/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Trash2 } from "lucide-react";

/**
 * StackDetailPage component.
 * @returns The rendered component.
 */
export default function StackDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const [yamlContent, setYamlContent] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const stackName = params.name as string;

  useEffect(() => {
      async function load() {
          try {
              const yaml = await apiClient.getStackConfig(stackName);
              setYamlContent(yaml);
          } catch (e) {
              console.error("Failed to load stack config", e);
              toast({ variant: "destructive", title: "Error", description: "Failed to load stack configuration." });
          } finally {
              setLoading(false);
          }
      }
      load();
  }, [stackName, toast]);

  const handleSave = async (yamlContent: string) => {
      setSaving(true);
      try {
          await apiClient.saveStackConfig(stackName, yamlContent);
          toast({
              title: "Stack Updated",
              description: `Configuration for "${stackName}" has been saved.`
          });
      } catch (e: any) {
          console.error("Failed to update stack", e);
          toast({
              variant: "destructive",
              title: "Update Failed",
              description: e.message || "Failed to update stack."
          });
      } finally {
          setSaving(false);
      }
  };

  const handleDelete = async () => {
      if (!confirm(`Are you sure you want to delete stack "${stackName}"?`)) return;
      try {
          await apiClient.deleteCollection(stackName);
          toast({ title: "Stack Deleted", description: "Stack has been removed." });
          router.push("/stacks");
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  if (loading) {
      return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">{stackName}</h1>
            <p className="text-muted-foreground">Manage stack configuration.</p>
        </div>
        <Button variant="destructive" onClick={handleDelete}>
            <Trash2 className="mr-2 h-4 w-4" /> Delete Stack
        </Button>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden">
          <CardHeader>
              <CardTitle>Configuration</CardTitle>
              <CardDescription>
                  Edit the YAML configuration for this stack.
              </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden p-0">
              <div className="h-full p-6 pt-0">
                <StackEditor
                    initialYaml={yamlContent}
                    onSave={handleSave}
                    onCancel={() => router.back()}
                    isSaving={saving}
                />
              </div>
          </CardContent>
      </Card>
    </div>
  );
}
