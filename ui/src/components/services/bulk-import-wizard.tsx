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
import { Progress } from "@/components/ui/progress";
import { Check, Loader2, Upload, CheckCircle2, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import yaml from 'js-yaml';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";

interface BulkImportWizardProps {
    onImportSuccess: () => void;
    onCancel: () => void;
}

type Step = "input" | "validation" | "import" | "result";
type InputType = "paste" | "url" | "file";
type ValidationStatus = "pending" | "valid" | "error";

interface ServiceImportItem {
    id: string; // temp id for list
    config: UpstreamServiceConfig;
    status: ValidationStatus;
    message?: string;
    selected: boolean;
}

/**
 * A wizard component for bulk importing services with validation and selection.
 */
export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
    const [step, setStep] = useState<Step>("input");
    const [inputType, setInputType] = useState<InputType>("paste");
    const [rawContent, setRawContent] = useState("");
    const [url, setUrl] = useState("");
    const [importItems, setImportItems] = useState<ServiceImportItem[]>([]);
    const [isValidating, setIsValidating] = useState(false);
    const [isImporting, setIsImporting] = useState(false);
    const [importProgress, setImportProgress] = useState(0);
    const [importResult, setImportResult] = useState<{success: number, failure: number}>({ success: 0, failure: 0 });

    const { toast } = useToast();

    const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (event) => {
                if (event.target?.result) {
                    setRawContent(event.target.result as string);
                    setInputType("paste"); // Switch to paste view to confirm
                }
            };
            reader.readAsText(file);
        }
    };

    const validateContent = async () => {
        setIsValidating(true);
        try {
            let parsed: unknown;
            if (inputType === "url") {
                const res = await fetch(url);
                if (!res.ok) throw new Error(`Failed to fetch URL: ${res.statusText}`);
                const text = await res.text();
                // Try JSON then YAML
                try {
                    parsed = JSON.parse(text);
                } catch {
                    parsed = yaml.load(text);
                }
            } else {
                 try {
                    parsed = JSON.parse(rawContent);
                } catch {
                    parsed = yaml.load(rawContent);
                }
            }

            // Normalize to array
            // Handle { services: [...] } or [...] or single object
            let configs: unknown[] = [];
            if (Array.isArray(parsed)) {
                configs = parsed;
            } else if (parsed && typeof parsed === 'object' && 'services' in parsed && Array.isArray((parsed as {services: unknown[]}).services)) {
                configs = (parsed as {services: unknown[]}).services;
            } else if (parsed) {
                configs = [parsed];
            }

            // Filter out empty
            configs = configs.filter(c => c && typeof c === 'object');

            if (configs.length === 0) throw new Error("No valid services found in content.");

            const items: ServiceImportItem[] = configs.map((c: unknown, i: number) => ({
                id: `item-${i}`,
                config: c as UpstreamServiceConfig,
                status: "pending",
                selected: true
            }));

            setImportItems(items);
            setStep("validation");

            // Validate each sequentially to avoid overwhelming server? Or parallel?
            // Parallel is fine for validation.
            await Promise.all(items.map(async (item) => {
                try {
                    // Ensure name is present for validation context
                    if (!item.config.name) {
                         setImportItems(prev => prev.map(p => p.id === item.id ? {
                            ...p,
                            status: "error",
                            message: "Service name is required",
                            selected: false
                        } : p));
                        return;
                    }

                    const res = await apiClient.validateService(item.config);
                    setImportItems(prev => prev.map(p => p.id === item.id ? {
                        ...p,
                        status: res.valid ? "valid" : "error",
                        message: res.error || (res.valid ? "Valid" : "Invalid configuration"),
                        selected: res.valid // Auto-select valid ones
                    } : p));
                } catch (e) {
                     const errorMsg = e instanceof Error ? e.message : "Validation failed";
                     setImportItems(prev => prev.map(p => p.id === item.id ? {
                        ...p,
                        status: "error",
                        message: errorMsg,
                        selected: false
                    } : p));
                }
            }));

        } catch (e) {
            const errorMsg = e instanceof Error ? e.message : "Parse Error";
            toast({ variant: "destructive", title: "Parse Error", description: errorMsg });
        } finally {
            setIsValidating(false);
        }
    };

    const handleImport = async () => {
        setIsImporting(true);
        setStep("import");
        setImportProgress(0);

        const selected = importItems.filter(i => i.selected);
        let successCount = 0;
        let failureCount = 0;

        for (let i = 0; i < selected.length; i++) {
            const item = selected[i];
            try {
                await apiClient.registerService(item.config);
                successCount++;
                // Update status in list (though not visible in progress step usually)
            } catch (e) {
                failureCount++;
                const errorMsg = e instanceof Error ? e.message : "Import failed";
                console.error(`Failed to import ${item.config.name}`, errorMsg);
                // Could store per-item error for result step
            }
            setImportProgress(Math.round(((i + 1) / selected.length) * 100));
        }

        setImportResult({ success: successCount, failure: failureCount });
        setStep("result");
        setIsImporting(false);
        if (successCount > 0) {
             onImportSuccess(); // Trigger refresh in parent
        }
    };

    const handleToggleSelect = (id: string, checked: boolean) => {
        setImportItems(prev => prev.map(p => p.id === id ? { ...p, selected: checked } : p));
    };

    const handleSelectAll = (checked: boolean) => {
         setImportItems(prev => prev.map(p => p.status === "valid" ? { ...p, selected: checked } : p));
    };

    return (
        <div className="space-y-6">
             {/* Progress Indicator */}
             <div className="flex justify-between text-xs text-muted-foreground mb-4 border-b pb-2">
                 <span className={cn(step === "input" && "font-bold text-primary")}>1. Source</span>
                 <span className={cn(step === "validation" && "font-bold text-primary")}>2. Validate & Select</span>
                 <span className={cn((step === "import" || step === "result") && "font-bold text-primary")}>3. Import</span>
             </div>

             {step === "input" && (
                 <div className="space-y-4">
                     <Tabs value={inputType} onValueChange={(v) => setInputType(v as InputType)}>
                         <TabsList className="grid w-full grid-cols-3">
                             <TabsTrigger value="paste">Paste JSON/YAML</TabsTrigger>
                             <TabsTrigger value="url">From URL</TabsTrigger>
                             <TabsTrigger value="file">Upload File</TabsTrigger>
                         </TabsList>

                         <TabsContent value="paste" className="mt-4">
                             <div className="space-y-2">
                                 <Label htmlFor="paste-area">Configuration Content</Label>
                                 <Textarea
                                     id="paste-area"
                                     placeholder='[{"name": "service-1", ...}]'
                                     className="font-mono text-xs h-64 resize-none"
                                     value={rawContent}
                                     onChange={(e) => setRawContent(e.target.value)}
                                 />
                                 <p className="text-[10px] text-muted-foreground">Supports JSON array or YAML list of services.</p>
                             </div>
                         </TabsContent>

                         <TabsContent value="url" className="mt-4">
                             <div className="space-y-4 py-8">
                                 <div className="space-y-2">
                                     <Label htmlFor="url-input">Configuration URL</Label>
                                     <div className="flex gap-2">
                                         <Input
                                             id="url-input"
                                             placeholder="https://example.com/config.json"
                                             value={url}
                                             onChange={(e) => setUrl(e.target.value)}
                                         />
                                     </div>
                                     <p className="text-[10px] text-muted-foreground">URL must be reachable by the server.</p>
                                 </div>
                             </div>
                         </TabsContent>

                         <TabsContent value="file" className="mt-4">
                             <div
                                className="border-2 border-dashed rounded-lg h-64 flex flex-col items-center justify-center text-muted-foreground gap-4 hover:bg-muted/50 transition-colors cursor-pointer relative"
                                onClick={() => document.getElementById("file-upload")?.click()}
                             >
                                 <Upload className="h-10 w-10 opacity-50" />
                                 <div className="text-center">
                                     <p className="font-medium text-sm">Click to upload or drag and drop</p>
                                     <p className="text-xs opacity-70">JSON or YAML files supported</p>
                                 </div>
                                 <Input
                                     type="file"
                                     className="hidden"
                                     id="file-upload"
                                     accept=".json,.yaml,.yml"
                                     onChange={handleFileUpload}
                                 />
                             </div>
                         </TabsContent>
                     </Tabs>

                     <div className="flex justify-end gap-2 pt-4">
                         <Button variant="ghost" onClick={onCancel}>Cancel</Button>
                         <Button onClick={validateContent} disabled={(inputType === "url" ? !url : !rawContent) || isValidating}>
                             {isValidating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                             Next: Validate
                         </Button>
                     </div>
                 </div>
             )}

             {step === "validation" && (
                 <div className="space-y-4">
                     <div className="border rounded-md max-h-[400px] overflow-y-auto">
                         <Table>
                             <TableHeader>
                                 <TableRow>
                                     <TableHead className="w-[50px]">
                                         <Checkbox
                                            checked={importItems.some(i => i.selected) && importItems.every(i => i.status !== "valid" || i.selected)}
                                            onCheckedChange={(c) => handleSelectAll(!!c)}
                                         />
                                     </TableHead>
                                     <TableHead>Status</TableHead>
                                     <TableHead>Service Name</TableHead>
                                     <TableHead>Type</TableHead>
                                     <TableHead>Message</TableHead>
                                 </TableRow>
                             </TableHeader>
                             <TableBody>
                                 {importItems.map((item) => (
                                     <TableRow key={item.id} className={item.status === "error" ? "bg-destructive/5" : ""}>
                                         <TableCell>
                                             <Checkbox
                                                checked={item.selected}
                                                onCheckedChange={(c) => handleToggleSelect(item.id, !!c)}
                                                disabled={item.status === "error"}
                                             />
                                         </TableCell>
                                         <TableCell>
                                             {item.status === "pending" && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                                             {item.status === "valid" && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                                             {item.status === "error" && <XCircle className="h-4 w-4 text-red-500" />}
                                         </TableCell>
                                         <TableCell className="font-medium">{item.config.name || "Unnamed"}</TableCell>
                                         <TableCell>
                                             <Badge variant="outline" className="text-[10px]">
                                                {item.config.httpService ? "HTTP" :
                                                 item.config.grpcService ? "gRPC" :
                                                 item.config.commandLineService ? "CLI" :
                                                 item.config.mcpService ? "MCP" : "Unknown"}
                                             </Badge>
                                         </TableCell>
                                         <TableCell className="text-xs text-muted-foreground">{item.message}</TableCell>
                                     </TableRow>
                                 ))}
                             </TableBody>
                         </Table>
                     </div>

                     <div className="flex justify-between items-center pt-4">
                         <div className="text-sm text-muted-foreground">
                             {importItems.filter(i => i.selected).length} services selected
                         </div>
                         <div className="flex gap-2">
                            <Button variant="ghost" onClick={() => setStep("input")}>Back</Button>
                            <Button onClick={handleImport} disabled={importItems.filter(i => i.selected).length === 0 || isImporting}>
                                {isImporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                                Import Selected
                            </Button>
                         </div>
                     </div>
                 </div>
             )}

             {step === "import" && (
                 <div className="flex flex-col items-center justify-center py-12 space-y-4">
                     <Loader2 className="h-8 w-8 animate-spin text-primary" />
                     <h3 className="font-medium">Importing Services...</h3>
                     <Progress value={importProgress} className="w-[60%]" />
                     <p className="text-xs text-muted-foreground">{importProgress}% Complete</p>
                 </div>
             )}

             {step === "result" && (
                 <div className="flex flex-col items-center justify-center py-8 space-y-6">
                     <div className="flex items-center justify-center h-16 w-16 rounded-full bg-green-100 dark:bg-green-900/20">
                         <Check className="h-8 w-8 text-green-600 dark:text-green-500" />
                     </div>
                     <div className="text-center space-y-2">
                         <h3 className="text-xl font-bold">Import Complete</h3>
                         <p className="text-muted-foreground">
                             Successfully imported {importResult.success} services.
                             {importResult.failure > 0 && ` Failed to import ${importResult.failure} services.`}
                         </p>
                     </div>
                     <Button onClick={onCancel}>Close</Button>
                 </div>
             )}
        </div>
    );
}
