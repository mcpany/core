import { useState } from "react";
import { PromptDefinition, apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Loader2, Play, Sparkles, Copy, Edit } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";

interface PromptRunnerProps {
    prompt: PromptDefinition;
    onEdit: () => void;
}

export function PromptRunner({ prompt, onEdit }: PromptRunnerProps) {
    const [args, setArgs] = useState<Record<string, string>>({});
    const [result, setResult] = useState<any | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const getArguments = () => {
        if (!prompt.inputSchema || !prompt.inputSchema.properties) return [];
        const props = prompt.inputSchema.properties as Record<string, any>;
        const required = (prompt.inputSchema.required as string[]) || [];
        return Object.entries(props).map(([key, value]) => ({
            name: key,
            description: value.description,
            required: required.includes(key),
            type: value.type
        }));
    };

    const handleExecute = async () => {
        setLoading(true);
        try {
            const res = await apiClient.executePrompt(prompt.name, args);
            setResult(res);
            toast({ title: "Executed", description: "Prompt executed successfully." });
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to execute prompt.", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    const copyToClipboard = () => {
        if (result) {
            navigator.clipboard.writeText(JSON.stringify(result, null, 2));
            toast({ title: "Copied", description: "Result copied to clipboard." });
        }
    };

    return (
        <div className="flex flex-col h-full bg-background">
            <div className="p-6 border-b bg-background shrink-0">
                <div className="flex items-start justify-between">
                    <div>
                        <h2 className="text-2xl font-bold tracking-tight">{prompt.name}</h2>
                        <p className="text-muted-foreground mt-1">{prompt.description || "No description provided."}</p>
                    </div>
                    <Button variant="outline" onClick={onEdit}>
                        <Edit className="mr-2 h-4 w-4" /> Edit Definition
                    </Button>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 h-full">
                    {/* Arguments */}
                    <div className="flex flex-col gap-6">
                        <div>
                            <h3 className="text-sm font-medium mb-4 flex items-center gap-2 text-primary">
                                Configuration
                            </h3>
                            <Card>
                                <CardContent className="p-4 space-y-4">
                                    {getArguments().length > 0 ? (
                                        getArguments().map((arg) => (
                                            <div key={arg.name} className="space-y-1.5">
                                                <Label htmlFor={arg.name} className="flex items-center gap-1 text-xs font-mono uppercase text-muted-foreground">
                                                    {arg.name}
                                                    {arg.required && <span className="text-red-500">*</span>}
                                                </Label>
                                                <Input
                                                    id={arg.name}
                                                    placeholder={arg.description}
                                                    value={args[arg.name] || ""}
                                                    onChange={(e) => setArgs({ ...args, [arg.name]: e.target.value })}
                                                />
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-sm text-muted-foreground italic text-center py-4">
                                            No arguments required.
                                        </div>
                                    )}
                                    <Button
                                        className="w-full mt-4"
                                        onClick={handleExecute}
                                        disabled={loading}
                                    >
                                        {loading ? (
                                            <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Generating...</>
                                        ) : (
                                            <><Play className="mr-2 h-4 w-4" /> Generate Messages</>
                                        )}
                                    </Button>
                                </CardContent>
                            </Card>
                        </div>
                    </div>

                    {/* Result */}
                    <div className="flex flex-col h-full min-h-[400px]">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                <Sparkles className="h-4 w-4" /> Output Preview
                            </h3>
                            <div className="flex items-center gap-2">
                                <Button variant="ghost" size="sm" onClick={copyToClipboard} disabled={!result}>
                                    <Copy className="h-3 w-3" />
                                </Button>
                            </div>
                        </div>
                        <Card className="flex-1 flex flex-col overflow-hidden bg-muted/30 border-dashed">
                            <CardContent className="flex-1 p-0 overflow-auto">
                                {result ? (
                                    <div className="p-4 space-y-4">
                                        {(result.messages || []).map((msg: any, idx: number) => (
                                            <div key={idx} className="space-y-1">
                                                <div className="text-[10px] font-mono uppercase text-muted-foreground flex items-center gap-2">
                                                    <span className={cn(
                                                        "w-2 h-2 rounded-full",
                                                        msg.role === "user" ? "bg-blue-500" : "bg-green-500"
                                                    )} />
                                                    {msg.role}
                                                </div>
                                                <div className="bg-background border rounded-md p-3 text-sm whitespace-pre-wrap font-mono">
                                                    {msg.content?.text || JSON.stringify(msg.content)}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="flex flex-col items-center justify-center h-full text-muted-foreground text-sm p-8 text-center">
                                        <Sparkles className="h-10 w-10 opacity-20 mb-3" />
                                        <p>Configure arguments and click Generate to see the prompt result.</p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
