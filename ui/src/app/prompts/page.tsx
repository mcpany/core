
"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Settings, Plus, MessageSquare } from "lucide-react";

interface Prompt {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
}

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<Prompt[]>([]);

  useEffect(() => {
    fetch("/api/prompts")
      .then(res => res.json())
      .then(setPrompts);
  }, []);

  const togglePrompt = async (id: string) => {
    const prompt = prompts.find(p => p.id === id);
    if (!prompt) return;

    setPrompts(prompts.map(p => p.id === id ? { ...p, enabled: !p.enabled } : p));

    try {
        await fetch("/api/prompts", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id, enabled: !prompt.enabled })
        });
    } catch (e) {
        console.error("Failed to toggle prompt", e);
        setPrompts(prompts); // Revert
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
        <Button disabled><Plus className="mr-2 h-4 w-4" /> Add Prompt</Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
        <CardHeader>
          <CardTitle>Prompt Library</CardTitle>
          <CardDescription>Manage your AI prompts.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {prompts.map((prompt) => (
                <TableRow key={prompt.id}>
                  <TableCell className="font-medium flex items-center">
                      <MessageSquare className="mr-2 h-4 w-4 text-muted-foreground" />
                      {prompt.name}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{prompt.description}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={prompt.enabled}
                            onCheckedChange={() => togglePrompt(prompt.id)}
                        />
                         <span className="text-sm text-muted-foreground min-w-[60px]">
                            {prompt.enabled ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                      <Button variant="ghost" size="icon">
                          <Settings className="h-4 w-4" />
                      </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
