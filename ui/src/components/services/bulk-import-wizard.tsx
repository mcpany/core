/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Upload, FileJson, Link, Check, AlertCircle, Loader2, XCircle, CheckCircle2, AlertTriangle, ArrowRight, ArrowLeft } from "lucide-react";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Progress } from "@/components/ui/progress";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

/**
 * Steps in the bulk import wizard.
 */
export enum WizardStep {
    INPUT = "input",
    VALIDATION = "validation",
    SELECTION = "selection",
    IMPORT = "import",
    RESULT = "result"
}

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

interface ServiceValidationResult {
    config: UpstreamServiceConfig;
    isValid: boolean;
    error?: string;
    details?: string;
}

interface ImportResult {
    name: string;
    success: boolean;
    error?: string;
}

/**
 * BulkImportWizard component.
 * Allows users to import multiple services from a JSON config or URL with validation and selection.
 * @param props - Component props.
 * @returns The BulkImportWizard component.
 */
export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [step, setStep] = useState<WizardStep>(WizardStep.INPUT);
    const [inputType, setInputType] = useState<"json" | "file" | "url">("json");

    // Input State
    const [jsonContent, setJsonContent] = useState("");
    const [url, setUrl] = useState("");
    const [fileName, setFileName] = useState("");

    // Process State
    const [loading, setLoading] = useState(false);
    const [validationResults, setValidationResults] = useState<ServiceValidationResult[]>([]);
    const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set());
    const [importProgress, setImportProgress] = useState(0);
    const [importResults, setImportResults] = useState<ImportResult[]>([]);

    const { toast } = useToast();

    const parseInput = async (): Promise<UpstreamServiceConfig[]> => {
        if (inputType === 'url') {
            const res = await fetch(url);
            if (!res.ok) throw new Error(`Failed to fetch from URL: ${res.statusText}`);
            const data = await res.json();
             // Handle OpenAPI logic here if needed, consistent with legacy import
            if (data.openapi || data.swagger) {
                 return [{
                    name: data.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service",
                    version: data.info?.version || "1.0.0",
                    openapiService: {
                        address: url,
                        specUrl: url,
                        tools: [], resources: [], calls: {}, prompts: []
                    }
                } as any];
            }
            return Array.isArray(data) ? data : (data.services || [data]);
        }

        // File or JSON content
        if (!jsonContent.trim()) throw new Error("Content is empty");
        const data = JSON.parse(jsonContent);
        return Array.isArray(data) ? data : (data.services || [data]);
    };

    const handleNext = async () => {
        if (step === WizardStep.INPUT) {
            setLoading(true);
            try {
                const services = await parseInput();
                if (services.length === 0) throw new Error("No services found in input.");

                // Run Validation
                const results = await Promise.all(services.map(async (svc) => {
                     try {
                         // Ensure name is present for UI even if invalid
                         if (!svc.name) (svc as any).name = "unnamed-service";

                         const res = await apiClient.validateService(svc);
                         return { config: svc, isValid: res.valid, error: res.error, details: res.details };
                     } catch (e: any) {
                         return { config: svc, isValid: false, error: e.message };
                     }
                }));

                setValidationResults(results);
                // Pre-select valid ones
                const validIndices = results.map((r, i) => r.isValid ? i : -1).filter(i => i !== -1);
                setSelectedIndices(new Set(validIndices));

                setStep(WizardStep.VALIDATION);
            } catch (e: any) {
                toast({
                    variant: "destructive",
                    title: "Parsing Failed",
                    description: e.message
                });
            } finally {
                setLoading(false);
            }
        } else if (step === WizardStep.VALIDATION) {
             if (selectedIndices.size === 0) {
                 toast({
                     variant: "destructive",
                     title: "Selection Required",
                     description: "Please select at least one service to import."
                 });
                 return;
             }
            setStep(WizardStep.SELECTION); // Skip logic for now if validation and selection are merged?
            // Actually, let's merge Validation display and Selection into one step for better UX
            // Or keep them separate if we want "Review Validation" -> "Select subset".
            // Let's treat Step 2 as "Validation Report & Selection".
            setStep(WizardStep.IMPORT); // Jump to Import if we treat step 2 as selection too
            startImport();
        }
    };

    const startImport = async () => {
        setStep(WizardStep.IMPORT);
        setLoading(true);
        setImportProgress(0);
        const results: ImportResult[] = [];
        const indices = Array.from(selectedIndices);

        for (let i = 0; i < indices.length; i++) {
            const index = indices[i];
            const item = validationResults[index];
            try {
                await apiClient.registerService(item.config);
                results.push({ name: item.config.name, success: true });
            } catch (e: any) {
                results.push({ name: item.config.name, success: false, error: e.message });
            }
            setImportProgress(Math.round(((i + 1) / indices.length) * 100));
        }

        setImportResults(results);
        setLoading(false);
        setStep(WizardStep.RESULT);

        // Notify parent if all success?
        const successCount = results.filter(r => r.success).length;
        if (successCount > 0) {
             toast({
                title: "Import Completed",
                description: `Successfully imported ${successCount} services.`,
            });
            onImportSuccess(); // Refresh parent list
        }
    };

    const handleBack = () => {
        if (step === WizardStep.VALIDATION) setStep(WizardStep.INPUT);
        else if (step === WizardStep.RESULT) setStep(WizardStep.VALIDATION); // Allow going back to retry failed?
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            setFileName(file.name);
            const reader = new FileReader();
            reader.onload = (event) => {
                setJsonContent(event.target?.result as string);
            };
            reader.readAsText(file);
        }
    };

    const toggleSelection = (index: number) => {
        const newSet = new Set(selectedIndices);
        if (newSet.has(index)) newSet.delete(index);
        else newSet.add(index);
        setSelectedIndices(newSet);
    };

    const toggleAll = () => {
        if (selectedIndices.size === validationResults.length) {
            setSelectedIndices(new Set());
        } else {
            // Select only valid ones? Or all? User might want to try force import?
            // Usually only valid ones.
            const validIndices = validationResults.map((r, i) => r.isValid ? i : -1).filter(i => i !== -1);
            setSelectedIndices(new Set(validIndices));
        }
    };

    return (
        <div className="flex flex-col h-[500px]">
            {/* Stepper Header */}
            <div className="flex items-center justify-between mb-6 px-1 shrink-0">
                <div className="flex items-center gap-2 w-full">
                    <StepIndicator current={step} step={WizardStep.INPUT} label="Source" number={1} />
                    <StepSeparator />
                    <StepIndicator current={step} step={WizardStep.VALIDATION} label="Review" number={2} />
                    <StepSeparator />
                    <StepIndicator current={step} step={WizardStep.IMPORT} label="Import" number={3} />
                    <StepSeparator />
                    <StepIndicator current={step} step={WizardStep.RESULT} label="Done" number={4} />
                </div>
            </div>

            <div className="flex-1 overflow-hidden min-h-0 flex flex-col">
                {step === WizardStep.INPUT && (
                    <Tabs value={inputType} onValueChange={(v) => setInputType(v as any)} className="w-full flex-1 flex flex-col">
                        <TabsList className="grid w-full grid-cols-3 shrink-0">
                            <TabsTrigger value="json"><FileJson className="mr-2 h-4 w-4" /> Paste JSON</TabsTrigger>
                            <TabsTrigger value="file"><Upload className="mr-2 h-4 w-4" /> Upload File</TabsTrigger>
                            <TabsTrigger value="url"><Link className="mr-2 h-4 w-4" /> Import from URL</TabsTrigger>
                        </TabsList>

                        <div className="flex-1 mt-4 min-h-0 overflow-y-auto pr-1">
                            <TabsContent value="json" className="space-y-4 m-0 h-full flex flex-col">
                                <div className="space-y-2 flex-1 flex flex-col">
                                    <Label htmlFor="json-input">Service Configuration (JSON)</Label>
                                    <Textarea
                                        id="json-input"
                                        placeholder='[{"name": "my-service", ...}]'
                                        className="font-mono text-xs flex-1 resize-none"
                                        value={jsonContent}
                                        onChange={(e) => setJsonContent(e.target.value)}
                                    />
                                </div>
                            </TabsContent>
                            <TabsContent value="file" className="space-y-4 m-0">
                                <div
                                    className="border-2 border-dashed rounded-lg p-10 flex flex-col items-center justify-center text-center space-y-4 hover:bg-muted/50 transition-colors cursor-pointer"
                                    onClick={() => document.getElementById('file-upload')?.click()}
                                >
                                    <Upload className="h-10 w-10 text-muted-foreground" />
                                    <div>
                                        <p className="font-medium">Click to upload or drag and drop</p>
                                        <p className="text-xs text-muted-foreground mt-1">Supports .json configuration files</p>
                                    </div>
                                    <Input
                                        id="file-upload"
                                        type="file"
                                        accept=".json"
                                        className="hidden"
                                        onChange={handleFileChange}
                                    />
                                    {fileName && (
                                        <div className="flex items-center gap-2 text-sm bg-primary/10 text-primary px-3 py-1 rounded-full mt-2">
                                            <FileJson className="h-4 w-4" />
                                            {fileName}
                                        </div>
                                    )}
                                </div>
                            </TabsContent>
                            <TabsContent value="url" className="space-y-4 m-0">
                                <div className="space-y-2">
                                    <Label htmlFor="url-input">Configuration URL</Label>
                                    <Input
                                        id="url-input"
                                        placeholder="https://example.com/mcp-services.json"
                                        value={url}
                                        onChange={(e) => setUrl(e.target.value)}
                                    />
                                    <p className="text-xs text-muted-foreground">
                                        Enter a URL to a JSON configuration file.
                                    </p>
                                </div>
                            </TabsContent>
                        </div>
                    </Tabs>
                )}

                {step === WizardStep.VALIDATION && (
                    <div className="flex flex-col h-full gap-4">
                        <Alert>
                            <AlertCircle className="h-4 w-4" />
                            <AlertTitle>Review & Select</AlertTitle>
                            <AlertDescription>
                                Found {validationResults.length} services. {validationResults.filter(r => !r.isValid).length} have validation errors.
                            </AlertDescription>
                        </Alert>

                        <div className="rounded-md border flex-1 overflow-hidden">
                            <ScrollArea className="h-full">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead className="w-[50px]">
                                                <Checkbox
                                                    checked={selectedIndices.size > 0 && selectedIndices.size === validationResults.filter(r => r.isValid).length}
                                                    onCheckedChange={toggleAll}
                                                />
                                            </TableHead>
                                            <TableHead>Status</TableHead>
                                            <TableHead>Name</TableHead>
                                            <TableHead>Details</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {validationResults.map((result, index) => (
                                            <TableRow key={index} className={!result.isValid ? "bg-destructive/5" : ""}>
                                                <TableCell>
                                                    <Checkbox
                                                        checked={selectedIndices.has(index)}
                                                        onCheckedChange={() => toggleSelection(index)}
                                                        disabled={!result.isValid} // Force disable invalid?
                                                    />
                                                </TableCell>
                                                <TableCell>
                                                    {result.isValid ? (
                                                        <div className="flex items-center text-green-600 gap-1">
                                                            <CheckCircle2 className="h-4 w-4" />
                                                            <span className="text-xs font-medium">Valid</span>
                                                        </div>
                                                    ) : (
                                                        <div className="flex items-center text-destructive gap-1">
                                                            <XCircle className="h-4 w-4" />
                                                            <span className="text-xs font-medium">Error</span>
                                                        </div>
                                                    )}
                                                </TableCell>
                                                <TableCell className="font-medium">
                                                    {result.config.name}
                                                </TableCell>
                                                <TableCell className="text-xs text-muted-foreground">
                                                    {result.error ? (
                                                        <span className="text-destructive block max-w-[300px] truncate" title={result.error}>
                                                            {result.error}
                                                        </span>
                                                    ) : result.details ? (
                                                         <span className="text-yellow-600 block max-w-[300px] truncate" title={result.details}>
                                                            {result.details}
                                                        </span>
                                                    ) : (
                                                        "Ready to import"
                                                    )}
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </ScrollArea>
                        </div>
                    </div>
                )}

                {step === WizardStep.IMPORT && (
                     <div className="flex flex-col items-center justify-center h-full gap-6">
                        <div className="w-full max-w-md space-y-2">
                             <div className="flex justify-between text-sm">
                                <span>Importing services...</span>
                                <span>{Math.round(importProgress)}%</span>
                             </div>
                             <Progress value={importProgress} className="h-2" />
                        </div>
                        <p className="text-muted-foreground text-sm">Please wait while we register your services.</p>
                    </div>
                )}

                {step === WizardStep.RESULT && (
                    <div className="flex flex-col h-full gap-4">
                         <div className="flex items-center gap-2 p-4 bg-muted/20 rounded-lg">
                            {importResults.every(r => r.success) ? (
                                <CheckCircle2 className="h-8 w-8 text-green-500" />
                            ) : (
                                <AlertTriangle className="h-8 w-8 text-yellow-500" />
                            )}
                            <div>
                                <h3 className="font-medium text-lg">Import Complete</h3>
                                <p className="text-sm text-muted-foreground">
                                    {importResults.filter(r => r.success).length} successful, {importResults.filter(r => !r.success).length} failed.
                                </p>
                            </div>
                         </div>

                         <div className="rounded-md border flex-1 overflow-hidden">
                             <ScrollArea className="h-full">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Status</TableHead>
                                            <TableHead>Service</TableHead>
                                            <TableHead>Message</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {importResults.map((result, index) => (
                                            <TableRow key={index}>
                                                <TableCell>
                                                    {result.success ? (
                                                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                                                    ) : (
                                                        <XCircle className="h-4 w-4 text-destructive" />
                                                    )}
                                                </TableCell>
                                                <TableCell className="font-medium">{result.name}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">
                                                    {result.error || "Successfully registered"}
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                             </ScrollArea>
                         </div>
                    </div>
                )}
            </div>

            {/* Footer Actions */}
            <div className="flex justify-between mt-6 pt-4 border-t shrink-0">
                {step === WizardStep.RESULT ? (
                     <Button variant="outline" onClick={onCancel} className="w-full sm:w-auto">
                        Done
                    </Button>
                ) : (
                    <>
                        <Button variant="ghost" onClick={step === WizardStep.INPUT ? onCancel : handleBack} disabled={step === WizardStep.IMPORT || loading}>
                            {step === WizardStep.INPUT ? "Cancel" : <><ArrowLeft className="mr-2 h-4 w-4"/> Back</>}
                        </Button>

                        {step !== WizardStep.IMPORT && (
                            <Button onClick={handleNext} disabled={loading || (step === WizardStep.INPUT && !jsonContent && !url)} className="min-w-[100px]">
                                {loading ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                ) : step === WizardStep.INPUT ? (
                                    <>Review <ArrowRight className="ml-2 h-4 w-4"/></>
                                ) : (
                                    <>Import ({selectedIndices.size})</>
                                )}
                            </Button>
                        )}
                    </>
                )}
            </div>
        </div>
    );
}

function StepIndicator({ current, step, label, number }: { current: WizardStep, step: WizardStep, label: string, number: number }) {
    const steps = [WizardStep.INPUT, WizardStep.VALIDATION, WizardStep.IMPORT, WizardStep.RESULT];
    const currentIndex = steps.indexOf(current);
    const stepIndex = steps.indexOf(step);

    // Merge SELECTION into VALIDATION for display index
    // INPUT(0), VALIDATION(1), IMPORT(2), RESULT(3)

    const isActive = current === step;
    const isCompleted = currentIndex > stepIndex;

    return (
        <div className={`flex items-center gap-2 ${isActive ? "text-primary font-medium" : isCompleted ? "text-primary/80" : "text-muted-foreground"}`}>
            <div className={`flex items-center justify-center w-6 h-6 rounded-full text-xs border ${isActive ? "border-primary bg-primary text-primary-foreground" : isCompleted ? "border-primary bg-primary text-primary-foreground" : "border-muted-foreground"}`}>
                {isCompleted ? <Check className="h-3 w-3" /> : number}
            </div>
            <span className="text-sm hidden sm:inline">{label}</span>
        </div>
    );
}

function StepSeparator() {
    return <div className="h-px w-full min-w-[10px] bg-border hidden sm:block flex-1" />;
}
