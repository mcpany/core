/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import { Card, CardContent } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Upload,
  Link as LinkIcon,
  Code,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Loader2,
  ArrowRight,
  ArrowLeft,
} from "lucide-react";
import { apiClient } from "@/lib/client";

interface BulkImportWizardProps {
  onImportSuccess: () => void;
  onCancel: () => void;
}

type Step = "input" | "validate" | "import" | "result";

export function BulkImportWizard({ onImportSuccess, onCancel }: BulkImportWizardProps) {
  const [step, setStep] = useState<Step>("input");

  // Input State
  const [activeTab, setActiveTab] = useState("json");
  const [jsonContent, setJsonContent] = useState("");
  const [importUrl, setImportUrl] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [parseError, setParseError] = useState<string | null>(null);

  // Validation State
  const [validating, setValidating] = useState(false);
  const [validationResults, setValidationResults] = useState<any[]>([]);
  const [selectedServices, setSelectedServices] = useState<Set<number>>(new Set());

  // Import State
  const [importing, setImporting] = useState(false);
  const [importProgress, setImportProgress] = useState(0);
  const [importResults, setImportResults] = useState<any[]>([]);

  // Handlers for Navigation
  const handleNext = async () => {
    if (step === "input") {
      setParseError(null);
      setValidating(true);
      try {
        const services = await parseInput();
        if (services.length === 0) {
           setParseError("No services found in the input.");
           setValidating(false);
           return;
        }

        // Initialize validation
        const initialResults = services.map((s, idx) => ({
            id: idx,
            config: s,
            status: 'pending',
            message: 'Validating...'
        }));
        setValidationResults(initialResults);
        setSelectedServices(new Set());
        setStep("validate");

        // Run validation asynchronously
        validateServices(initialResults);

      } catch (e: any) {
        setParseError(e.message || "Failed to parse input.");
      } finally {
        setValidating(false);
      }
    } else if (step === "validate") {
      setStep("import");
      const selected = validationResults.filter(r => selectedServices.has(r.id));
      await importServices(selected);
    }
  };

  const importServices = async (items: any[]) => {
      setImporting(true);
      setImportProgress(0);
      const results = [];

      let completed = 0;
      for (const item of items) {
          try {
              await apiClient.registerService(item.config);
              results.push({ ...item, status: 'success', message: 'Imported successfully' });
          } catch (e: any) {
              results.push({ ...item, status: 'error', message: e.message || 'Import failed' });
          }
          completed++;
          setImportProgress(Math.round((completed / items.length) * 100));
      }

      setImportResults(results);
      setImporting(false);
      setStep("result");
  };

  const validateServices = async (items: any[]) => {
      for (const item of items) {
          try {
              const res = await apiClient.validateService(item.config);

              setValidationResults(prev => prev.map(p =>
                  p.id === item.id
                    ? {
                        ...p,
                        status: res.valid ? 'valid' : 'error',
                        message: res.valid ? 'Valid configuration' : (res.error || 'Invalid configuration'),
                        details: res.details
                      }
                    : p
              ));

              if (res.valid) {
                  setSelectedServices(prev => {
                      const newSet = new Set(prev);
                      newSet.add(item.id);
                      return newSet;
                  });
              }
          } catch (e: any) {
               setValidationResults(prev => prev.map(p =>
                  p.id === item.id
                    ? { ...p, status: 'error', message: e.message || 'Validation failed' }
                    : p
              ));
          }
      }
  };

  const handleBack = () => {
    if (step === "validate") setStep("input");
    if (step === "import") setStep("validate");
  };

  const parseInput = async (): Promise<any[]> => {
    let content = "";
    if (activeTab === "json") {
      content = jsonContent;
    } else if (activeTab === "file" && file) {
      content = await file.text();
    } else if (activeTab === "url" && importUrl) {
      const res = await fetch(importUrl);
      if (!res.ok) throw new Error("Failed to fetch from URL");
      content = await res.text();
    }

    try {
      const data = JSON.parse(content);

      // Handle OpenAPI/Swagger Spec
      if (data.openapi || data.swagger) {
          const serviceName = data.info?.title?.toLowerCase().replace(/\s+/g, '-') || "openapi-service";
          return [{
              name: serviceName,
              openapiService: {
                  address: importUrl || "",
                  specUrl: importUrl || "",
                  specContent: !importUrl ? content : undefined,
                  tools: [], resources: [], calls: {}, prompts: []
              }
          }];
      }

      // Handle array or single object or { services: [...] } wrapper
      let services = [];
      if (Array.isArray(data)) {
        services = data;
      } else if (data.services && Array.isArray(data.services)) {
        services = data.services;
      } else {
        services = [data];
      }
      return services;
    } catch (e) {
      throw new Error("Invalid JSON content");
    }
  };

  // Render Steps
  const renderInputStep = () => (
    <div className="space-y-4">
      <Tabs defaultValue="json" value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="json"><Code className="mr-2 h-4 w-4" /> JSON / Paste</TabsTrigger>
          <TabsTrigger value="file"><Upload className="mr-2 h-4 w-4" /> File Upload</TabsTrigger>
          <TabsTrigger value="url"><LinkIcon className="mr-2 h-4 w-4" /> URL Import</TabsTrigger>
        </TabsList>

        <TabsContent value="json" className="space-y-4 pt-4">
          <div className="space-y-2">
            <Label>Paste JSON Configuration</Label>
            <Textarea
              placeholder='[{"name": "service1", ...}]'
              className="h-64 font-mono text-xs"
              value={jsonContent}
              onChange={(e) => setJsonContent(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Paste an array of service configurations or a single service object.
            </p>
          </div>
        </TabsContent>

        <TabsContent value="file" className="space-y-4 pt-4">
          <div className="border-2 border-dashed rounded-lg p-10 flex flex-col items-center justify-center space-y-4 hover:bg-muted/50 transition-colors">
            <Upload className="h-10 w-10 text-muted-foreground" />
            <div className="text-center space-y-1">
              <p className="font-medium text-sm">Drag and drop your JSON file here</p>
              <p className="text-xs text-muted-foreground">Or click to browse</p>
            </div>
            <Input
              type="file"
              accept=".json"
              className="max-w-xs"
              onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f) setFile(f);
              }}
            />
            {file && (
                <div className="flex items-center gap-2 text-sm text-green-600 bg-green-50 px-3 py-1 rounded-full border border-green-200">
                    <CheckCircle2 className="h-4 w-4" />
                    {file.name}
                </div>
            )}
          </div>
        </TabsContent>

        <TabsContent value="url" className="space-y-4 pt-4">
          <div className="space-y-2">
            <Label>Import from URL</Label>
            <div className="flex gap-2">
              <Input
                placeholder="https://example.com/mcp-config.json"
                value={importUrl}
                onChange={(e) => setImportUrl(e.target.value)}
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Provide a URL to a JSON configuration file or OpenAPI spec.
            </p>
          </div>
        </TabsContent>
      </Tabs>

      {parseError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{parseError}</AlertDescription>
        </Alert>
      )}

      <div className="flex justify-end gap-2 pt-4">
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={handleNext} disabled={
            (activeTab === "json" && !jsonContent) ||
            (activeTab === "file" && !file) ||
            (activeTab === "url" && !importUrl) ||
            validating
        }>
          {validating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
          Next <ArrowRight className="ml-2 h-4 w-4" />
        </Button>
      </div>
    </div>
  );

  const renderValidateStep = () => (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium">Review & Validate</h3>
        <span className="text-sm text-muted-foreground">
             {selectedServices.size} selected
        </span>
      </div>

      <div className="border rounded-md h-64 overflow-y-auto bg-muted/5">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]">
                <Checkbox
                    checked={validationResults.some(r => r.status === 'valid') && selectedServices.size === validationResults.filter(r => r.status === 'valid').length}
                    onCheckedChange={(checked) => {
                        if (checked) {
                            setSelectedServices(new Set(validationResults.filter(r => r.status === 'valid').map(r => r.id)));
                        } else {
                            setSelectedServices(new Set());
                        }
                    }}
                    disabled={!validationResults.some(r => r.status === 'valid')}
                />
              </TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Service Name</TableHead>
              <TableHead>Message</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {validationResults.map((result) => (
              <TableRow key={result.id}>
                <TableCell>
                  <Checkbox
                    checked={selectedServices.has(result.id)}
                    onCheckedChange={(checked) => {
                        setSelectedServices(prev => {
                            const next = new Set(prev);
                            if (checked) next.add(result.id);
                            else next.delete(result.id);
                            return next;
                        });
                    }}
                    disabled={result.status !== 'valid'}
                  />
                </TableCell>
                <TableCell>
                    {result.status === 'pending' && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                    {result.status === 'valid' && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                    {result.status === 'error' && (
                        <div className="group relative">
                            <AlertCircle className="h-4 w-4 text-destructive cursor-help" />
                        </div>
                    )}
                </TableCell>
                <TableCell className="font-medium">{result.config.name || "Unnamed Service"}</TableCell>
                <TableCell className="text-xs text-muted-foreground">
                    {result.message}
                    {result.details && <div className="text-[10px] opacity-75">{result.details}</div>}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex justify-between pt-4">
        <Button variant="ghost" onClick={handleBack}>
           <ArrowLeft className="mr-2 h-4 w-4" /> Back
        </Button>
        <div className="flex gap-2">
            <Button variant="outline" onClick={onCancel}>Cancel</Button>
            <Button onClick={handleNext} disabled={selectedServices.size === 0}>
            Import Selected
            </Button>
        </div>
      </div>
    </div>
  );

  const renderImportStep = () => (
    <div className="space-y-8 py-8">
      <div className="space-y-2 text-center">
        <h3 className="text-lg font-medium">Importing Services...</h3>
        <p className="text-sm text-muted-foreground">Please wait while we register your services.</p>
      </div>

      <div className="space-y-2 max-w-md mx-auto">
        <Progress value={importProgress} className="h-2" />
        <p className="text-xs text-center text-muted-foreground">{importProgress}% complete</p>
      </div>
    </div>
  );

  const renderResultStep = () => (
    <div className="space-y-6 py-4">
      <div className="flex flex-col items-center justify-center text-center space-y-2">
        <div className="h-12 w-12 rounded-full bg-green-100 flex items-center justify-center text-green-600 mb-2">
          <CheckCircle2 className="h-6 w-6" />
        </div>
        <h3 className="text-xl font-medium">Import Complete</h3>
        <p className="text-muted-foreground">
          Successfully processed {importResults.length} services.
        </p>
      </div>

      <div className="border rounded-md max-h-48 overflow-y-auto bg-muted/5">
        <Table>
          <TableBody>
            {importResults.map((result, idx) => (
              <TableRow key={idx}>
                <TableCell className="w-[30px]">
                    {result.status === 'success' ? <CheckCircle2 className="h-4 w-4 text-green-500" /> : <AlertTriangle className="h-4 w-4 text-destructive" />}
                </TableCell>
                <TableCell className="font-medium">{result.config.name}</TableCell>
                <TableCell className="text-xs text-muted-foreground">{result.message}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex justify-center pt-4">
        <Button onClick={onImportSuccess} className="min-w-[120px]">
          Done
        </Button>
      </div>
    </div>
  );

  return (
    <div className="w-full max-w-2xl mx-auto">
      {step === "input" && renderInputStep()}
      {step === "validate" && renderValidateStep()}
      {step === "import" && renderImportStep()}
      {step === "result" && renderResultStep()}
    </div>
  );
}
