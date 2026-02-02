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
import { Plus, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import { AlertRule, Severity } from "./types";

/**
 * CreateRuleDialog component.
 * @returns The rendered component.
 */
export function CreateRuleDialog() {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const [formData, setFormData] = useState<Partial<AlertRule>>({
      name: "",
      metric: "",
      operator: ">",
      threshold: 0,
      duration: "5m",
      severity: "warning",
      enabled: true
  });

  const handleSave = async () => {
    setLoading(true);
    try {
        await apiClient.createAlertRule({
            ...formData,
            threshold: Number(formData.threshold)
        });
        toast({
            title: "Rule Created",
            description: "Alert rule has been successfully created."
        });
        setOpen(false);
        // Reset form
        setFormData({
            name: "",
            metric: "",
            operator: ">",
            threshold: 0,
            duration: "5m",
            severity: "warning",
            enabled: true
        });
    } catch (err) {
        console.error(err);
        toast({
            title: "Error",
            description: "Failed to create rule",
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
            <Label htmlFor="name" className="text-right">Name</Label>
            <Input
                id="name"
                value={formData.name}
                onChange={e => setFormData({...formData, name: e.target.value})}
                placeholder="e.g. High CPU"
                className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="metric" className="text-right">Metric</Label>
            <Input
                id="metric"
                value={formData.metric}
                onChange={e => setFormData({...formData, metric: e.target.value})}
                placeholder="e.g. cpu_usage"
                className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="operator" className="text-right">Operator</Label>
            <Select
                value={formData.operator}
                onValueChange={v => setFormData({...formData, operator: v})}
            >
                <SelectTrigger className="col-span-3">
                    <SelectValue placeholder="Select operator" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value=">">{">"}</SelectItem>
                    <SelectItem value="<">{"<"}</SelectItem>
                    <SelectItem value="=">{"="}</SelectItem>
                </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="threshold" className="text-right">Threshold</Label>
            <Input
                id="threshold"
                type="number"
                value={formData.threshold}
                onChange={e => setFormData({...formData, threshold: Number(e.target.value)})}
                className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="duration" className="text-right">Duration</Label>
            <Input
                id="duration"
                value={formData.duration}
                onChange={e => setFormData({...formData, duration: e.target.value})}
                placeholder="e.g. 5m"
                className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="severity" className="text-right">Severity</Label>
            <Select
                value={formData.severity}
                onValueChange={v => setFormData({...formData, severity: v as Severity})}
            >
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
        </div>
        <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Rule
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
