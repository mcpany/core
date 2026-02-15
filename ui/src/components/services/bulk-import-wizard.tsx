/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CheckCircle2, Upload, FileText, Globe, AlertCircle, RefreshCw, AlertTriangle, PlayCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { Checkbox } from "@/components/ui/checkbox";
import { Progress } from "@/components/ui/progress";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type WizardStep = 'input' | 'validation' | 'selection' | 'import';

/**
 * BulkImportWizard guides the user through importing multiple services.
 */
export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [currentStep, setCurrentStep] = useState<WizardStep>('input');
    const [jsonContent, setJsonContent] = useState("");
    const [importUrl, setImportUrl] = useState("");
    const [parsedServices, setParsedServices] = useState<UpstreamServiceConfig[]>([]);
    const [inputError, setInputError] = useState<string | null>(null);

    // Validation State
    const [validationResults, setValidationResults] = useState<Map<string, { valid: boolean, error?: string, details?: string }>>(new Map());
    const [validating, setValidating] = useState(false);

    // Selection State
    const [selectedServices, setSelectedServices] = useState<Set<string>>(new Set());

    // Import State
    const [importing, setImporting] = useState(false);
    const [importProgress, setImportProgress] = useState<{current: number, total: number, success: number, failed: number}>({ current: 0, total: 0, success: 0, failed: 0 });
    const [importResults, setImportResults] = useState<Map<string, { success: boolean, error?: string }>>(new Map());


    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            setJsonContent(event.target?.result as string);
            setInputError(null);
        };
        reader.readAsText(file);
    };

    const handleParse = async () => {
        setInputError(null);
        try {
            let services: UpstreamServiceConfig[] = [];

            if (importUrl.trim()) {
                 const res = await fetch(importUrl);
                 if (!res.ok) throw new Error("Failed to fetch from URL.");
                 const data = await res.json();
                 services = Array.isArray(data) ? data : (data.services || [data]);
            } else if (jsonContent.trim()) {
                const data = JSON.parse(jsonContent);
                services = Array.isArray(data) ? data : (data.services || [data]);
            } else {
                throw new Error("Please provide JSON content or a URL.");
            }

            if (!services || services.length === 0) {
                 throw new Error("No services found in input.");
            }

            // Basic check if it looks like a service config
            services = services.filter(s => s && (s.name || s.id));

            if (services.length === 0) {
                 throw new Error("No valid service configurations found.");
            }

            setParsedServices(services);
            setValidationResults(new Map()); // Reset validation
            setSelectedServices(new Set(services.map(s => s.name))); // Select all by default
            setCurrentStep('validation');
        } catch (e: any) {
            console.error("Parse error", e);
            setInputError(e.message || "Failed to parse input.");
        }
    };

    useEffect(() => {
        if (currentStep === 'validation' && parsedServices.length > 0 && !validating && validationResults.size === 0) {
            validateAll();
        }
    }, [currentStep, parsedServices]);

    const validateAll = async () => {
        setValidating(true);
        const results = new Map();

        await Promise.all(parsedServices.map(async (service) => {
             try {
                 const res = await apiClient.validateService(service);
                 results.set(service.name, res);
             } catch (e: any) {
                 results.set(service.name, { valid: false, error: e.message });
             }
        }));

        setValidationResults(results);
        setValidating(false);
    };

    const handleImport = async () => {
        setImporting(true);
        setCurrentStep('import');
        setImportResults(new Map());

        const toImport = parsedServices.filter(s => selectedServices.has(s.name));
        setImportProgress({ current: 0, total: toImport.length, success: 0, failed: 0 });

        for (let i = 0; i < toImport.length; i++) {
            const service = toImport[i];
            try {
                await apiClient.registerService(service);
                setImportResults(prev => new Map(prev).set(service.name, { success: true }));
                setImportProgress(prev => ({ ...prev, current: i + 1, success: prev.success + 1 }));
            } catch (e: any) {
                setImportResults(prev => new Map(prev).set(service.name, { success: false, error: e.message }));
                setImportProgress(prev => ({ ...prev, current: i + 1, failed: prev.failed + 1 }));
            }
        }
        setImporting(false);
    };

    const getServiceType = (service: UpstreamServiceConfig) => {
        if (service.httpService) return "HTTP";
        if (service.grpcService) return "gRPC";
        if (service.commandLineService) return "CLI";
        if (service.mcpService) return "MCP";
        return "Unknown";
    }

    const toggleSelection = (name: string, checked: boolean) => {
        const newSet = new Set(selectedServices);
        if (checked) newSet.add(name);
        else newSet.delete(name);
        setSelectedServices(newSet);
    };

    const toggleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelectedServices(new Set(parsedServices.map(s => s.name)));
        } else {
            setSelectedServices(new Set());
        }
    };

    return (
        <div className="flex flex-col h-[600px]">
            {/* Stepper Header */}
            <div className="flex items-center justify-between mb-6 px-2">
                <StepIndicator step="input" current={currentStep} label="Source" index={1} />
                <StepConnector active={currentStep !== 'input'} />
                <StepIndicator step="validation" current={currentStep} label="Validate" index={2} />
                <StepConnector active={currentStep === 'selection' || currentStep === 'import'} />
                <StepIndicator step="selection" current={currentStep} label="Select" index={3} />
                <StepConnector active={currentStep === 'import'} />
                <StepIndicator step="import" current={currentStep} label="Import" index={4} />
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto min-h-0 border rounded-md bg-muted/10 p-4">
                {currentStep === 'input' && (
                    <div className="space-y-4">
                        <div className="text-sm text-muted-foreground mb-4">
                            Choose how you want to import services. You can upload a JSON file, paste JSON content, or provide a URL to a configuration file.
                        </div>
                        <Tabs defaultValue="paste" className="w-full" onValueChange={() => {
                            setJsonContent("");
                            setImportUrl("");
                            setInputError(null);
                        }}>
                            <TabsList className="grid w-full grid-cols-3">
                                <TabsTrigger value="paste">
                                    <FileText className="mr-2 h-4 w-4" /> Paste JSON
                                </TabsTrigger>
                                <TabsTrigger value="upload">
                                    <Upload className="mr-2 h-4 w-4" /> Upload File
                                </TabsTrigger>
                                <TabsTrigger value="url">
                                    <Globe className="mr-2 h-4 w-4" /> From URL
                                </TabsTrigger>
                            </TabsList>
                            <TabsContent value="paste" className="mt-4 space-y-4">
                                <Textarea
                                    placeholder='[{"name": "service1", ...}]'
                                    className="h-64 font-mono text-xs"
                                    value={jsonContent}
                                    onChange={(e) => setJsonContent(e.target.value)}
                                />
                            </TabsContent>
                            <TabsContent value="upload" className="mt-4 space-y-4">
                                <div className="border-2 border-dashed rounded-lg p-12 flex flex-col items-center justify-center text-center hover:bg-muted/50 transition-colors">
                                     <Upload className="h-8 w-8 text-muted-foreground mb-4" />
                                     <Label htmlFor="file-upload" className="cursor-pointer">
                                        <span className="bg-primary text-primary-foreground px-4 py-2 rounded-md hover:bg-primary/90">Select JSON File</span>
                                        <Input
                                            id="file-upload"
                                            type="file"
                                            accept=".json"
                                            className="hidden"
                                            onChange={handleFileUpload}
                                        />
                                     </Label>
                                     {jsonContent && <div className="mt-4 text-xs font-mono text-muted-foreground truncate max-w-xs">File loaded ({jsonContent.length} bytes)</div>}
                                </div>
                            </TabsContent>
                             <TabsContent value="url" className="mt-4 space-y-4">
                                <div className="space-y-2">
                                    <Label>Configuration URL</Label>
                                    <Input
                                        placeholder="https://example.com/mcp-config.json"
                                        value={importUrl}
                                        onChange={(e) => setImportUrl(e.target.value)}
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
                    </div>
                )}

                {currentStep === 'validation' && (
                    <div className="space-y-4">
                        <div className="flex justify-between items-center">
                            <h3 className="text-lg font-medium">Validation Results</h3>
                             <Button size="sm" variant="outline" onClick={validateAll} disabled={validating}>
                                {validating ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
                                Re-validate
                            </Button>
                        </div>
                        <div className="border rounded-md">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Status</TableHead>
                                        <TableHead>Service Name</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead>Message</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {parsedServices.map((service) => {
                                        const result = validationResults.get(service.name);
                                        return (
                                            <TableRow key={service.name}>
                                                <TableCell>
                                                    {validating || !result ? (
                                                        <div className="h-4 w-4 rounded-full border-2 border-muted border-t-primary animate-spin" />
                                                    ) : result.valid ? (
                                                        <CheckCircle2 className="h-5 w-5 text-green-500" data-testid="validation-status-valid" />
                                                    ) : (
                                                        <AlertTriangle className="h-5 w-5 text-destructive" data-testid="validation-status-error" />
                                                    )}
                                                </TableCell>
                                                <TableCell className="font-medium">{service.name}</TableCell>
                                                <TableCell>{getServiceType(service)}</TableCell>
                                                <TableCell className="text-sm text-muted-foreground">
                                                    {result?.error || result?.details || "Valid"}
                                                </TableCell>
                                            </TableRow>
                                        );
                                    })}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                )}

                {currentStep === 'selection' && (
                     <div className="space-y-4">
                        <div className="flex justify-between items-center">
                            <h3 className="text-lg font-medium">Select Services to Import</h3>
                            <div className="text-sm text-muted-foreground">
                                {selectedServices.size} of {parsedServices.length} selected
                            </div>
                        </div>
                        <div className="border rounded-md">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-[50px]">
                                             <Checkbox
                                                checked={parsedServices.length > 0 && selectedServices.size === parsedServices.length}
                                                onCheckedChange={(checked) => toggleSelectAll(!!checked)}
                                             />
                                        </TableHead>
                                        <TableHead>Status</TableHead>
                                        <TableHead>Service Name</TableHead>
                                        <TableHead>Type</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {parsedServices.map((service) => {
                                        const result = validationResults.get(service.name);
                                        return (
                                            <TableRow key={service.name}>
                                                 <TableCell>
                                                    <Checkbox
                                                        checked={selectedServices.has(service.name)}
                                                        onCheckedChange={(checked) => toggleSelection(service.name, !!checked)}
                                                    />
                                                </TableCell>
                                                <TableCell>
                                                    {result?.valid ? (
                                                        <CheckCircle2 className="h-5 w-5 text-green-500" />
                                                    ) : (
                                                        <AlertTriangle className="h-5 w-5 text-destructive" />
                                                    )}
                                                </TableCell>
                                                <TableCell className="font-medium">{service.name}</TableCell>
                                                <TableCell>{getServiceType(service)}</TableCell>
                                            </TableRow>
                                        );
                                    })}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                )}

                {currentStep === 'import' && (
                    <div className="space-y-6 flex flex-col items-center justify-center h-full">
                        <div className="w-full max-w-md space-y-4">
                            <h3 className="text-lg font-medium text-center">
                                {importing ? "Importing Services..." : "Import Complete"}
                            </h3>
                            <Progress value={(importProgress.current / importProgress.total) * 100} className="w-full" />
                            <div className="flex justify-between text-sm text-muted-foreground">
                                <span>Total: {importProgress.total}</span>
                                <span className="text-green-600">Success: {importProgress.success}</span>
                                <span className="text-destructive">Failed: {importProgress.failed}</span>
                            </div>

                            {!importing && (
                                <div className="mt-4 border rounded-md max-h-48 overflow-y-auto text-sm p-2 bg-muted/20">
                                    {Array.from(importResults.entries()).map(([name, res]) => (
                                        <div key={name} className="flex items-center gap-2 py-1">
                                            {res.success ? <CheckCircle2 className="h-4 w-4 text-green-500" /> : <AlertTriangle className="h-4 w-4 text-destructive" />}
                                            <span className="font-medium">{name}:</span>
                                            <span>{res.success ? "Success" : res.error}</span>
                                        </div>
                                    ))}
                                </div>
                            )}

                             {!importing && (
                                <Button className="w-full mt-4" onClick={onImportSuccess}>
                                    Finish
                                </Button>
                            )}
                        </div>
                    </div>
                )}
            </div>

            {/* Footer Actions */}
            <div className="flex justify-between mt-6 pt-2 border-t">
                <Button variant="ghost" onClick={onCancel} disabled={importing}>
                    {currentStep === 'import' && !importing ? "Close" : "Cancel"}
                </Button>
                <div className="flex gap-2">
                    {currentStep !== 'input' && currentStep !== 'import' && (
                        <Button variant="outline" onClick={() => {
                             if (currentStep === 'validation') setCurrentStep('input');
                             if (currentStep === 'selection') setCurrentStep('validation');
                        }}>
                            Back
                        </Button>
                    )}

                    {currentStep === 'input' && (
                        <Button onClick={handleParse} disabled={!jsonContent && !importUrl}>
                            Next: Validate
                        </Button>
                    )}
                    {currentStep === 'validation' && (
                        <Button onClick={() => setCurrentStep('selection')}>
                            Next: Select
                        </Button>
                    )}
                    {currentStep === 'selection' && (
                         <Button onClick={handleImport} disabled={selectedServices.size === 0}>
                            Import {selectedServices.size} Services
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
}

function StepIndicator({ step, current, label, index }: { step: WizardStep, current: WizardStep, label: string, index: number }) {
    const steps: WizardStep[] = ['input', 'validation', 'selection', 'import'];
    const currentIndex = steps.indexOf(current);
    const stepIndex = steps.indexOf(step);

    const isActive = step === current;
    const isCompleted = stepIndex < currentIndex;
    const isPending = stepIndex > currentIndex;

    return (
        <div className="flex flex-col items-center gap-1">
            <div className={`
                w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-colors duration-200
                ${isActive ? 'bg-primary text-primary-foreground ring-2 ring-offset-2 ring-primary' : ''}
                ${isCompleted ? 'bg-primary text-primary-foreground' : ''}
                ${isPending ? 'bg-muted text-muted-foreground' : ''}
            `}>
                {isCompleted ? <CheckCircle2 className="h-5 w-5" /> : index}
            </div>
            <span className={`text-xs font-medium ${isActive ? 'text-foreground' : 'text-muted-foreground'}`}>
                {label}
            </span>
        </div>
    );
}

function StepConnector({ active }: { active: boolean }) {
    return (
        <div className={`flex-1 h-0.5 mx-2 ${active ? 'bg-primary' : 'bg-muted'}`} />
    );
}
