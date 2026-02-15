"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Progress } from "@/components/ui/progress";
import { FileJson, Link as LinkIcon, Upload, ArrowRight, ArrowLeft, Check, AlertCircle, Loader2, AlertTriangle, CheckCircle2 } from "lucide-react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import * as yaml from "js-yaml";

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type Step = "input" | "validation" | "import" | "result";

interface ParsedService {
    id: string;
    originalConfig: UpstreamServiceConfig;
    name: string;
    status: "pending" | "valid" | "error" | "warning";
    message?: string;
    details?: string;
    selected: boolean;
}

export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [step, setStep] = useState<Step>("input");
    const [activeTab, setActiveTab] = useState<string>("json");
    const { toast } = useToast();

    // Input State
    const [jsonContent, setJsonContent] = useState("");
    const [url, setUrl] = useState("");
    const [file, setFile] = useState<File | null>(null);

    // Validation State
    const [parsedServices, setParsedServices] = useState<ParsedService[]>([]);
    const [isValidating, setIsValidating] = useState(false);

    // Import State
    const [importProgress, setImportProgress] = useState(0);
    const [importResults, setImportResults] = useState<{ name: string; success: boolean; error?: string }[]>([]);
    const [isImporting, setIsImporting] = useState(false);

    const parseContent = async (): Promise<UpstreamServiceConfig[]> => {
        let data: unknown = null;

        if (activeTab === "json") {
             if (!jsonContent.trim()) throw new Error("Content is empty");
             try {
                data = JSON.parse(jsonContent);
             } catch {
                try {
                    data = yaml.load(jsonContent);
                } catch {
                    throw new Error("Failed to parse content as JSON or YAML.");
                }
             }
        } else if (activeTab === "url") {
             if (!url.trim()) throw new Error("URL is empty");
             const res = await fetch(url);
             if (!res.ok) throw new Error(`Failed to fetch URL: ${res.statusText}`);
             const text = await res.text();
             try {
                data = JSON.parse(text);
             } catch {
                try {
                    data = yaml.load(text);
                } catch {
                    throw new Error("Failed to parse URL content as JSON or YAML.");
                }
             }
        } else if (activeTab === "file") {
             if (!file) throw new Error("No file selected");
             const text = await file.text();
             try {
                data = JSON.parse(text);
             } catch {
                try {
                    data = yaml.load(text);
                } catch {
                    throw new Error("Failed to parse file content as JSON or YAML.");
                }
             }
        }

        if (!data) throw new Error("No data found");

        // OpenAPI / Swagger Auto-detection
        // Cast data to any to check for properties safely
        const anyData = data as Record<string, unknown>;
        if (!Array.isArray(data) && (anyData.openapi || anyData.swagger)) {
             return [{
                name: (anyData.info as { title: string })?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                openapiService: {
                    address: activeTab === "url" ? url : undefined,
                    specUrl: activeTab === "url" ? url : undefined,
                    specContent: activeTab === "json" ? jsonContent : (activeTab === "file" && file ? (await file.text()) : undefined),
                    tools: [], resources: [], prompts: [], calls: {}
                }
             } as UpstreamServiceConfig];
        }

        // Normalize
        if (Array.isArray(data)) return data as UpstreamServiceConfig[];
        if (anyData.services && Array.isArray(anyData.services)) return anyData.services as UpstreamServiceConfig[];
        // Single object
        return [data as UpstreamServiceConfig];
    };

    const handleValidateStart = async () => {
        setStep("validation");
        setParsedServices([]);
        setIsValidating(true);

        try {
            const rawServices = await parseContent();

            if (rawServices.length === 0) {
                throw new Error("No services found in configuration.");
            }

            const initialServices: ParsedService[] = rawServices.map((s, idx) => ({
                id: crypto.randomUUID(),
                originalConfig: s,
                name: s.name || `Service-${idx + 1}`,
                status: "pending",
                selected: true // Default select all
            }));

            setParsedServices(initialServices);

            // Run validation
            const results = await Promise.all(initialServices.map(async (s) => {
                try {
                    // Check if name is missing first
                    if (!s.originalConfig.name) {
                         return { ...s, status: "error", message: "Service name is required", selected: false };
                    }

                    const res = await apiClient.validateService(s.originalConfig);
                    if (res.valid) {
                         return { ...s, status: "valid", message: "Valid configuration" };
                    } else {
                         return {
                             ...s,
                             status: "error",
                             message: res.error || "Validation failed",
                             details: res.details,
                             selected: false // Auto-deselect invalid
                         };
                    }
                } catch (e: unknown) {
                    const errMsg = e instanceof Error ? e.message : "Validation error";
                    return {
                        ...s,
                        status: "error",
                        message: errMsg,
                        selected: false
                    };
                }
            }));

            setParsedServices(results);

        } catch (e: unknown) {
            const errMsg = e instanceof Error ? e.message : "Unknown error occurred";
            console.error(e);
            toast({
                title: "Import Error",
                description: errMsg,
                variant: "destructive"
            });
            setStep("input");
        } finally {
            setIsValidating(false);
        }
    };

    const handleImportStart = async () => {
        setStep("import");
        setIsImporting(true);
        setImportResults([]);
        setImportProgress(0);

        const selectedServices = parsedServices.filter(s => s.selected);
        const total = selectedServices.length;
        let completed = 0;

        const results = [];

        for (const service of selectedServices) {
            try {
                await apiClient.registerService(service.originalConfig);
                results.push({ name: service.name, success: true });
            } catch (e: unknown) {
                const errMsg = e instanceof Error ? e.message : "Unknown error";
                results.push({ name: service.name, success: false, error: errMsg });
            }
            completed++;
            setImportProgress(Math.round((completed / total) * 100));
        }

        setImportResults(results);
        setIsImporting(false);
        setStep("result");
    };

    const handleNext = () => {
        if (step === "input") {
            handleValidateStart();
        } else if (step === "validation") {
            // Ensure at least one selected
            if (parsedServices.filter(s => s.selected).length === 0) {
                toast({ title: "Selection Required", description: "Please select at least one service to import.", variant: "destructive" });
                return;
            }
            handleImportStart();
        } else if (step === "result") {
             onImportSuccess();
        }
    };

    const handleBack = () => {
        if (step === "validation") setStep("input");
        else if (step === "import") {
            // Cannot go back during import
        }
    };

    const toggleSelection = (id: string) => {
        setParsedServices(prev => prev.map(s => s.id === id ? { ...s, selected: !s.selected } : s));
    };

    const toggleSelectAll = (checked: boolean) => {
        setParsedServices(prev => prev.map(s => s.status === "valid" ? { ...s, selected: checked } : s));
    };

    const renderInputStep = () => (
        <div className="space-y-6 py-4">
            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="json" className="gap-2"><FileJson className="h-4 w-4" /> JSON / YAML</TabsTrigger>
                    <TabsTrigger value="url" className="gap-2"><LinkIcon className="h-4 w-4" /> URL</TabsTrigger>
                    <TabsTrigger value="file" className="gap-2"><Upload className="h-4 w-4" /> File Upload</TabsTrigger>
                </TabsList>

                <TabsContent value="json" className="pt-4 space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="json-content">Configuration Content</Label>
                        <Textarea
                            id="json-content"
                            placeholder='[{"name": "service-1", ...}]'
                            className="min-h-[300px] font-mono text-xs"
                            value={jsonContent}
                            onChange={(e) => setJsonContent(e.target.value)}
                        />
                        <p className="text-xs text-muted-foreground">Paste your service configuration here (JSON or YAML).</p>
                    </div>
                </TabsContent>

                <TabsContent value="url" className="pt-4 space-y-4">
                     <div className="space-y-2">
                        <Label htmlFor="config-url">Configuration URL</Label>
                        <div className="flex gap-2">
                            <Input
                                id="config-url"
                                placeholder="https://example.com/services.json"
                                value={url}
                                onChange={(e) => setUrl(e.target.value)}
                            />
                        </div>
                        <p className="text-xs text-muted-foreground">Enter a URL to fetch the configuration from.</p>
                     </div>
                </TabsContent>

                <TabsContent value="file" className="pt-4 space-y-4">
                    <div className="border-2 border-dashed rounded-lg p-12 flex flex-col items-center justify-center text-center hover:bg-muted/50 transition-colors relative cursor-pointer group">
                        <Input
                            type="file"
                            className="absolute inset-0 opacity-0 cursor-pointer z-10"
                            accept=".json,.yaml,.yml"
                            onChange={(e) => setFile(e.target.files?.[0] || null)}
                        />
                        <div className="bg-primary/10 p-4 rounded-full mb-4 group-hover:scale-110 transition-transform">
                            <Upload className="h-8 w-8 text-primary" />
                        </div>
                        <h3 className="font-medium mb-1">{file ? file.name : "Click to upload or drag and drop"}</h3>
                        <p className="text-sm text-muted-foreground">
                            {file ? `${(file.size / 1024).toFixed(1)} KB` : "Supports JSON or YAML configuration files"}
                        </p>
                    </div>
                </TabsContent>
            </Tabs>

            <div className="flex justify-end gap-2 pt-2 border-t">
                <Button variant="ghost" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleNext} disabled={
                    (activeTab === "json" && !jsonContent.trim()) ||
                    (activeTab === "url" && !url.trim()) ||
                    (activeTab === "file" && !file)
                }>
                    Review & Validate <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
            </div>
        </div>
    );

    const renderValidationStep = () => {
        if (isValidating) {
             return (
                <div className="space-y-6 py-4">
                    <div className="flex flex-col items-center justify-center py-20 text-muted-foreground space-y-4">
                        <Loader2 className="h-10 w-10 animate-spin text-primary" />
                        <p>Validating configuration...</p>
                    </div>
                </div>
            );
        }

        const validCount = parsedServices.filter(s => s.status === "valid").length;
        const selectedCount = parsedServices.filter(s => s.selected).length;
        const hasValid = validCount > 0;

        return (
             <div className="space-y-4 py-4">
                <div className="flex items-center justify-between">
                    <div>
                        <h3 className="text-lg font-medium">Validation Results</h3>
                        <p className="text-sm text-muted-foreground">
                            Found {parsedServices.length} services. {validCount} valid.
                        </p>
                    </div>
                    <div className="text-sm text-muted-foreground">
                        {selectedCount} selected for import
                    </div>
                </div>

                <div className="border rounded-md">
                    <ScrollArea className="h-[300px]">
                        <Table>
                            <TableHeader className="bg-muted/50 sticky top-0">
                                <TableRow>
                                    <TableHead className="w-[50px]">
                                        <Checkbox
                                            checked={hasValid && selectedCount === validCount}
                                            onCheckedChange={(c) => toggleSelectAll(!!c)}
                                            disabled={!hasValid}
                                        />
                                    </TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Service Name</TableHead>
                                    <TableHead>Message</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {parsedServices.map((service) => (
                                    <TableRow key={service.id} className={service.status === "error" ? "bg-destructive/5" : ""}>
                                        <TableCell>
                                            <Checkbox
                                                checked={service.selected}
                                                onCheckedChange={() => toggleSelection(service.id)}
                                                disabled={service.status === "error"}
                                            />
                                        </TableCell>
                                        <TableCell>
                                            {service.status === "valid" && (
                                                <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-200 gap-1">
                                                    <Check className="h-3 w-3" /> Valid
                                                </Badge>
                                            )}
                                            {service.status === "error" && (
                                                <TooltipProvider>
                                                    <Tooltip>
                                                        <TooltipTrigger>
                                                            <Badge variant="destructive" className="gap-1 cursor-help">
                                                                <AlertCircle className="h-3 w-3" /> Error
                                                            </Badge>
                                                        </TooltipTrigger>
                                                        <TooltipContent>
                                                            <p>{service.details || service.message}</p>
                                                        </TooltipContent>
                                                    </Tooltip>
                                                </TooltipProvider>
                                            )}
                                        </TableCell>
                                        <TableCell className="font-medium">{service.name}</TableCell>
                                        <TableCell className="text-sm text-muted-foreground max-w-[300px] truncate" title={service.message}>
                                            {service.message}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </ScrollArea>
                </div>

                <div className="flex justify-between pt-2 border-t">
                    <Button variant="ghost" onClick={handleBack}><ArrowLeft className="mr-2 h-4 w-4" /> Back</Button>
                    <Button onClick={handleNext} disabled={selectedCount === 0}>
                        Import {selectedCount} Services <ArrowRight className="ml-2 h-4 w-4" />
                    </Button>
                </div>
            </div>
        );
    };

    const renderImportStep = () => (
         <div className="space-y-6 py-4">
             <div className="space-y-4">
                <div className="flex items-center justify-between">
                    <h3 className="font-medium">
                        {isImporting ? "Importing Services..." : "Import Complete"}
                    </h3>
                    <span className="text-sm text-muted-foreground">{importProgress}%</span>
                </div>
                <Progress value={importProgress} className="h-2" />
                <p className="text-center text-sm text-muted-foreground">
                    Please wait while we register your services.
                </p>
             </div>
        </div>
    );

    const renderResultStep = () => {
        const successCount = importResults.filter(r => r.success).length;
        const failedCount = importResults.filter(r => !r.success).length;

        return (
             <div className="space-y-6 py-4">
                <div className="text-center space-y-2">
                    {failedCount === 0 ? (
                        <div className="bg-green-100 dark:bg-green-900/20 p-4 rounded-full w-fit mx-auto">
                            <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
                        </div>
                    ) : (
                        <div className="bg-orange-100 dark:bg-orange-900/20 p-4 rounded-full w-fit mx-auto">
                            <AlertTriangle className="h-8 w-8 text-orange-600 dark:text-orange-400" />
                        </div>
                    )}
                    <h3 className="text-xl font-semibold">Import Complete</h3>
                    <p className="text-muted-foreground">
                        Successfully imported {successCount} services.
                        {failedCount > 0 && ` Failed to import ${failedCount} services.`}
                    </p>
                </div>

                {failedCount > 0 && (
                    <div className="border rounded-md">
                        <ScrollArea className="h-[200px]">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Service</TableHead>
                                        <TableHead>Status</TableHead>
                                        <TableHead>Error</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {importResults.filter(r => !r.success).map((res) => (
                                        <TableRow key={res.name}>
                                            <TableCell className="font-medium">{res.name}</TableCell>
                                            <TableCell><Badge variant="destructive">Failed</Badge></TableCell>
                                            <TableCell className="text-sm text-muted-foreground">{res.error}</TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </ScrollArea>
                    </div>
                )}

                <div className="flex justify-center pt-4">
                    <Button onClick={onImportSuccess} className="w-full sm:w-auto min-w-[120px]">
                        Close
                    </Button>
                </div>
            </div>
        );
    };

    return (
        <div className="w-full">
            {step === "input" && renderInputStep()}
            {step === "validation" && renderValidationStep()}
            {step === "import" && renderImportStep()}
            {step === "result" && renderResultStep()}
        </div>
    );
}
