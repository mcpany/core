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
import { Webhook } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

/**
 * WebhookDialog component.
 * @returns The rendered component.
 */
export function WebhookDialog() {
  const [open, setOpen] = useState(false);
  const [url, setUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (open) {
      setLoading(true);
      apiClient.getWebhookURL()
        .then((data) => setUrl(data.url))
        .catch(() => {
          toast({
            title: "Error",
            description: "Failed to load webhook URL",
            variant: "destructive",
          });
        })
        .finally(() => setLoading(false));
    }
  }, [open, toast]);

  const handleSave = async () => {
    try {
      setLoading(true);
      await apiClient.saveWebhookURL(url);
      toast({
        title: "Webhook Configured",
        description: "Global webhook URL has been updated.",
      });
      setOpen(false);
    } catch (error) {
        console.error(error);
      toast({
        title: "Error",
        description: "Failed to save webhook URL",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="gap-2">
            <Webhook className="h-4 w-4" /> Configure Webhook
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px] bg-background">
        <DialogHeader>
          <DialogTitle>Configure Global Webhook</DialogTitle>
          <DialogDescription>
            Enter a URL to receive notifications when upstream service health changes (alerts).
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="url" className="text-right">
              Webhook URL
            </Label>
            <Input
                id="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://example.com/webhook"
                className="col-span-3"
            />
          </div>
        </div>
        <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={loading}>
                {loading ? "Saving..." : "Save Configuration"}
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
