/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient } from "@/lib/client";
import { StackEditor } from "@/components/stacks/stack-editor";
import { useToast } from "@/hooks/use-toast";
import yaml from "js-yaml";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function StackEditPage() {
    const params = useParams();
    const router = useRouter();
    const { toast } = useToast();
    const stackId = params.stackId as string;
    const isNew = stackId === "new";

    const [yamlContent, setYamlContent] = useState("");
    const [loading, setLoading] = useState(!isNew);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        if (!isNew) {
            loadStack();
        } else {
            // Default template for new stack
            setYamlContent(
`name: my-stack
description: A collection of services
version: 1.0.0
services:
  - name: my-service
    commandLineService:
      command: echo "Hello World"
`);
        }
    }, [stackId]);

    const loadStack = async () => {
        try {
            setLoading(true);
            const collection = await apiClient.getCollection(stackId);
            if (collection) {
                // Convert to YAML
                const yamlStr = yaml.dump(collection);
                setYamlContent(yamlStr);
            }
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

    const handleSave = async (content: string) => {
        setSaving(true);
        try {
            // Parse YAML to JSON
            const parsed = yaml.load(content) as any;
            if (!parsed || typeof parsed !== 'object') {
                throw new Error("Invalid YAML");
            }

            // Save via API
            await apiClient.saveCollection(parsed);

            toast({
                title: "Stack Saved",
                description: "Configuration deployed successfully."
            });

            if (isNew && parsed.name) {
                // Redirect to the edit page of the new stack
                router.push(`/stacks/${parsed.name}`);
            }
        } catch (e: any) {
            console.error("Failed to save stack", e);
            toast({
                variant: "destructive",
                title: "Save Failed",
                description: e.message || "Failed to save stack configuration."
            });
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return <div className="p-8">Loading...</div>;
    }

    return (
        <div className="flex flex-col h-[calc(100vh-4rem)] p-8 pt-6">
            <div className="flex items-center gap-4 mb-6">
                <Link href="/stacks">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="h-4 w-4" />
                    </Button>
                </Link>
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">
                        {isNew ? "New Stack" : `Edit ${stackId}`}
                    </h1>
                    <p className="text-muted-foreground">
                        {isNew ? "Create a new stack configuration." : "Edit stack configuration via YAML."}
                    </p>
                </div>
            </div>

            <div className="flex-1 min-h-0">
                <StackEditor
                    initialValue={yamlContent}
                    onSave={handleSave}
                    isSaving={saving}
                />
            </div>
        </div>
    );
}
