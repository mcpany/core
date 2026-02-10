/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";

const DEFAULT_STACK_YAML = `name: my-stack
description: "A description of my stack"
services:
  - name: weather
    mcpService:
      httpConnection:
        httpAddress: http://localhost:8082
`;

/**
 * StackDetailPage component.
 * @returns The rendered component.
 */
export default function StackDetailPage() {
  const { stackId } = useParams<{ stackId: string }>();
  const router = useRouter();
  const { toast } = useToast();
  const [yaml, setYaml] = useState<string>("");
  const [stackData, setStackData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("config");

  useEffect(() => {
    if (stackId === "new") {
      setYaml(DEFAULT_STACK_YAML);
      setLoading(false);
      setActiveTab("config");
      return;
    }

    // Default to overview for existing stacks
    setActiveTab("overview");

    const loadStack = async () => {
      try {
        const [content, data] = await Promise.all([
            apiClient.getStackYaml(stackId),
            apiClient.getCollection(stackId).catch(() => null) // Ignore error if collection fetch fails but yaml works?
        ]);
        setYaml(content);
        setStackData(data);
      } catch (e) {
        console.error("Failed to load stack", e);
        toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load stack configuration."
        });
      } finally {
        setLoading(false);
      }
    };
    loadStack();
  }, [stackId, toast]);

  const handleSave = async (content: string) => {
    try {
      let targetId = stackId;
      if (stackId === "new") {
        const match = content.match(/^name:\s*(.+)$/m);
        if (match) {
            targetId = match[1].trim().replace(/['"]/g, "");
        } else {
            toast({
                variant: "destructive",
                title: "Invalid Config",
                description: "Stack name is required in YAML (name: ...)"
            });
            throw new Error("Stack name missing");
        }
      }

      await apiClient.saveStackYaml(targetId, content);
      toast({
        title: "Stack Saved",
        description: "Configuration has been applied successfully."
      });

      if (stackId === "new") {
        router.push(`/stacks/${targetId}`);
      } else {
          // Refresh data
          const data = await apiClient.getCollection(targetId);
          setStackData(data);
      }
    } catch (e) {
      console.error("Failed to save stack", e);
      toast({
        variant: "destructive",
        title: "Save Failed",
        description: String(e)
      });
      throw e;
    }
  };

  if (loading) {
    return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4">
      <div className="flex items-center gap-4">
        <Link href="/stacks">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {stackId === "new" ? "New Stack" : stackId}
          </h1>
          <p className="text-muted-foreground text-sm">
            {stackId === "new" ? "Define a new stack using YAML." : "Manage stack services and configuration."}
          </p>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-fit">
            <TabsTrigger value="overview" disabled={stackId === "new"}>Overview</TabsTrigger>
            <TabsTrigger value="config">Configuration</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="flex-1 overflow-auto mt-4">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {stackData?.services?.map((svc: any) => (
                    <Card key={svc.name}>
                        <CardHeader>
                            <CardTitle className="text-base flex items-center justify-between">
                                {svc.name}
                                <Badge variant="outline">{svc.version || "latest"}</Badge>
                            </CardTitle>
                            <CardDescription className="text-xs truncate">
                                {svc.id || svc.name}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="text-xs text-muted-foreground">
                                {svc.mcpService ? "MCP Service" :
                                 svc.commandLineService ? "Command Line" :
                                 svc.httpService ? "HTTP Service" : "Unknown Type"}
                            </div>
                        </CardContent>
                    </Card>
                ))}
                {(!stackData?.services || stackData.services.length === 0) && (
                    <div className="col-span-full text-center p-8 text-muted-foreground border-2 border-dashed rounded-lg">
                        No services in this stack.
                    </div>
                )}
            </div>
        </TabsContent>

        <TabsContent value="config" className="flex-1 min-h-0 mt-4 h-full">
             <StackEditor
                initialValue={yaml}
                onSave={handleSave}
                onCancel={() => router.push("/stacks")}
            />
        </TabsContent>
      </Tabs>
    </div>
  );
}
