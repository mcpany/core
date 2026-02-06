/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect, useRef } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ScrollArea } from "@/components/ui/scroll-area";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2, CheckCircle2, XCircle, Terminal } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface AutoCraftWizardProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onComplete: (config: any) => void;
}

interface LogEntry {
    timestamp: string;
    message: string;
}

/**
 * AutoCraftWizard component.
 * Allows users to generate MCP server configurations using AI.
 *
 * @param props - Component properties.
 * @param props.open - Whether the wizard is open.
 * @param props.onOpenChange - Callback when open state changes.
 * @param props.onComplete - Callback when configuration is generated.
 */
export function AutoCraftWizard({ open, onOpenChange, onComplete }: AutoCraftWizardProps) {
    const { toast } = useToast();
    const [step, setStep] = useState<'input' | 'processing' | 'review'>('input');

    // Input State
    const [goal, setGoal] = useState("");
    const [serviceName, setServiceName] = useState("");
    const [providerType, setProviderType] = useState("openai");
    const [isSubmitting, setIsSubmitting] = useState(false);

    // Processing State
    const [jobId, setJobId] = useState<string | null>(null);
    const [status, setStatus] = useState<"in_progress" | "success" | "failed">("in_progress");
    const [logs, setLogs] = useState<string[]>([]);
    const [generatedConfig, setGeneratedConfig] = useState<any>(null);
    const [errorMessage, setErrorMessage] = useState("");
    const pollingRef = useRef<NodeJS.Timeout | null>(null);

    // Cleanup polling on unmount or close
    useEffect(() => {
        return () => {
            if (pollingRef.current) clearInterval(pollingRef.current);
        };
    }, []);

    const startPolling = (id: string) => {
        if (pollingRef.current) clearInterval(pollingRef.current);

        pollingRef.current = setInterval(async () => {
            try {
                const result = await apiClient.getAutoCraftJob(id);
                if (result) {
                    if (result.logs) setLogs(result.logs);
                    if (result.status === 'success') {
                        setGeneratedConfig(JSON.parse(result.configJson));
                        setStatus('success');
                        setStep('review');
                        if (pollingRef.current) clearInterval(pollingRef.current);
                    } else if (result.status === 'failed') {
                        setStatus('failed');
                        setErrorMessage(result.errorMessage || "Unknown error");
                        if (pollingRef.current) clearInterval(pollingRef.current);
                    }
                }
            } catch (e) {
                console.error("Polling error", e);
            }
        }, 1000);
    };

    const handleSubmit = async () => {
        if (!goal || !serviceName) {
            toast({ title: "Validation Error", description: "Goal and Service Name are required.", variant: "destructive" });
            return;
        }

        setIsSubmitting(true);
        try {
            const res = await apiClient.submitAutoCraftJob({
                serviceName,
                providerType,
                goal
            });
            setJobId(res.jobId);
            setStep('processing');
            setStatus('in_progress');
            setLogs(["Job submitted...", "Waiting for worker..."]);
            startPolling(res.jobId);
        } catch (e) {
            toast({ title: "Submission Failed", description: String(e), variant: "destructive" });
            setIsSubmitting(false);
        }
    };

    const handleSave = async () => {
         onComplete(generatedConfig);
         onOpenChange(false);
    };

    const reset = () => {
        setStep('input');
        setGoal("");
        setServiceName("");
        setJobId(null);
        setLogs([]);
        setGeneratedConfig(null);
        setStatus("in_progress");
        setIsSubmitting(false);
        if (pollingRef.current) clearInterval(pollingRef.current);
    };

    const logsEndRef = useRef<HTMLDivElement>(null);
    useEffect(() => {
        logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [logs]);

    return (
        <Dialog open={open} onOpenChange={(val) => {
            if (!val) reset();
            onOpenChange(val);
        }}>
            <DialogContent className="sm:max-w-[700px] flex flex-col h-[80vh]">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        {step === 'input' && "Auto Craft MCP Server"}
                        {step === 'processing' && "Crafting Server..."}
                        {step === 'review' && "Review Generated Config"}
                    </DialogTitle>
                    <DialogDescription>
                        {step === 'input' && "Describe your desired MCP server, and AI will generate the configuration for you."}
                        {step === 'processing' && "AI is researching and generating your configuration."}
                        {step === 'review' && "Review the generated configuration before saving."}
                    </DialogDescription>
                </DialogHeader>

                <div className="flex-1 overflow-y-auto py-4 px-1">
                    {step === 'input' && (
                        <div className="space-y-6">
                            <div className="grid gap-2">
                                <Label htmlFor="serviceName">Service Name</Label>
                                <Input
                                    id="serviceName"
                                    placeholder="e.g., weather-service"
                                    value={serviceName}
                                    onChange={(e) => setServiceName(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground">Unique identifier for this service.</p>
                            </div>

                            <div className="grid gap-2">
                                <Label>AI Provider</Label>
                                <Select value={providerType} onValueChange={setProviderType}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select Provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="openai">OpenAI (GPT-4)</SelectItem>
                                        <SelectItem value="anthropic">Anthropic (Claude 3)</SelectItem>
                                        <SelectItem value="gemini">Google Gemini</SelectItem>
                                    </SelectContent>
                                </Select>
                                <p className="text-xs text-muted-foreground">The AI model used to generate the config.</p>
                            </div>

                            <div className="grid gap-2">
                                <Label htmlFor="goal">Goal / Description</Label>
                                <Textarea
                                    id="goal"
                                    placeholder="e.g., I want an MCP server that wraps the OpenWeatherMap API to provide current weather data. It should have a tool 'get_weather' taking 'city' as argument."
                                    className="h-32"
                                    value={goal}
                                    onChange={(e) => setGoal(e.target.value)}
                                />
                                <p className="text-xs text-muted-foreground">Be as specific as possible about tools and functionality.</p>
                            </div>
                        </div>
                    )}

                    {step === 'processing' && (
                        <div className="flex flex-col h-full gap-4">
                            <div className="flex items-center gap-4 p-4 border rounded-lg bg-muted/50">
                                {status === 'in_progress' && <Loader2 className="h-8 w-8 animate-spin text-primary" />}
                                {status === 'failed' && <XCircle className="h-8 w-8 text-destructive" />}
                                {status === 'success' && <CheckCircle2 className="h-8 w-8 text-green-500" />}

                                <div className="flex-1">
                                    <h3 className="font-semibold">
                                        {status === 'in_progress' ? "AI Agent Working" : (status === 'success' ? "Generation Complete" : "Generation Failed")}
                                    </h3>
                                    <p className="text-sm text-muted-foreground">
                                        {status === 'in_progress' ? "Researching and generating configuration..." : (status === 'success' ? "Configuration ready for review." : errorMessage)}
                                    </p>
                                </div>
                            </div>

                            <div className="flex-1 border rounded-md bg-black p-4 font-mono text-xs overflow-hidden flex flex-col">
                                <div className="flex items-center gap-2 text-muted-foreground mb-2 border-b border-gray-800 pb-2">
                                    <Terminal className="h-3 w-3" />
                                    <span>Agent Logs</span>
                                </div>
                                <ScrollArea className="flex-1">
                                    <div className="space-y-1">
                                        {logs.map((log, i) => (
                                            <div key={i} className="text-green-400">
                                                <span className="opacity-50 mr-2">{new Date().toLocaleTimeString()}</span>
                                                {log}
                                            </div>
                                        ))}
                                        <div ref={logsEndRef} />
                                    </div>
                                </ScrollArea>
                            </div>
                        </div>
                    )}

                    {step === 'review' && generatedConfig && (
                        <div className="space-y-4">
                            <div className="grid gap-2">
                                <Label>Generated Configuration</Label>
                                <div className="h-[300px] border rounded-md overflow-hidden">
                                    <ScrollArea className="h-full bg-muted/30 p-4">
                                        <pre className="text-xs font-mono">
                                            {JSON.stringify(generatedConfig, null, 2)}
                                        </pre>
                                    </ScrollArea>
                                </div>
                            </div>
                            <div className="flex items-center gap-2 p-3 bg-blue-50 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200 rounded-md text-sm">
                                <CheckCircle2 className="h-4 w-4" />
                                This configuration has been validated and is ready to use.
                            </div>
                        </div>
                    )}
                </div>

                <DialogFooter className="mt-4">
                    {step === 'input' && (
                        <Button onClick={handleSubmit} disabled={isSubmitting}>
                            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Start Auto Craft
                        </Button>
                    )}
                    {step === 'processing' && status === 'failed' && (
                        <Button onClick={reset} variant="outline">Retry</Button>
                    )}
                    {step === 'review' && (
                        <>
                            <Button variant="outline" onClick={reset}>Discard</Button>
                            <Button onClick={handleSave}>Save Template</Button>
                        </>
                    )}
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
