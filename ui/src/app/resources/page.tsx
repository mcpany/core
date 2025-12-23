
"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Settings, Plus, Database, FileText } from "lucide-react";

interface Resource {
  id: string;
  name: string;
  uri: string;
  mime_type: string;
  enabled: boolean;
}

export default function ResourcesPage() {
  const [resources, setResources] = useState<Resource[]>([]);

  useEffect(() => {
    fetch("/api/resources")
      .then(res => res.json())
      .then(setResources);
  }, []);

  const toggleResource = async (id: string) => {
    const resource = resources.find(r => r.id === id);
    if (!resource) return;

    setResources(resources.map(r => r.id === id ? { ...r, enabled: !r.enabled } : r));

    try {
        await fetch("/api/resources", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id, enabled: !resource.enabled })
        });
    } catch (e) {
        console.error("Failed to toggle resource", e);
        setResources(resources); // Revert
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
        <Button disabled><Plus className="mr-2 h-4 w-4" /> Add Resource</Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
        <CardHeader>
          <CardTitle>Available Resources</CardTitle>
          <CardDescription>Manage static and dynamic resources.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>URI</TableHead>
                <TableHead>MIME Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {resources.map((res) => (
                <TableRow key={res.id}>
                  <TableCell className="font-medium flex items-center">
                      <FileText className="mr-2 h-4 w-4 text-muted-foreground" />
                      {res.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{res.uri}</TableCell>
                  <TableCell>
                      <Badge variant="secondary">{res.mime_type}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={res.enabled}
                            onCheckedChange={() => toggleResource(res.id)}
                        />
                        <span className="text-sm text-muted-foreground min-w-[60px]">
                            {res.enabled ? "Enabled" : "Disabled"}
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
