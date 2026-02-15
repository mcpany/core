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
import {
    AlertCircle,
    Upload,
    CheckCircle2,
    FileJson,
    Globe,
    FileCode,
    Loader2,
    ArrowRight,
    ArrowLeft,
    XCircle,
    AlertTriangle
} from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Checkbox } from "@/components/ui/checkbox";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Progress } from "@/components/ui/progress";
import { apiClient, UpstreamServiceConfig, mapUpstreamServiceConfig } from "@/lib/client";
import { cn } from "@/lib/utils";

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type Step = "input" | "validate" | "select" | "import" | "result";

interface ParsedService {
    config: UpstreamServiceConfig;
    raw: any;
    valid: boolean;
    error?: string;
    details?: string;
}

interface ImportResult {
    name: string;
    success: boolean;
    error?: string;
}

export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [step, setStep] = useState<Step>("input");

    // Input State
    const [inputType, setInputType] = useState<"text" | "file" | "url">("text");
    const [textInput, setTextInput] = useState("");
    const [urlInput, setUrlInput] = useState("");
    const [fileInput, setFileInput] = useState<File | null>(null);
    const [inputError, setInputError] = useState<string | null>(null);

    // Processing State
    const [isProcessing, setIsProcessing] = useState(false);

    // Parsed Data
    const [parsedServices, setParsedServices] = useState<ParsedService[]>([]);

    // Selection State
    const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set());

    // Import State
    const [importProgress, setImportProgress] = useState(0);
    const [importResults, setImportResults] = useState<ImportResult[]>([]);

    // Step 1: Input Handlers
    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setFileInput(file);
            setInputError(null);
        }
    };

    const handleNextToValidate = async () => {
        setIsProcessing(true);
        setInputError(null);
        setParsedServices([]);

        try {
            let content = "";
            let sourceName = "";

            if (inputType === "text") {
                content = textInput;
                sourceName = "Paste";
            } else if (inputType === "url") {
                if (!urlInput) throw new Error("URL is required");
                const res = await fetch(urlInput);
                if (!res.ok) throw new Error(`Failed to fetch URL: ${res.statusText}`);
                content = await res.text();
                sourceName = urlInput;
            } else if (inputType === "file") {
                if (!fileInput) throw new Error("File is required");
                content = await fileInput.text();
                sourceName = fileInput.name;
            }

            if (!content.trim()) throw new Error("Input is empty");

            // Parse JSON
            let data;
            try {
                data = JSON.parse(content);
            } catch (e) {
                // TODO: Add YAML support if needed
                throw new Error("Invalid JSON format");
            }

            // Normalize to array
            let services: any[] = [];
            if (Array.isArray(data)) {
                services = data;
            } else if (data.services && Array.isArray(data.services)) {
                services = data.services;
            } else if (data.openapi || data.swagger) {
                 // OpenAPI handling (basic)
                 services = [{
                    name: data.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                    openapiService: {
                        address: inputType === 'url' ? urlInput : undefined,
                        specContent: content,
                    }
                }];
            } else {
                services = [data]; // Single object
            }

            if (services.length === 0) throw new Error("No services found in input");

            // Validate each service
            const results: ParsedService[] = [];

            // Parallel validation? Limit concurrency for large lists
            for (const s of services) {
                 // Basic shape check
                 if (!s.name && !s.id) {
                     results.push({
                         config: s,
                         raw: s,
                         valid: false,
                         error: "Service name or ID is required"
                     });
                     continue;
                 }

                 // Normalize snake_case to camelCase
                 let config = s;
                 // Heuristic: if any snake_case property exists, map it.
                 // Otherwise assume camelCase (or mixed, which map handles poorly but safe enough for now).
                 if (s.http_service || s.grpc_service || s.command_line_service || s.mcp_service || s.openapi_service) {
                     config = mapUpstreamServiceConfig(s);
                 }

                 // Server-side validation
                 try {
                     const validation = await apiClient.validateService(config);
                     results.push({
                         config: config,
                         raw: s,
                         valid: validation.valid,
                         error: validation.error,
                         details: validation.details || validation.message
                     });
                 } catch (e: any) {
                     results.push({
                         config: s,
                         raw: s,
                         valid: false,
                         error: e.message || "Validation failed"
                     });
                 }
            }

            setParsedServices(results);

            // Auto-select valid ones
            const validIndices = results.map((r, i) => r.valid ? i : -1).filter(i => i !== -1);
            setSelectedIndices(new Set(validIndices));

            setStep("select");

        } catch (e: any) {
            setInputError(e.message);
        } finally {
            setIsProcessing(false);
        }
    };

    // Step 2: Selection Handlers (skipped strict 'validate' step UI, combined with select)
    // Actually the plan said "Step 2: Analysis & Validation" and "Step 3: Selection".
    // I am combining the display of validation results with selection checkboxes in one view ("select" step).

    const handleSelectAll = (checked: boolean) => {
        if (checked) {
            const validIndices = parsedServices.map((r, i) => r.valid ? i : -1).filter(i => i !== -1);
            setSelectedIndices(new Set(validIndices));
        } else {
            setSelectedIndices(new Set());
        }
    };

    const handleSelectOne = (index: number, checked: boolean) => {
        const newSet = new Set(selectedIndices);
        if (checked) newSet.add(index);
        else newSet.delete(index);
        setSelectedIndices(newSet);
    };

    // Step 3: Import Handlers
    const handleImport = async () => {
        setStep("import");
        setImportResults([]);
        setImportProgress(0);

        const indices = Array.from(selectedIndices);
        const total = indices.length;
        const results: ImportResult[] = [];

        for (let i = 0; i < total; i++) {
            const index = indices[i];
            const service = parsedServices[index];

            try {
                await apiClient.registerService(service.config);
                results.push({ name: service.config.name, success: true });
            } catch (e: any) {
                results.push({ name: service.config.name, success: false, error: e.message });
            }

            setImportResults([...results]);
            setImportProgress(Math.round(((i + 1) / total) * 100));
        }

        setStep("result");
    };

    // Render Steps
    const renderInputStep = () => (
        <div className="space-y-4">
            <Tabs value={inputType} onValueChange={(v) => setInputType(v as any)} className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="text"><FileCode className="mr-2 h-4 w-4" /> JSON Text</TabsTrigger>
                    <TabsTrigger value="file"><Upload className="mr-2 h-4 w-4" /> Upload File</TabsTrigger>
                    <TabsTrigger value="url"><Globe className="mr-2 h-4 w-4" /> URL</TabsTrigger>
                </TabsList>

                <TabsContent value="text" className="mt-4">
                    <Label htmlFor="text-input">Paste JSON Configuration</Label>
                    <Textarea
                        id="text-input"
                        placeholder='[{"name": "service-1", ...}]'
                        className="h-64 font-mono text-xs mt-2"
                        value={textInput}
                        onChange={(e) => setTextInput(e.target.value)}
                    />
                </TabsContent>

                <TabsContent value="file" className="mt-4">
                    <Label htmlFor="file-input">Upload JSON Config File</Label>
                    <div className="mt-2 border-2 border-dashed rounded-lg p-10 flex flex-col items-center justify-center text-center hover:bg-muted/50 transition-colors">
                        <Upload className="h-10 w-10 text-muted-foreground mb-4" />
                        <Input
                            id="file-input"
                            type="file"
                            accept=".json"
                            className="hidden"
                            onChange={handleFileUpload}
                        />
                         <Button variant="secondary" onClick={() => document.getElementById('file-input')?.click()}>
                            Select File
                        </Button>
                        {fileInput && (
                            <div className="mt-4 flex items-center gap-2 text-sm font-medium">
                                <FileJson className="h-4 w-4 text-primary" />
                                {fileInput.name}
                            </div>
                        )}
                    </div>
                </TabsContent>

                <TabsContent value="url" className="mt-4">
                    <Label htmlFor="url-input">Import from URL</Label>
                    <div className="flex gap-2 mt-2">
                        <Input
                            id="url-input"
                            placeholder="https://example.com/config.json"
                            value={urlInput}
                            onChange={(e) => setUrlInput(e.target.value)}
                        />
                    </div>
                </TabsContent>
            </Tabs>

            {inputError && (
                <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{inputError}</AlertDescription>
                </Alert>
            )}

            <div className="flex justify-end pt-4">
                <Button onClick={handleNextToValidate} disabled={isProcessing}>
                    {isProcessing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                    Next: Validate
                </Button>
            </div>
        </div>
    );

    const renderSelectStep = () => (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Review & Select</h3>
                    <p className="text-sm text-muted-foreground">
                        Found {parsedServices.length} services. {parsedServices.filter(s => s.valid).length} valid.
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={() => setStep("input")}>
                        <ArrowLeft className="mr-2 h-4 w-4" /> Back
                    </Button>
                </div>
            </div>

            <div className="border rounded-md">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[50px]">
                                <Checkbox
                                    checked={parsedServices.length > 0 && selectedIndices.size === parsedServices.filter(s => s.valid).length}
                                    onCheckedChange={handleSelectAll}
                                    disabled={parsedServices.filter(s => s.valid).length === 0}
                                />
                            </TableHead>
                            <TableHead className="w-[100px]">Status</TableHead>
                            <TableHead>Service Name</TableHead>
                            <TableHead>Type</TableHead>
                            <TableHead className="text-right">Validation Details</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {parsedServices.map((service, index) => (
                            <TableRow key={index} className={!service.valid ? "bg-destructive/5" : ""}>
                                <TableCell>
                                    <Checkbox
                                        checked={selectedIndices.has(index)}
                                        onCheckedChange={(checked) => handleSelectOne(index, !!checked)}
                                        disabled={!service.valid}
                                    />
                                </TableCell>
                                <TableCell>
                                    {service.valid ? (
                                        <div className="flex items-center text-green-600">
                                            <CheckCircle2 className="h-4 w-4 mr-1" /> Valid
                                        </div>
                                    ) : (
                                        <div className="flex items-center text-destructive">
                                            <XCircle className="h-4 w-4 mr-1" /> Error
                                        </div>
                                    )}
                                </TableCell>
                                <TableCell className="font-medium">{service.config.name || "Unknown"}</TableCell>
                                <TableCell className="text-muted-foreground text-xs">
                                    {service.config.httpService ? "HTTP" :
                                     service.config.grpcService ? "gRPC" :
                                     service.config.commandLineService ? "CLI" :
                                     service.config.mcpService ? "MCP" : "Other"}
                                </TableCell>
                                <TableCell className="text-right text-xs max-w-[200px] truncate">
                                    {service.error ? (
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger>
                                                    <span className="text-destructive underline decoration-dotted">{service.error}</span>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p>{service.error}</p>
                                                    {service.details && <p className="mt-1 text-xs opacity-80">{service.details}</p>}
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    ) : (
                                        <span className="text-muted-foreground">Ready to import</span>
                                    )}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>

            <div className="flex justify-end gap-2 pt-4">
                 <Button variant="outline" onClick={onCancel}>Cancel</Button>
                 <Button onClick={handleImport} disabled={selectedIndices.size === 0}>
                    Import {selectedIndices.size} Services <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
            </div>
        </div>
    );

    const renderImportStep = () => (
        <div className="space-y-6 py-10">
            <div className="text-center space-y-2">
                <h3 className="text-lg font-medium">Importing Services...</h3>
                <p className="text-sm text-muted-foreground">
                    Processed {importResults.length} of {selectedIndices.size}
                </p>
            </div>

            <Progress value={importProgress} className="w-full" />

            <div className="border rounded-md h-48 overflow-y-auto p-4 bg-muted/20 text-sm font-mono">
                {importResults.map((res, i) => (
                    <div key={i} className="flex items-center gap-2 mb-1">
                        {res.success ? (
                            <CheckCircle2 className="h-3 w-3 text-green-500" />
                        ) : (
                            <XCircle className="h-3 w-3 text-destructive" />
                        )}
                        <span>{res.name}</span>
                        {res.error && <span className="text-destructive ml-2">- {res.error}</span>}
                    </div>
                ))}
                {importResults.length === 0 && <span className="text-muted-foreground opacity-50">Starting import...</span>}
            </div>
        </div>
    );

    const renderResultStep = () => {
        const successes = importResults.filter(r => r.success).length;
        const failures = importResults.filter(r => !r.success).length;

        return (
            <div className="space-y-6 py-4">
                <div className="flex flex-col items-center justify-center text-center gap-4">
                    {failures === 0 ? (
                        <div className="h-16 w-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center">
                            <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-500" />
                        </div>
                    ) : (
                        <div className="h-16 w-16 bg-yellow-100 dark:bg-yellow-900/20 rounded-full flex items-center justify-center">
                            <AlertTriangle className="h-8 w-8 text-yellow-600 dark:text-yellow-500" />
                        </div>
                    )}

                    <div>
                        <h3 className="text-xl font-bold">Import Complete</h3>
                        <p className="text-muted-foreground mt-1">
                            Successfully imported {successes} services.
                            {failures > 0 && ` ${failures} failed.`}
                        </p>
                    </div>
                </div>

                {failures > 0 && (
                     <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Import Errors</AlertTitle>
                        <AlertDescription>
                            Some services failed to import. Check the log below.
                        </AlertDescription>
                    </Alert>
                )}

                <div className="border rounded-md h-48 overflow-y-auto p-4 bg-muted/20 text-sm font-mono">
                     {importResults.map((res, i) => (
                        <div key={i} className={cn("flex items-center gap-2 mb-1", !res.success && "text-destructive")}>
                            {res.success ? (
                                <CheckCircle2 className="h-3 w-3 text-green-500" />
                            ) : (
                                <XCircle className="h-3 w-3" />
                            )}
                            <span>{res.name}</span>
                            {res.error && <span className="ml-2">- {res.error}</span>}
                        </div>
                    ))}
                </div>

                <div className="flex justify-end pt-4">
                    <Button onClick={onImportSuccess}>Close & View Services</Button>
                </div>
            </div>
        );
    }

    return (
        <div className="w-full">
            {step === "input" && renderInputStep()}
            {step === "select" && renderSelectStep()}
            {step === "import" && renderImportStep()}
            {step === "result" && renderResultStep()}
        </div>
    );
}
