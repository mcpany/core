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
import { AlertCircle, Upload, CheckCircle2, ArrowRight, ArrowLeft, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";

interface BulkServiceImportProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type ImportStep = 'input' | 'preview' | 'importing' | 'success';

/**
 * BulkServiceImport component allows users to import multiple services from a JSON config or URL.
 * It supports OpenAPI definitions and custom JSON lists.
 * @param props - Component props.
 * @returns The BulkServiceImport component.
 */
export function BulkServiceImport({ onImportSuccess, onCancel }: BulkServiceImportProps) {
    const [step, setStep] = useState<ImportStep>('input');
    const [jsonContent, setJsonContent] = useState("");
    const [importUrl, setImportUrl] = useState("");
    const [previewServices, setPreviewServices] = useState<any[]>([]);
    const [error, setError] = useState<string | null>(null);
    const { toast } = useToast();

    const parseServices = async (source: 'url' | 'json') => {
        let data: any;
        if (source === 'url') {
            const res = await fetch(importUrl);
            if (!res.ok) throw new Error(`Failed to fetch from URL: ${res.statusText}`);
            data = await res.json();
        } else {
            if (!jsonContent.trim()) throw new Error("JSON content is empty");
            data = JSON.parse(jsonContent);
        }

        let services: any[] = [];
        // Support OpenAPI/Swagger
        if (data.openapi || data.swagger) {
             services = [{
                name: data.info?.title?.toLowerCase().replace(/[^a-z0-9-]/g, '-') || "openapi-service",
                type: "OpenAPI",
                openapiService: {
                    address: source === 'url' ? importUrl : undefined,
                    specSource: source === 'url' ? { $case: "specUrl", specUrl: importUrl } : { $case: "specData", specData: JSON.stringify(data) },
                }
            }];
        } else {
            // Support generic list or wrapped object
            services = Array.isArray(data) ? data : (data.services || [data]);
        }

        if (!services || services.length === 0) {
            throw new Error("No services found in the provided configuration.");
        }

        return services.map(s => ({
            ...s,
            _type: s.openapiService ? "OpenAPI" : s.grpcService ? "gRPC" : s.httpService ? "HTTP" : s.mcpService ? "MCP" : "Unknown"
        }));
    };

    const handlePreview = async () => {
        setError(null);
        setStep('importing'); // transient state for loading preview
        try {
            const services = await parseServices(importUrl ? 'url' : 'json');
            setPreviewServices(services);
            setStep('preview');
        } catch (e: any) {
            console.error("Preview failed", e);
            setError(e.message || "Failed to parse configuration.");
            setStep('input');
        }
    };

    const handleImport = async () => {
        setStep('importing');
        setError(null);
        try {
            await Promise.all(previewServices.map(s => {
                // Remove internal fields before sending
                const { _type, ...serviceConfig } = s;
                return apiClient.registerService(serviceConfig);
            }));

            toast({
                title: "Import Successful",
                description: `Successfully imported ${previewServices.length} services.`,
                action: <CheckCircle2 className="h-5 w-5 text-green-500" />
            });
            setStep('success');
            setTimeout(() => onImportSuccess(), 1000);
        } catch (e: any) {
            console.error("Bulk import failed", e);
            setError(e.message || "Failed to import services.");
            setStep('preview');
        }
    };

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            setJsonContent(event.target?.result as string);
            setImportUrl("");
            setError(null);
        };
        reader.readAsText(file);
    };

    const isInputValid = importUrl.trim().length > 0 || jsonContent.trim().length > 0;

    if (step === 'input' || (step === 'importing' && previewServices.length === 0)) {
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
                                if (e.target.value) setJsonContent("");
                                setError(null);
                            }}
                            disabled={step === 'importing'}
                        />
                    </div>

                    <div className="relative">
                        <div className="absolute inset-0 flex items-center">
                            <span className="w-full border-t border-muted" />
                        </div>
                        <div className="relative flex justify-center text-xs uppercase">
                            <span className="bg-background px-2 text-muted-foreground">Or upload file</span>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>JSON Content</Label>
                        <div className="flex items-center gap-2 mb-2">
                             <div className="relative">
                                <Input
                                    type="file"
                                    accept=".json"
                                    onChange={handleFileUpload}
                                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                                    disabled={step === 'importing'}
                                />
                                <Button variant="outline" size="sm" className="pointer-events-none">
                                    <Upload className="h-4 w-4 mr-2" /> Upload JSON
                                </Button>
                             </div>
                        </div>
                        <Textarea
                            placeholder='[{"name": "service1", "httpService": {"address": "http://..."}}, ...]'
                            className="h-48 font-mono text-xs resize-none"
                            value={jsonContent}
                            onChange={(e) => {
                                setJsonContent(e.target.value);
                                if (e.target.value) setImportUrl("");
                                setError(null);
                            }}
                            disabled={step === 'importing'}
                        />
                    </div>
                </div>

                {error && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Error</AlertTitle>
                        <AlertDescription>{error}</AlertDescription>
                    </Alert>
                )}

                <div className="flex justify-end gap-2">
                    <Button variant="ghost" onClick={onCancel}>Cancel</Button>
                    <Button onClick={handlePreview} disabled={!isInputValid || step === 'importing'}>
                        {step === 'importing' ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <ArrowRight className="h-4 w-4 mr-2" />}
                        Preview Import
                    </Button>
                </div>
            </div>
        );
    }

    // Preview Step
    return (
        <div className="space-y-6">
            <div className="flex flex-col gap-2">
                <div className="text-sm text-muted-foreground">
                    Found <span className="font-medium text-foreground">{previewServices.length}</span> services to import.
                    Please review the list below.
                </div>

                <ScrollArea className="h-[300px] border rounded-md">
                    <Table>
                        <TableHeader className="bg-muted/50 sticky top-0 backdrop-blur-sm z-10">
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Address/Config</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {previewServices.map((service, i) => (
                                <TableRow key={i}>
                                    <TableCell className="font-medium">{service.name || "Unnamed"}</TableCell>
                                    <TableCell>
                                        <Badge variant="secondary" className="text-xs">{service._type}</Badge>
                                    </TableCell>
                                    <TableCell className="text-xs font-mono text-muted-foreground truncate max-w-[200px]">
                                        {service.openapiService?.address ||
                                         service.httpService?.address ||
                                         service.grpcService?.address ||
                                         "Local / Custom"}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </ScrollArea>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Import Failed</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={() => setStep('input')} disabled={step === 'importing'}>
                    <ArrowLeft className="h-4 w-4 mr-2" /> Back
                </Button>
                <Button onClick={handleImport} disabled={step === 'importing'}>
                    {step === 'importing' ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <CheckCircle2 className="h-4 w-4 mr-2" />}
                    Confirm Import
                </Button>
            </div>
        </div>
    );
}
