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
import { Plus } from "lucide-react";
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
    metric: "cpu_usage",
    operator: ">",
    threshold: 90,
    duration: "5m",
    severity: "warning",
    enabled: true
  });

  const handleChange = (field: string, value: any) => {
      setFormData(prev => ({ ...prev, [field]: value }));
  };

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
            metric: "cpu_usage",
            operator: ">",
            threshold: 90,
            duration: "5m",
            severity: "warning",
            enabled: true
        });
    } catch (error) {
        console.error(error);
        toast({
            title: "Error",
            description: "Failed to create alert rule",
            variant: "destructive",
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
            <Label htmlFor="metric" className="text-right">
              Metric
            </Label>
            <Input
                id="metric"
                placeholder="e.g. cpu_usage"
                className="col-span-3"
                value={formData.metric}
                onChange={(e) => handleChange("metric", e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="operator" className="text-right">
              Condition
            </Label>
            <div className="col-span-3 flex gap-2">
                <Select value={formData.operator} onValueChange={(val) => handleChange("operator", val)}>
                    <SelectTrigger className="w-[80px]">
                        <SelectValue placeholder="Op" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value=">">{">"}</SelectItem>
                        <SelectItem value="<">{"<"}</SelectItem>
                        <SelectItem value="=">{"="}</SelectItem>
                        <SelectItem value=">=">{">="}</SelectItem>
                        <SelectItem value="<=">{"<="}</SelectItem>
                    </SelectContent>
                </Select>
                <Input
                    type="number"
                    placeholder="Threshold"
                    value={formData.threshold}
                    onChange={(e) => handleChange("threshold", e.target.value)}
                    className="flex-1"
                />
            </div>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="duration" className="text-right">
              Duration
            </Label>
            <Input
                id="duration"
                placeholder="e.g. 5m"
                className="col-span-3"
                value={formData.duration}
                onChange={(e) => handleChange("duration", e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="severity" className="text-right">
              Severity
            </Label>
            <Select value={formData.severity} onValueChange={(val) => handleChange("severity", val)}>
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
            <Button onClick={handleSave} disabled={loading || !formData.name}>
                {loading ? "Creating..." : "Create Rule"}
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
