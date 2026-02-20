/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle, Upload, CheckCircle2, FileJson, Link as LinkIcon, Loader2, XCircle, AlertTriangle, FileCode } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import yaml from "js-yaml";

interface BulkServiceImportProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type ValidationStatus = "pending" | "valid" | "invalid" | "warning";

interface ServiceImportItem {
    config: UpstreamServiceConfig;
    validationStatus: ValidationStatus;
    validationMessage?: string;
    importStatus?: "pending" | "success" | "error";
    importError?: string;
    selected: boolean;
}

/**
 * BulkServiceImport provides a wizard-like interface for importing multiple service configurations.
 * It supports JSON/YAML input, file uploads, and URL imports with validation steps.
 */
export function BulkServiceImport({ onImportSuccess, onCancel }: BulkServiceImportProps) {
    const [step, setStep] = useState<"input" | "review" | "import">("input");

    // Input Step State
    const [inputType, setInputType] = useState<"json" | "url" | "file">("json");
    const [jsonContent, setJsonContent] = useState("");
    const [importUrl, setImportUrl] = useState("");
    const [parsingError, setParsingError] = useState<string | null>(null);
    const [detectedFormat, setDetectedFormat] = useState<string | null>(null);

    // Review Step State
    const [items, setItems] = useState<ServiceImportItem[]>([]);
    const [isValidating, setIsValidating] = useState(false);

    // Import Step State
    const [isImporting, setIsImporting] = useState(false);
    const [progress, setProgress] = useState(0);
    const [importSummary, setImportSummary] = useState<{ success: number; failed: number } | null>(null);

    const { toast } = useToast();

    // --- Step 1: Input Handling ---

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            setJsonContent(event.target?.result as string);
            setInputType("json"); // Switch to JSON view to verify
        };
        reader.readAsText(file);
    };

    const mapClaudeConfig = (mcpServers: any): UpstreamServiceConfig[] => {
        return Object.entries(mcpServers).map(([name, config]: [string, any]) => {
            // Check for SSE (url)
            if (config.url) {
                return {
                    id: "",
                    name: name,
                    version: "1.0.0",
                    priority: 0,
                    disable: false,
                    mcpService: {
                        httpConnection: {
                            httpAddress: config.url
                        },
                        toolAutoDiscovery: true,
                        tools: [], resources: [], prompts: [], calls: {}
                    }
                } as UpstreamServiceConfig;
            }

            // Claude config usually has "command", "args", "env"
            const command = config.command || "";
            const args = Array.isArray(config.args) ? config.args : [];
            const fullCommand = [command, ...args].join(" ");

            return {
                id: "", // Will be generated
                name: name,
                version: "1.0.0",
                priority: 0,
                disable: false,
                commandLineService: {
                    command: fullCommand,
                    workingDirectory: "",
                    env: config.env || {},
                    tools: [],
                    resources: [],
                    prompts: [],
                    calls: {},
                    communicationProtocol: 0,
                    local: false
                },
                mcpService: {
                   toolAutoDiscovery: true,
                   tools: [], resources: [], prompts: [], calls: {}
                }
            } as UpstreamServiceConfig;
        });
    };

    const detectAndParse = (content: string): UpstreamServiceConfig[] => {
        let parsed: any;
        let format = "unknown";

        // Try JSON first
        try {
            parsed = JSON.parse(content);
            format = "JSON";
        } catch (e) {
            // Try YAML
            try {
                parsed = yaml.load(content);
                format = "YAML";
            } catch (e2: any) {
                 throw new Error(`Failed to parse content as JSON or YAML. YAML Error: ${e2.message}`);
            }
        }

        if (!parsed || typeof parsed !== 'object') {
            throw new Error("Parsed content must be an object or array.");
        }

        setDetectedFormat(format);

        // Detect Schema
        if (parsed.mcpServers) {
            setDetectedFormat(`${format} (Claude Desktop)`);
            return mapClaudeConfig(parsed.mcpServers);
        } else if (parsed.services && Array.isArray(parsed.services)) {
             setDetectedFormat(`${format} (MCP Any Native)`);
             return parsed.services;
        } else if (Array.isArray(parsed)) {
             setDetectedFormat(`${format} (Array)`);
             return parsed;
        } else if (parsed.openapi || parsed.swagger) {
             setDetectedFormat(`${format} (OpenAPI)`);
             return [{
                name: parsed.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                openapiService: {
                    address: importUrl || "imported-spec",
                    specUrl: "",
                    specContent: content, // Store full spec? Or just assume structure?
                    tools: [], resources: [], calls: [], prompts: []
                }
            } as any];
        } else {
            // Assume single object is a service if it looks like one, or try to be flexible
             setDetectedFormat(`${format} (Single Service)`);
             return [parsed];
        }
    };

    const parseAndValidate = async () => {
        setParsingError(null);
        setIsValidating(true);
        setDetectedFormat(null);
        let parsedServices: any[] = [];

        try {
            let contentToParse = jsonContent;

            if (inputType === "url") {
                if (!importUrl.trim()) throw new Error("URL is required.");
                const res = await fetch(importUrl);
                if (!res.ok) throw new Error(`Failed to fetch from URL: ${res.statusText}`);
                // Get text first to allow detection
                contentToParse = await res.text();
            } else {
                 if (!jsonContent.trim()) throw new Error("Content is required.");
            }

            parsedServices = detectAndParse(contentToParse);

            if (!parsedServices.length) throw new Error("No services found in input.");

            // Initial items setup
            const initialItems: ServiceImportItem[] = parsedServices.map(s => ({
                config: s,
                validationStatus: "pending",
                selected: true // Select by default
            }));

            setItems(initialItems);
            setStep("review");

            // Trigger async validation for each
            validateItems(initialItems);

        } catch (e: any) {
            setParsingError(e.message || "Failed to parse input.");
            setIsValidating(false);
        }
    };

    const validateItems = async (itemsToValidate: ServiceImportItem[]) => {
        const validatedItems = [...itemsToValidate];

        // Parallel validation
        await Promise.all(validatedItems.map(async (item, index) => {
            try {
                // Ensure basic structure matches UpstreamServiceConfig
                // We assume input is roughly correct, validation endpoint will catch specifics
                const res = await apiClient.validateService(item.config);

                // Check connectivity result if available (assuming validateService returns details)
                if (res.valid) {
                    validatedItems[index].validationStatus = "valid";
                } else {
                     // Check if it's a warning or error based on details?
                     // For now, strict validation failure is invalid.
                     // But if we want to allow import anyway (e.g. offline), maybe warning?
                     // Let's treat connection errors as Warnings if config is syntactically valid?
                     // The backend handler returns valid=false for both.
                     // Let's use the error message to distinguish.
                     const isConnectionError = res.error?.toLowerCase().includes("reachability") || res.error?.toLowerCase().includes("failed to connect");

                     validatedItems[index].validationStatus = isConnectionError ? "warning" : "invalid";
                     validatedItems[index].validationMessage = res.error || res.message;

                     // Deselect invalid items by default, keep warnings selected
                     if (!isConnectionError) {
                         validatedItems[index].selected = false;
                     }
                }
            } catch (e: any) {
                validatedItems[index].validationStatus = "invalid";
                validatedItems[index].validationMessage = e.message;
                validatedItems[index].selected = false;
            }
        }));

        setItems(validatedItems);
        setIsValidating(false);
    };

    // --- Step 2: Review & Selection ---

    const toggleSelection = (index: number) => {
        setItems(prev => prev.map((item, i) => i === index ? { ...item, selected: !item.selected } : item));
    };

    const toggleSelectAll = (checked: boolean) => {
         setItems(prev => prev.map(item => ({
             ...item,
             selected: checked && item.validationStatus !== "invalid"
         })));
    };

    const selectedCount = items.filter(i => i.selected).length;
    const validCount = items.filter(i => i.validationStatus === "valid").length;
    const warningCount = items.filter(i => i.validationStatus === "warning").length;

    const startImport = async () => {
        setStep("import");
        setIsImporting(true);
        setProgress(0);

        const itemsToImport = items.filter(i => i.selected);
        let successCount = 0;
        let failureCount = 0;

        const results = [...items];

        for (let i = 0; i < itemsToImport.length; i++) {
            const item = itemsToImport[i];
            const originalIndex = items.findIndex(it => it === item);

            try {
                // Register
                await apiClient.registerService(item.config);
                results[originalIndex].importStatus = "success";
                successCount++;
            } catch (e: any) {
                results[originalIndex].importStatus = "error";
                results[originalIndex].importError = e.message;
                failureCount++;
            }

            // Update progress
            setProgress(Math.round(((i + 1) / itemsToImport.length) * 100));
            setItems([...results]); // Trigger re-render
        }

        setIsImporting(false);
        setImportSummary({ success: successCount, failed: failureCount });

        if (successCount > 0) {
             toast({
                title: "Import Complete",
                description: `Successfully imported ${successCount} services.`,
                variant: failureCount > 0 ? "default" : "default" // could use warning variant if partial
            });
        }
    };

    // --- Rendering ---

    if (step === "input") {
        return (
            <div className="space-y-6">
                <Tabs value={inputType} onValueChange={(v) => setInputType(v as any)} className="w-full">
                    <TabsList className="grid w-full grid-cols-3">
                        <TabsTrigger value="json"><FileCode className="mr-2 h-4 w-4" /> JSON / YAML</TabsTrigger>
                        <TabsTrigger value="file"><Upload className="mr-2 h-4 w-4" /> File Upload</TabsTrigger>
                        <TabsTrigger value="url"><LinkIcon className="mr-2 h-4 w-4" /> URL Import</TabsTrigger>
                    </TabsList>

                    <div className="mt-4 space-y-4">
                        {inputType === "file" && (
                            <div className="flex flex-col items-center justify-center border-2 border-dashed rounded-lg p-12 text-center hover:bg-muted/50 transition-colors">
                                <Upload className="h-10 w-10 text-muted-foreground mb-4" />
                                <h3 className="text-lg font-medium mb-2">Upload Configuration File</h3>
                                <p className="text-sm text-muted-foreground mb-4">
                                    Drag and drop your JSON/YAML config file here, or click to browse.
                                </p>
                                <Input
                                    type="file"
                                    accept=".json,.yaml,.yml"
                                    onChange={handleFileUpload}
                                    className="max-w-xs cursor-pointer"
                                />
                            </div>
                        )}

                        {(inputType === "json" || inputType === "file") && (
                            <div className={inputType === "file" ? "hidden" : "block"}>
                                <Label className="mb-2 block">Configuration Content</Label>
                                <Textarea
                                    placeholder='Paste JSON or YAML here...'
                                    className="h-[300px] font-mono text-xs"
                                    value={jsonContent}
                                    onChange={(e) => setJsonContent(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground mt-2">
                                    Supports: Standard MCP JSON, Claude Desktop Config (mcpServers), and YAML.
                                </p>
                            </div>
                        )}

                        {inputType === "url" && (
                            <div className="space-y-4 py-8">
                                <div className="space-y-2">
                                    <Label>Configuration URL</Label>
                                    <div className="flex gap-2">
                                        <Input
                                            placeholder="https://example.com/mcp-config.yaml"
                                            value={importUrl}
                                            onChange={(e) => setImportUrl(e.target.value)}
                                        />
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        Enter a URL to a JSON/YAML configuration file or OpenAPI specification.
                                    </p>
                                </div>
                            </div>
                        )}
                    </div>
                </Tabs>

                {parsingError && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>Parsing Error</AlertTitle>
                        <AlertDescription>{parsingError}</AlertDescription>
                    </Alert>
                )}

                <div className="flex justify-end gap-2 pt-4 border-t">
                    <Button variant="outline" onClick={onCancel}>Cancel</Button>
                    <Button onClick={parseAndValidate} disabled={
                        (inputType === "json" && !jsonContent.trim()) ||
                        (inputType === "url" && !importUrl.trim()) ||
                        isValidating
                    }>
                        {isValidating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                        Next: Review
                    </Button>
                </div>
            </div>
        );
    }

    if (step === "review") {
        return (
            <div className="space-y-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-lg font-medium">Review Services</h3>
                        <p className="text-sm text-muted-foreground flex items-center gap-2">
                            Found {items.length} services. {validCount} valid, {warningCount} warnings.
                            {detectedFormat && <Badge variant="secondary" className="text-[10px] ml-2">{detectedFormat}</Badge>}
                        </p>
                    </div>
                </div>

                <div className="border rounded-md max-h-[400px] overflow-y-auto">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[50px]">
                                    <Checkbox
                                        checked={selectedCount > 0 && selectedCount === items.filter(i => i.validationStatus !== 'invalid').length}
                                        onCheckedChange={(c) => toggleSelectAll(!!c)}
                                    />
                                </TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Name</TableHead>
                                <TableHead>Type</TableHead>
                                <TableHead>Address / Command</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {items.map((item, idx) => (
                                <TableRow key={idx} className={!item.selected ? "opacity-60" : ""}>
                                    <TableCell>
                                        <Checkbox
                                            checked={item.selected}
                                            onCheckedChange={() => toggleSelection(idx)}
                                            disabled={item.validationStatus === "invalid"}
                                        />
                                    </TableCell>
                                    <TableCell>
                                        {item.validationStatus === "pending" && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                                        {item.validationStatus === "valid" && (
                                            <TooltipProvider>
                                                <Tooltip>
                                                    <TooltipTrigger>
                                                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                                                    </TooltipTrigger>
                                                    <TooltipContent>Configuration is valid</TooltipContent>
                                                </Tooltip>
                                            </TooltipProvider>
                                        )}
                                        {item.validationStatus === "warning" && (
                                            <TooltipProvider>
                                                <Tooltip>
                                                    <TooltipTrigger>
                                                        <AlertTriangle className="h-4 w-4 text-yellow-500" />
                                                    </TooltipTrigger>
                                                    <TooltipContent>{item.validationMessage || "Warning"}</TooltipContent>
                                                </Tooltip>
                                            </TooltipProvider>
                                        )}
                                        {item.validationStatus === "invalid" && (
                                            <TooltipProvider>
                                                <Tooltip>
                                                    <TooltipTrigger>
                                                        <XCircle className="h-4 w-4 text-red-500" />
                                                    </TooltipTrigger>
                                                    <TooltipContent>{item.validationMessage || "Invalid configuration"}</TooltipContent>
                                                </Tooltip>
                                            </TooltipProvider>
                                        )}
                                    </TableCell>
                                    <TableCell className="font-medium">{item.config.name || "Unnamed"}</TableCell>
                                    <TableCell>
                                        <Badge variant="outline">
                                            {item.config.httpService ? "HTTP" :
                                             item.config.grpcService ? "gRPC" :
                                             item.config.commandLineService ? "CLI" :
                                             item.config.mcpService ? "MCP" : "Other"}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs max-w-[200px] truncate" title={
                                         item.config.httpService?.address ||
                                         item.config.grpcService?.address ||
                                         item.config.commandLineService?.command ||
                                         item.config.mcpService?.httpConnection?.httpAddress ||
                                         item.config.mcpService?.stdioConnection?.command ||
                                         "-"
                                    }>
                                        {item.config.httpService?.address ||
                                         item.config.grpcService?.address ||
                                         item.config.commandLineService?.command ||
                                         item.config.mcpService?.httpConnection?.httpAddress ||
                                         item.config.mcpService?.stdioConnection?.command ||
                                         "-"}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>

                <div className="flex justify-between items-center pt-4 border-t">
                    <Button variant="ghost" onClick={() => setStep("input")}>Back</Button>
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={onCancel}>Cancel</Button>
                        <Button onClick={startImport} disabled={selectedCount === 0 || isValidating}>
                            Import {selectedCount} Services
                        </Button>
                    </div>
                </div>
            </div>
        );
    }

    if (step === "import") {
        return (
            <div className="space-y-6 py-8">
                <div className="space-y-2 text-center">
                    <h3 className="text-xl font-medium">
                        {isImporting ? "Importing Services..." : "Import Complete"}
                    </h3>
                    <p className="text-muted-foreground">
                        {isImporting
                            ? `Processing ${items.filter(i => i.selected).length} services.`
                            : `Successfully imported ${importSummary?.success} services.`}
                    </p>
                </div>

                <div className="space-y-2">
                    <Progress value={progress} className="h-2" />
                    <div className="flex justify-between text-xs text-muted-foreground">
                        <span>0%</span>
                        <span>{progress}%</span>
                    </div>
                </div>

                {!isImporting && importSummary && (
                    <div className="space-y-4">
                         {importSummary.failed > 0 && (
                            <Alert variant="destructive">
                                <AlertCircle className="h-4 w-4" />
                                <AlertTitle>Import Failures</AlertTitle>
                                <AlertDescription>
                                    {importSummary.failed} services failed to import. Check the list below for details.
                                </AlertDescription>
                            </Alert>
                        )}

                        <div className="border rounded-md max-h-[200px] overflow-y-auto">
                             <Table>
                                <TableBody>
                                    {items.filter(i => i.selected).map((item, idx) => (
                                        <TableRow key={idx}>
                                            <TableCell className="w-[30px]">
                                                {item.importStatus === "success" && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                                                {item.importStatus === "error" && <XCircle className="h-4 w-4 text-red-500" />}
                                            </TableCell>
                                            <TableCell>{item.config.name}</TableCell>
                                            <TableCell className="text-xs text-red-500">
                                                {item.importError}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>

                        <div className="flex justify-center pt-4">
                            <Button onClick={onImportSuccess}>Close</Button>
                        </div>
                    </div>
                )}
            </div>
        );
    }

    return null;
}
