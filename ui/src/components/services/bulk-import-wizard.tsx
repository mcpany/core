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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FileText, Globe, Loader2, ArrowRight, ArrowLeft, RefreshCw, AlertTriangle, XCircle, CheckCircle2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

interface ValidationResult {
    valid: boolean;
    error?: string;
    details?: string;
    checking: boolean;
}

type Step = 'input' | 'validate' | 'select' | 'import' | 'summary';

export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [step, setStep] = useState<Step>('input');
    const [inputType, setInputType] = useState<'json' | 'url'>('json');
    const [inputText, setInputText] = useState("");
    const [inputUrl, setInputUrl] = useState("");

    const [parsedServices, setParsedServices] = useState<UpstreamServiceConfig[]>([]);
    const [validationResults, setValidationResults] = useState<Record<string, ValidationResult>>({});

    const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set());

    const [importResults, setImportResults] = useState<Record<string, { success: boolean, error?: string }>>({});
    const [importProgress, setImportProgress] = useState(0);
    const [isProcessing, setIsProcessing] = useState(false);

    const { toast } = useToast();

    // Step 1: Parse Input
    const parseInput = async () => {
        setIsProcessing(true);
        try {
            let services: any[] = [];

            if (inputType === 'url') {
                if (!inputUrl) throw new Error("URL is required");
                const res = await fetch(inputUrl);
                if (!res.ok) throw new Error(`Failed to fetch URL: ${res.statusText}`);
                const data = await res.json();
                 if (data.openapi || data.swagger) {
                    services = [{
                        name: data.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                        openapiService: {
                            address: inputUrl,
                            specUrl: inputUrl,
                            tools: [], resources: [], calls: [], prompts: []
                        }
                    }];
                } else {
                    services = Array.isArray(data) ? data : (data.services || [data]);
                }
            } else {
                if (!inputText) throw new Error("JSON content is required");
                const data = JSON.parse(inputText);
                services = Array.isArray(data) ? data : (data.services || [data]);
            }

            if (!services || services.length === 0) {
                throw new Error("No services found in input.");
            }

            // Normalize and assign temp IDs for tracking if needed
            const configs = services.map((s, i) => {
                if (!s.name) s.name = `service-${i+1}`;
                return s as UpstreamServiceConfig;
            });

            setParsedServices(configs);
            setStep('validate');

            // Trigger validation automatically
            validateServices(configs);

        } catch (e: any) {
            toast({
                title: "Parse Error",
                description: e.message || "Failed to parse input.",
                variant: "destructive"
            });
        } finally {
            setIsProcessing(false);
        }
    };

    const validateServices = async (configs: UpstreamServiceConfig[]) => {
        const results: Record<string, ValidationResult> = {};
        // Initialize state
        configs.forEach((_, i) => {
            results[i] = { valid: false, checking: true };
        });
        setValidationResults(results);

        // Run in parallel but don't block UI
        configs.forEach(async (config, index) => {
            try {
                // Remove ID to treat as new check
                const checkConfig = { ...config, id: "" };
                const res = await apiClient.validateService(checkConfig);

                setValidationResults(prev => ({
                    ...prev,
                    [index]: {
                        valid: res.valid,
                        error: res.error,
                        details: res.details || res.message, // handle various response formats
                        checking: false
                    }
                }));

                // Pre-select valid ones
                if (res.valid) {
                    setSelectedIndices(prev => new Set(prev).add(index));
                }
            } catch (e: any) {
                 setValidationResults(prev => ({
                    ...prev,
                    [index]: {
                        valid: false,
                        error: e.message,
                        checking: false
                    }
                }));
            }
        });
    };

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            setInputText(event.target?.result as string);
            setInputType('json');
        };
        reader.readAsText(file);
    };

    const executeImport = async () => {
        setStep('import');
        setIsProcessing(true);
        setImportProgress(0);

        const indices = Array.from(selectedIndices);
        const total = indices.length;
        const results: Record<string, { success: boolean, error?: string }> = {};

        let completed = 0;

        for (const index of indices) {
            const config = parsedServices[index];
            try {
                // Ensure ID is empty for create
                const configToSave = { ...config, id: "" };
                await apiClient.registerService(configToSave);
                results[index] = { success: true };
            } catch (e: any) {
                results[index] = { success: false, error: e.message };
            }
            completed++;
            setImportProgress((completed / total) * 100);
            setImportResults({ ...results });
        }

        setIsProcessing(false);
        setStep('summary');

        const failures = Object.values(results).filter(r => !r.success).length;
        if (failures === 0) {
             toast({
                title: "Import Complete",
                description: `Successfully imported ${total} services.`,
                action: <CheckCircle2 className="h-5 w-5 text-green-500" />
            });
            // Don't auto close, let user see summary
        } else {
             toast({
                title: "Import Completed with Errors",
                description: `Imported ${total - failures}/${total} services.`,
                variant: "destructive"
            });
        }
    };

    // UI Renders

    const renderInputStep = () => (
        <div className="space-y-4 py-4">
            <Tabs value={inputType} onValueChange={(v) => setInputType(v as any)} className="w-full">
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="json"><FileText className="mr-2 h-4 w-4"/> Paste JSON / File</TabsTrigger>
                    <TabsTrigger value="url"><Globe className="mr-2 h-4 w-4"/> Import from URL</TabsTrigger>
                </TabsList>
                <TabsContent value="json" className="space-y-4 mt-4">
                     <div className="space-y-2">
                        <Label>Upload File</Label>
                        <div className="flex items-center gap-2">
                            <Input
                                type="file"
                                accept=".json,.yaml,.yml"
                                onChange={handleFileUpload}
                                className="text-xs h-9 cursor-pointer"
                            />
                        </div>
                    </div>
                    <div className="relative">
                        <div className="absolute inset-0 flex items-center">
                            <span className="w-full border-t" />
                        </div>
                        <div className="relative flex justify-center text-xs uppercase">
                            <span className="bg-background px-2 text-muted-foreground">Or paste content</span>
                        </div>
                    </div>
                     <div className="space-y-2">
                        <Label>Configuration Content (JSON/YAML)</Label>
                        <Textarea
                            placeholder='[{"name": "service1", ...}]'
                            className="h-64 font-mono text-xs"
                            value={inputText}
                            onChange={(e) => setInputText(e.target.value)}
                        />
                    </div>
                </TabsContent>
                <TabsContent value="url" className="space-y-4 mt-4">
                     <div className="space-y-2">
                        <Label>URL (JSON or OpenAPI)</Label>
                        <Input
                            placeholder="https://api.example.com/openapi.json"
                            value={inputUrl}
                            onChange={(e) => setInputUrl(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">
                            Supports generic JSON service lists or OpenAPI/Swagger definitions (auto-converted).
                        </p>
                    </div>
                </TabsContent>
            </Tabs>

            <div className="flex justify-end gap-2 pt-4">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={parseInput} disabled={isProcessing || (inputType === 'json' && !inputText) || (inputType === 'url' && !inputUrl)}>
                    {isProcessing ? <Loader2 className="mr-2 h-4 w-4 animate-spin"/> : null}
                    Next <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
            </div>
        </div>
    );

    const renderValidateStep = () => {
        const validCount = Object.values(validationResults).filter(r => r.valid).length;
        const totalCount = parsedServices.length;
        const isChecking = Object.values(validationResults).some(r => r.checking);

        return (
            <div className="space-y-4 py-4">
                <div className="flex items-center justify-between">
                    <div>
                         <h3 className="text-lg font-medium">Review & Validate</h3>
                         <p className="text-sm text-muted-foreground">
                             Found {totalCount} services. {validCount} passed validation.
                         </p>
                    </div>
                    <Button variant="ghost" size="sm" onClick={() => validateServices(parsedServices)} disabled={isChecking}>
                        <RefreshCw className={cn("mr-2 h-4 w-4", isChecking && "animate-spin")} /> Re-validate
                    </Button>
                </div>

                <div className="border rounded-md max-h-[400px] overflow-y-auto">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[40px]">
                                    <Checkbox
                                        checked={selectedIndices.size === parsedServices.length && parsedServices.length > 0}
                                        onCheckedChange={(checked) => {
                                            if (checked) setSelectedIndices(new Set(parsedServices.map((_, i) => i)));
                                            else setSelectedIndices(new Set());
                                        }}
                                    />
                                </TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Service Name</TableHead>
                                <TableHead>Message</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {parsedServices.map((service, index) => {
                                const result = validationResults[index] || { checking: true, valid: false };
                                return (
                                    <TableRow key={index} className={result.valid ? "" : "bg-muted/30"}>
                                        <TableCell>
                                            <Checkbox
                                                checked={selectedIndices.has(index)}
                                                onCheckedChange={(checked) => {
                                                    setSelectedIndices(prev => {
                                                        const next = new Set(prev);
                                                        if (checked) next.add(index);
                                                        else next.delete(index);
                                                        return next;
                                                    });
                                                }}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            {result.checking ? (
                                                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                                            ) : result.valid ? (
                                                <CheckCircle2 className="h-4 w-4 text-green-500" />
                                            ) : (
                                                <TooltipProvider>
                                                    <Tooltip>
                                                        <TooltipTrigger>
                                                            <XCircle className="h-4 w-4 text-destructive" />
                                                        </TooltipTrigger>
                                                        <TooltipContent>
                                                            <p className="text-xs">{result.error}</p>
                                                        </TooltipContent>
                                                    </Tooltip>
                                                </TooltipProvider>
                                            )}
                                        </TableCell>
                                        <TableCell className="font-medium">{service.name}</TableCell>
                                        <TableCell className="text-xs text-muted-foreground truncate max-w-[300px]" title={result.details || result.error}>
                                            {result.details || result.error || (result.valid ? "Ready to import" : "Waiting...")}
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                        </TableBody>
                    </Table>
                </div>

                 <div className="flex justify-between pt-4">
                    <Button variant="outline" onClick={() => setStep('input')}>
                        <ArrowLeft className="mr-2 h-4 w-4" /> Back
                    </Button>
                    <div className="flex gap-2">
                         <span className="flex items-center text-sm text-muted-foreground mr-2">
                             {selectedIndices.size} selected
                         </span>
                         <Button onClick={executeImport} disabled={selectedIndices.size === 0 || isChecking}>
                            Import Selected <ArrowRight className="ml-2 h-4 w-4" />
                        </Button>
                    </div>
                </div>
            </div>
        );
    };

    const renderImportStep = () => (
        <div className="space-y-6 py-8">
            <div className="space-y-2 text-center">
                <h3 className="text-lg font-medium">Importing Services...</h3>
                <p className="text-sm text-muted-foreground">Please wait while we register your services.</p>
            </div>

            <Progress value={importProgress} className="w-full h-2" />

            <div className="text-center text-xs text-muted-foreground">
                {Math.round(importProgress)}% Complete
            </div>
        </div>
    );

    const renderSummaryStep = () => {
        const indices = Array.from(selectedIndices);
        const successCount = indices.filter(i => importResults[i]?.success).length;
        const failCount = indices.filter(i => !importResults[i]?.success).length;

        return (
            <div className="space-y-4 py-4">
                <div className="flex flex-col items-center gap-2 py-4">
                    {failCount === 0 ? (
                        <div className="h-12 w-12 rounded-full bg-green-100 flex items-center justify-center">
                            <CheckCircle2 className="h-6 w-6 text-green-600" />
                        </div>
                    ) : (
                         <div className="h-12 w-12 rounded-full bg-yellow-100 flex items-center justify-center">
                            <AlertTriangle className="h-6 w-6 text-yellow-600" />
                        </div>
                    )}
                    <h3 className="text-xl font-semibold">Import Complete</h3>
                    <p className="text-muted-foreground text-center">
                        Successfully imported {successCount} services.
                        {failCount > 0 && ` Failed to import ${failCount} services.`}
                    </p>
                </div>

                {failCount > 0 && (
                     <div className="border rounded-md max-h-[200px] overflow-y-auto">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Service</TableHead>
                                    <TableHead>Error</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {indices.map(i => {
                                    const res = importResults[i];
                                    if (res?.success) return null;
                                    return (
                                        <TableRow key={i}>
                                            <TableCell className="font-medium">{parsedServices[i].name}</TableCell>
                                            <TableCell className="text-destructive text-xs">{res?.error}</TableCell>
                                        </TableRow>
                                    );
                                })}
                            </TableBody>
                        </Table>
                     </div>
                )}

                <div className="flex justify-end pt-4">
                    <Button onClick={onImportSuccess}>Close</Button>
                </div>
            </div>
        );
    };

    return (
        <Card className="border-0 shadow-none">
            <CardContent className="p-0">
                {step === 'input' && renderInputStep()}
                {step === 'validate' && renderValidateStep()}
                {step === 'import' && renderImportStep()}
                {step === 'summary' && renderSummaryStep()}
            </CardContent>
        </Card>
    );
}
