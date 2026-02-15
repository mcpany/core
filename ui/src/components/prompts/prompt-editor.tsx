import { useState, useEffect } from "react";
import { PromptDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Save, Trash2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface PromptEditorProps {
    prompt: PromptDefinition | null; // null means new
    onSave: (prompt: PromptDefinition) => Promise<void>;
    onDelete: (name: string) => Promise<void>;
    onCancel: () => void;
}

export function PromptEditor({ prompt, onSave, onDelete, onCancel }: PromptEditorProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [messagesJson, setMessagesJson] = useState("[]");
    const [argsJson, setArgsJson] = useState("{}");
    const [saving, setSaving] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        if (prompt) {
            setName(prompt.name);
            setDescription(prompt.description || "");
            // Handle message format nuances (proto struct vs JSON)
            // The API returns messages as objects.
            setMessagesJson(JSON.stringify(prompt.messages || [], null, 2));
            setArgsJson(JSON.stringify(prompt.inputSchema || {}, null, 2));
        } else {
            setName("");
            setDescription("");
            setMessagesJson('[\n  {\n    "role": "user",\n    "content": {\n      "text": "Hello {{name}}!"\n    }\n  }\n]');
            setArgsJson('{\n  "type": "object",\n  "properties": {\n    "name": {\n      "type": "string"\n    }\n  }\n}');
        }
    }, [prompt]);

    const handleSave = async () => {
        if (!name) {
            toast({ title: "Validation Error", description: "Name is required", variant: "destructive" });
            return;
        }

        setSaving(true);
        try {
            let messages, inputSchema;
            try {
                messages = JSON.parse(messagesJson);
            } catch (e) {
                toast({ title: "Invalid Messages JSON", description: (e as Error).message, variant: "destructive" });
                setSaving(false);
                return;
            }
            try {
                inputSchema = JSON.parse(argsJson);
            } catch (e) {
                toast({ title: "Invalid Arguments JSON", description: (e as Error).message, variant: "destructive" });
                setSaving(false);
                return;
            }

            await onSave({
                name,
                description,
                messages,
                inputSchema,
                disable: prompt?.disable || false,
                profiles: prompt?.profiles || []
            });
        } catch (e) {
            console.error(e);
            toast({ title: "Save Failed", description: "Could not save prompt.", variant: "destructive" });
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full overflow-hidden bg-background">
            <div className="p-6 border-b bg-background shrink-0">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-2xl font-bold tracking-tight">
                        {prompt ? `Edit ${prompt.name}` : "New Prompt"}
                    </h2>
                    <div className="flex gap-2">
                        {prompt && (
                            <Button variant="destructive" size="sm" onClick={() => onDelete(prompt.name)}>
                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                            </Button>
                        )}
                        <Button variant="ghost" onClick={onCancel}>Cancel</Button>
                        <Button onClick={handleSave} disabled={saving}>
                            {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Save Prompt
                        </Button>
                    </div>
                </div>
                <div className="grid gap-4 py-4 max-w-2xl">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">Name</Label>
                        <Input
                            id="name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            disabled={!!prompt} // ID cannot change usually
                            className="col-span-3"
                            placeholder="my_prompt_name"
                        />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="description" className="text-right">Description</Label>
                        <Input
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="col-span-3"
                            placeholder="What does this prompt do?"
                        />
                    </div>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-6 space-y-6 bg-muted/10">
                <Card>
                    <CardHeader>
                        <CardTitle>Arguments Schema (JSON Schema)</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <Textarea
                            value={argsJson}
                            onChange={(e) => setArgsJson(e.target.value)}
                            className="font-mono text-xs min-h-[150px]"
                        />
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Message Template (JSON)</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-xs text-muted-foreground mb-2">
                            Define the messages using mustache syntax <code>{"{{variable}}"}</code> for substitution.
                        </div>
                        <Textarea
                            value={messagesJson}
                            onChange={(e) => setMessagesJson(e.target.value)}
                            className="font-mono text-xs min-h-[300px]"
                        />
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
