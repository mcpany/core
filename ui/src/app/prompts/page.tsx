
"use client";

import { useState, useEffect } from "react";
import { apiClient, PromptDefinition } from "@/lib/client";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Play } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function PromptsPage() {
    const [prompts, setPrompts] = useState<PromptDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
    const [promptArgs, setPromptArgs] = useState<Record<string, string>>({});
    const [executionResult, setExecutionResult] = useState<any>(null);

    useEffect(() => {
        apiClient.listPrompts().then(res => {
            setPrompts(Array.isArray(res) ? res : (res.prompts || []));
            setLoading(false);
        }).catch(err => {
            console.error("Failed to load prompts", err);
            setLoading(false);
        });
    }, []);

    const handleExecute = async () => {
        if (!selectedPrompt) return;
        setExecutionResult(null);
        try {
            const res = await apiClient.executePrompt(selectedPrompt.name, promptArgs);
            setExecutionResult(res);
        } catch (e: any) {
            setExecutionResult({ error: e.message });
        }
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
            </div>
            <GlassCard>
                <CardHeader>
                    <CardTitle>Prompt Templates</CardTitle>
                    <CardDescription>Reusable prompt templates.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Description</TableHead>
                                <TableHead>Arguments</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={4} className="text-center">Loading...</TableCell>
                                </TableRow>
                            ) : prompts.map((prompt) => (
                                <TableRow key={prompt.name}>
                                    <TableCell className="font-medium">{prompt.name}</TableCell>
                                    <TableCell>{prompt.description}</TableCell>
                                    <TableCell>
                                        <div className="flex gap-1 flex-wrap">
                                            {prompt.arguments?.map(arg => (
                                                <span key={arg.name} className="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80">
                                                    {arg.name} {arg.required ? "*" : ""}
                                                </span>
                                            ))}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Button size="sm" variant="outline" onClick={() => {
                                            setSelectedPrompt(prompt);
                                            setPromptArgs({});
                                            setExecutionResult(null);
                                        }}>
                                            <Play className="mr-2 h-3 w-3" /> Run
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </GlassCard>

             <Dialog open={!!selectedPrompt} onOpenChange={(open) => !open && setSelectedPrompt(null)}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>Run Prompt: {selectedPrompt?.name}</DialogTitle>
                        <DialogDescription>
                            {selectedPrompt?.description}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        {selectedPrompt?.arguments?.map(arg => (
                             <div key={arg.name} className="grid grid-cols-4 items-center gap-4">
                                <Label htmlFor={`arg-${arg.name}`} className="text-right">
                                    {arg.name} {arg.required && "*"}
                                </Label>
                                <Input
                                    id={`arg-${arg.name}`}
                                    value={promptArgs[arg.name] || ""}
                                    onChange={(e) => setPromptArgs({...promptArgs, [arg.name]: e.target.value})}
                                    className="col-span-3"
                                    placeholder={arg.description}
                                />
                            </div>
                        ))}

                        {executionResult && (
                             <div className="grid gap-2">
                                <Label>Result</Label>
                                <div className="rounded-md bg-muted p-4 font-mono text-xs whitespace-pre-wrap max-h-[300px] overflow-auto">
                                    {JSON.stringify(executionResult, null, 2)}
                                </div>
                            </div>
                        )}
                    </div>
                    <div className="flex justify-end">
                        <Button onClick={handleExecute}>Generate</Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
