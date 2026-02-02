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
import { Severity } from "./types";

/**
 * CreateRuleDialog component.
 * @returns The rendered component.
 */
export function CreateRuleDialog() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [severity, setSeverity] = useState<Severity>("warning");
  const [service, setService] = useState("all");
  const [condition, setCondition] = useState("");
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleSave = async () => {
    if (!name || !condition) {
        toast({
            title: "Validation Error",
            description: "Name and Condition are required.",
            variant: "destructive"
        });
        return;
    }

    setLoading(true);
    try {
        await apiClient.createAlertRule({
            name,
            severity,
            service: service === "all" ? "" : service,
            condition,
            enabled: true
        });
        toast({
            title: "Rule Created",
            description: "Alert rule has been successfully created."
        });
        setOpen(false);
        // Reset form
        setName("");
        setCondition("");
        setSeverity("warning");
        setService("all");
    } catch (error: any) {
        toast({
            title: "Error",
            description: error.message || "Failed to create rule",
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
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="severity" className="text-right">
              Severity
            </Label>
            <Select value={severity} onValueChange={(v) => setSeverity(v as Severity)}>
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
            <Select value={service} onValueChange={setService}>
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
                    value={condition}
                    onChange={(e) => setCondition(e.target.value)}
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
