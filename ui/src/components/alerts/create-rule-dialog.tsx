/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

/**
 * CreateRuleDialog component.
 * @returns The rendered component.
 */
export function CreateRuleDialog() {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const [formData, setFormData] = useState({
      name: "",
      severity: "warning",
      service: "all",
      condition: ""
  });

  const handleChange = (key: string, value: string) => {
      setFormData(prev => ({ ...prev, [key]: value }));
  };

  const handleSave = async () => {
    if (!formData.name || !formData.condition) {
        toast({
            title: "Validation Error",
            description: "Name and condition are required.",
            variant: "destructive"
        });
        return;
    }

    setLoading(true);
    try {
        await apiClient.createRule({
            name: formData.name,
            severity: formData.severity,
            service: formData.service === "all" ? "" : formData.service,
            condition: formData.condition,
            enabled: true
        });
        toast({
            title: "Rule Created",
            description: "Alert rule has been successfully created."
        });
        setOpen(false);
        setFormData({ name: "", severity: "warning", service: "all", condition: "" });
    } catch (error) {
        toast({
            title: "Error",
            description: error instanceof Error ? error.message : "Failed to create rule",
            variant: "destructive"
        });
    } finally {
        setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> New Alert Rule
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px] bg-background">
        <DialogHeader>
          <DialogTitle>Create Alert Rule</DialogTitle>
          <DialogDescription>
            Configure a new condition to trigger alerts.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <Input
                id="name"
                placeholder="e.g. High CPU Warning"
                className="col-span-3"
                value={formData.name}
                onChange={(e) => handleChange("name", e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="severity" className="text-right">
              Severity
            </Label>
            <Select value={formData.severity} onValueChange={(v) => handleChange("severity", v)}>
                <SelectTrigger className="col-span-3">
                    <SelectValue placeholder="Select severity" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="info">Info</SelectItem>
                    <SelectItem value="warning">Warning</SelectItem>
                    <SelectItem value="critical">Critical</SelectItem>
                </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="service" className="text-right">
              Service
            </Label>
            <Select value={formData.service} onValueChange={(v) => handleChange("service", v)}>
                <SelectTrigger className="col-span-3">
                    <SelectValue placeholder="Select service (optional)" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">All Services</SelectItem>
                    <SelectItem value="weather-service">weather-service</SelectItem>
                    <SelectItem value="api-gateway">api-gateway</SelectItem>
                    <SelectItem value="database">database</SelectItem>
                </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-4 items-start gap-4">
            <Label htmlFor="condition" className="text-right mt-2">
              Condition
            </Label>
            <div className="col-span-3 space-y-2">
                <Textarea
                    id="condition"
                    placeholder="e.g. cpu_usage > 90 AND duration > 5m"
                    className="font-mono text-xs"
                    value={formData.condition}
                    onChange={(e) => handleChange("condition", e.target.value)}
                />
                <p className="text-[10px] text-muted-foreground">
                    Supports PromQL or simple expression syntax.
                </p>
            </div>
          </div>
        </div>
        <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>Cancel</Button>
            <Button onClick={handleSave} disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Rule
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
