/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, use } from "react";
import { useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { ArrowLeft, Save, Play, Loader2, Table as TableIcon, Code } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";

export default function StackDetailsPage({ params }: { params: Promise<{ stackId: string }> }) {
  const { stackId } = use(params);
  const isNew = stackId === "new";
  const router = useRouter();
  const { toast } = useToast();

  const [name, setName] = useState("");
  const [content, setContent] = useState(""); // JSON content
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState("editor");

  useEffect(() => {
    if (!isNew) {
        fetchStack();
    } else {
        // Default template
        setContent(JSON.stringify({
            name: "my-stack",
            description: "A new stack",
            version: "1.0.0",
            services: []
        }, null, 2));
    }
  }, [stackId]);

  const fetchStack = async () => {
      try {
          const res = await apiClient.getCollection(stackId);
          if (res) {
              setName(res.name);
              setContent(JSON.stringify(res, null, 2));
          }
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to load stack." });
      } finally {
          setLoading(false);
      }
  };

  const handleSave = async () => {
      setSaving(true);
      try {
          let collection: ServiceCollection;
          try {
              collection = JSON.parse(content);
          } catch (e) {
              toast({ variant: "destructive", title: "Invalid JSON", description: "Please ensure the content is valid JSON." });
              setSaving(false);
              return;
          }

          if (!isNew && collection.name !== stackId) {
               if (!confirm("Changing stack name will create a new stack. Continue?")) {
                   setSaving(false);
                   return;
               }
          }

          if (isNew && !collection.name) {
               toast({ variant: "destructive", title: "Error", description: "Stack name is required in JSON." });
               setSaving(false);
               return;
          }

          await apiClient.saveCollection(collection);
          toast({ title: "Stack Saved", description: "Configuration saved successfully." });

          if (isNew) {
              router.push(`/stacks/${collection.name}`);
          }
      } catch (e) {
          console.error(e);
          toast({ variant: "destructive", title: "Error", description: "Failed to save stack." });
      } finally {
          setSaving(false);
      }
  };

  const handleDeploy = async () => {
      await handleSave();
      const collectionName = isNew ? JSON.parse(content).name : stackId;
      if (!collectionName) return;

      try {
          await apiClient.applyCollection(collectionName);
          toast({ title: "Stack Deployed", description: "Stack applied successfully." });
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to deploy stack." });
      }
  };

  if (loading) {
      return <div className="flex items-center justify-center h-full"><Loader2 className="animate-spin" /></div>;
  }

  // Parse for preview
  let parsedServices: UpstreamServiceConfig[] = [];
  try {
      parsedServices = JSON.parse(content).services || [];
  } catch {}

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-8 pt-6 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" asChild>
                <Link href="/stacks">
                    <ArrowLeft className="h-4 w-4" />
                </Link>
            </Button>
            <h1 className="text-2xl font-bold tracking-tight">{isNew ? "New Stack" : name}</h1>
        </div>
        <div className="flex items-center gap-2">
            <Button variant="outline" onClick={handleSave} disabled={saving}>
                <Save className="mr-2 h-4 w-4" /> Save
            </Button>
            <Button onClick={handleDeploy} disabled={saving}>
                <Play className="mr-2 h-4 w-4" /> Deploy
            </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
          <TabsList>
              <TabsTrigger value="editor"><Code className="mr-2 h-4 w-4" /> Editor</TabsTrigger>
              <TabsTrigger value="preview"><TableIcon className="mr-2 h-4 w-4" /> Services Preview</TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="flex-1 mt-4 relative">
              <div className="flex flex-col h-full gap-2">
                  <Label className="sr-only">Configuration</Label>
                  <Textarea
                    className="font-mono text-xs flex-1 resize-none bg-muted/30"
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    spellCheck={false}
                  />
              </div>
          </TabsContent>

          <TabsContent value="preview" className="flex-1 mt-4 overflow-y-auto">
              <div className="grid gap-4">
                  {parsedServices.map((svc, idx) => (
                      <Card key={idx}>
                          <CardContent className="p-4 flex items-center justify-between">
                              <div className="flex items-center gap-4">
                                  <div className="font-semibold">{svc.name}</div>
                                  <Badge variant="outline">{svc.httpService ? "HTTP" : svc.commandLineService ? "CLI" : "Other"}</Badge>
                              </div>
                              <div className="text-sm text-muted-foreground font-mono">
                                  {svc.httpService?.address || svc.commandLineService?.command || "-"}
                              </div>
                          </CardContent>
                      </Card>
                  ))}
                  {parsedServices.length === 0 && (
                      <div className="text-center py-10 text-muted-foreground border-2 border-dashed rounded-lg">
                          No services defined in this stack.
                      </div>
                  )}
              </div>
          </TabsContent>
      </Tabs>
    </div>
  );
}
