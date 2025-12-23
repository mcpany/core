
"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Settings, Plus, Wrench } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/components/ui/sheet"
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface Tool {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  service_id: string;
}

export default function ToolsPage() {
  const [tools, setTools] = useState<Tool[]>([]);
  const [selectedTool, setSelectedTool] = useState<Tool | null>(null);

  useEffect(() => {
    fetch("/api/tools")
      .then(res => res.json())
      .then(setTools);
  }, []);

  const toggleTool = async (id: string) => {
    const tool = tools.find(t => t.id === id);
    if (!tool) return;

    // Optimistic update
    setTools(tools.map(t => t.id === id ? { ...t, enabled: !t.enabled } : t));

    try {
        await fetch("/api/tools", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id, enabled: !tool.enabled })
        });
    } catch (e) {
        console.error("Failed to toggle tool", e);
        setTools(tools); // Revert
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
        <Button disabled>
            <Plus className="mr-2 h-4 w-4" /> Add Tool
        </Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
        <CardHeader>
          <CardTitle>Registered Tools</CardTitle>
          <CardDescription>Manage tools exposed by your services.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service ID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tools.map((tool) => (
                <TableRow key={tool.id}>
                  <TableCell className="font-medium flex items-center">
                      <Wrench className="mr-2 h-4 w-4 text-muted-foreground" />
                      {tool.name}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{tool.description}</TableCell>
                  <TableCell>
                      <Badge variant="outline">{tool.service_id}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={tool.enabled}
                            onCheckedChange={() => toggleTool(tool.id)}
                        />
                        <span className="text-sm text-muted-foreground min-w-[60px]">
                            {tool.enabled ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                       <Sheet>
                           <SheetTrigger asChild>
                                <Button variant="ghost" size="icon" onClick={() => setSelectedTool(tool)}>
                                   <Settings className="h-4 w-4" />
                               </Button>
                           </SheetTrigger>
                           <SheetContent>
                               <SheetHeader>
                                   <SheetTitle>Tool Configuration</SheetTitle>
                                   <SheetDescription>
                                       View details for {tool.name}
                                   </SheetDescription>
                               </SheetHeader>
                               {selectedTool && (
                                <div className="grid gap-4 py-4">
                                   <div className="grid grid-cols-4 items-center gap-4">
                                       <Label className="text-right">Name</Label>
                                       <Input value={selectedTool.name} disabled className="col-span-3" />
                                   </div>
                                    <div className="grid grid-cols-4 items-center gap-4">
                                       <Label className="text-right">Service</Label>
                                       <Input value={selectedTool.service_id} disabled className="col-span-3" />
                                   </div>
                               </div>
                               )}
                           </SheetContent>
                       </Sheet>
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
