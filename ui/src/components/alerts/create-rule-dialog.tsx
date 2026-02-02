/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
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
import { Severity } from "./types";

/**
 * CreateRuleDialog component.
 * @returns The rendered component.
 */
export function CreateRuleDialog() {
  const [open, setOpen] = useState(false);
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [services, setServices] = useState<string[]>([]);

  const [name, setName] = useState("");
  const [severity, setSeverity] = useState<Severity>("warning");
  const [service, setService] = useState("all");
  const [metric, setMetric] = useState("");
  const [operator, setOperator] = useState(">");
  const [threshold, setThreshold] = useState("");
  const [duration, setDuration] = useState("5m");

  useEffect(() => {
      async function fetchServices() {
          try {
              const list = await apiClient.listServices();
              setServices(list.map((s: any) => s.name));
          } catch (e) {
              console.error("Failed to load services", e);
          }
      }
      if (open) {
          fetchServices();
      }
  }, [open]);

  const handleSave = async () => {
    if (!name || !metric || !threshold) {
        toast({
            title: "Validation Error",
            description: "Please fill in all required fields.",
            variant: "destructive",
        });
        return;
    }

    setLoading(true);
    try {
        await apiClient.createAlertRule({
            name,
            severity,
            service: service === "all" ? "" : service,
            metric,
            operator,
            threshold: parseFloat(threshold),
            duration,
            enabled: true,
        });

        toast({
            title: "Rule Created",
            description: "Alert rule has been successfully created."
        });
        setOpen(false);
        // Reset form
        setName("");
        setMetric("");
        setThreshold("");
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
                    {services.map(s => (
                        <SelectItem key={s} value={s}>{s}</SelectItem>
                    ))}
                </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
             <Label htmlFor="metric" className="text-right">Metric</Label>
             <Input
                id="metric"
                placeholder="e.g. cpu_usage"
                className="col-span-3"
                value={metric}
                onChange={(e) => setMetric(e.target.value)}
             />
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="operator" className="text-right">Condition</Label>
            <div className="col-span-3 flex gap-2">
                 <Select value={operator} onValueChange={setOperator}>
                    <SelectTrigger className="w-[80px]">
                        <SelectValue />
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
                    className="flex-1"
                    value={threshold}
                    onChange={(e) => setThreshold(e.target.value)}
                />
            </div>
          </div>

          <div className="grid grid-cols-4 items-center gap-4">
             <Label htmlFor="duration" className="text-right">Duration</Label>
             <Input
                id="duration"
                placeholder="e.g. 5m"
                className="col-span-3"
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
             />
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
