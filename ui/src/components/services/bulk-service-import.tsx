/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle, Upload, CheckCircle2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";

interface BulkServiceImportProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

/**
 * BulkServiceImport component allows users to import multiple services from a JSON config or URL.
 * It supports OpenAPI definitions and custom JSON lists.
 * @param props - Component props.
 * @returns The BulkServiceImport component.
 */
export function BulkServiceImport({ onImportSuccess, onCancel }: BulkServiceImportProps) {
    const [jsonContent, setJsonContent] = useState("");
    const [importUrl, setImportUrl] = useState("");
    const [importing, setImporting] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    const handleImport = async () => {
        setImporting(true);
        setError(null);
        try {
            let servicesToImport: any[] = [];

            if (importUrl.trim()) {
                // Try to determine if it's OpenAPI or a direct JSON config
                const res = await fetch(importUrl);
                if (!res.ok) throw new Error("Failed to fetch from URL.");
                const data = await res.json();

                // If it looks like OpenAPI, we might need a converter or server-side discovery
                // For now, assume it's a JSON array of configs OR we use a template-like logic
                if (data.openapi || data.swagger) {
                    servicesToImport = [{
                        name: data.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                        openapiService: {
                            address: importUrl,
                            specSource: { $case: "specUrl", specUrl: importUrl },
                            tools: [], resources: [], calls: {}, prompts: []
                        }
                    }];
                } else {
                    servicesToImport = Array.isArray(data) ? data : (data.services || [data]);
                }
            } else {
                const data = JSON.parse(jsonContent);
                servicesToImport = Array.isArray(data) ? data : (data.services || [data]);
            }

            if (servicesToImport.length === 0) {
                throw new Error("No services found.");
            }

            await Promise.all(servicesToImport.map((s: any) => apiClient.registerService(s)));

            toast({
                title: "Import Successful",
                description: `Successfully imported ${servicesToImport.length} services.`,
                action: <CheckCircle2 className="h-5 w-5 text-green-500" />
            });
            onImportSuccess();
        } catch (e: any) {
            console.error("Bulk import failed", e);
            setError(e.message || "Failed to parse or import. Please check the format/URL.");
        } finally {
            setImporting(false);
        }
    };

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            setJsonContent(event.target?.result as string);
            setImportUrl(""); // Clear URL if file uploaded
        };
        reader.readAsText(file);
    };

    return (
        <div className="space-y-6">
            <div className="space-y-4">
                <div className="space-y-2">
                    <Label htmlFor="import-url">Import from URL (OpenAPI/JSON)</Label>
                    <Input
                        id="import-url"
                        placeholder="https://api.example.com/openapi.json"
                        value={importUrl}
                        onChange={(e) => {
                            setImportUrl(e.target.value);
                            if (e.target.value) setJsonContent(""); // Clear JSON if URL set
                        }}
                    />
                </div>

                <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                        <span className="w-full border-t" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                        <span className="bg-background px-2 text-muted-foreground">Or upload file</span>
                    </div>
                </div>

                <div className="space-y-2">
                    <Label>JSON Content</Label>
                    <div className="flex items-center gap-2 mb-2">
                        <Input
                            type="file"
                            accept=".json"
                            onChange={handleFileUpload}
                            className="text-xs h-9 cursor-pointer"
                        />
                        <Button variant="outline" size="icon" className="h-9 w-9">
                            <Upload className="h-4 w-4" />
                        </Button>
                    </div>
                    <Textarea
                        placeholder='[{"name": "service1", "httpService": {"address": "http://..."}}, ...]'
                        className="h-48 font-mono text-xs"
                        value={jsonContent}
                        onChange={(e) => {
                            setJsonContent(e.target.value);
                            if (e.target.value) setImportUrl(""); // Clear URL if JSON set
                        }}
                    />
                </div>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Import Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleImport} disabled={importing || !jsonContent.trim()}>
                    {importing ? "Importing..." : "Import Services"}
                </Button>
            </div>
        </div>
    );
}
