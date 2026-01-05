/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useState } from "react";
import { MARKETPLACE_ITEMS, MarketplaceItem } from "@/lib/marketplace-data";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { HardDrive, Github, Database, Terminal, Download } from "lucide-react";

// Icon mapping
const ICON_MAP: Record<string, React.ComponentType<any>> = {
  HardDrive,
  Github,
  Database,
  Terminal,
};

export default function MarketplacePage() {
  const { toast } = useToast();
  const [selectedItem, setSelectedItem] = useState<MarketplaceItem | null>(null);
  const [envValues, setEnvValues] = useState<Record<string, string>>({});
  const [isInstalling, setIsInstalling] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleInstallClick = (item: MarketplaceItem) => {
    setSelectedItem(item);
    setEnvValues({});
    setIsDialogOpen(true);
  };

  const handleInputChange = (key: string, value: string) => {
    setEnvValues((prev) => ({ ...prev, [key]: value }));
  };

  const handleConfirmInstall = async () => {
    if (!selectedItem) return;

    // Validate required fields
    const missing = selectedItem.config.envVars
      .filter((v) => v.required && !envValues[v.name])
      .map((v) => v.name);

    if (missing.length > 0) {
      toast({
        title: "Missing Required Fields",
        description: `Please fill in: ${missing.join(", ")}`,
        variant: "destructive",
      });
      return;
    }

    setIsInstalling(true);

    try {
      // Build payload
      const finalArgs = [...selectedItem.config.args];
      const finalEnv: Record<string, string> = { ...envValues }; // Start with just the values

      // Process defined envVars to see if they should be appended to args
      selectedItem.config.envVars.forEach((ev) => {
        const val = envValues[ev.name];
        if (val && ev.addToArgs) {
          finalArgs.push(val);
        }
      });

      // Construct UpstreamServiceConfig
      await apiClient.registerService({
        name: selectedItem.id, // Using ID as name
        command_line_service: {
          command: selectedItem.config.command,
          args: finalArgs,
          env: finalEnv,
        },
      });

      toast({
        title: "Installation Started",
        description: `Service ${selectedItem.name} is being installed.`,
      });
      setIsDialogOpen(false);
    } catch (error) {
        console.error("Install failed", error);
        toast({
            title: "Installation Failed",
            description: String(error),
            variant: "destructive",
        });
    } finally {
      setIsInstalling(false);
    }
  };

  return (
    <div className="flex flex-col gap-6 p-8 h-[calc(100vh-4rem)]">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Marketplace</h1>
        <p className="text-muted-foreground mt-2">
          Discover and install standard MCP servers to extend your capabilities.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {MARKETPLACE_ITEMS.map((item) => {
          const Icon = ICON_MAP[item.icon] || Terminal;
          return (
            <Card key={item.id} className="flex flex-col">
              <CardHeader>
                <div className="flex items-center gap-2 mb-2">
                  <div className="p-2 rounded-md bg-muted text-muted-foreground">
                    <Icon className="h-6 w-6" />
                  </div>
                </div>
                <CardTitle>{item.name}</CardTitle>
                <CardDescription>{item.description}</CardDescription>
              </CardHeader>
              <CardContent className="flex-1">
                {/* Placeholder for tags or more info */}
              </CardContent>
              <CardFooter>
                <Button className="w-full" onClick={() => handleInstallClick(item)}>
                  <Download className="mr-2 h-4 w-4" /> Install
                </Button>
              </CardFooter>
            </Card>
          );
        })}
      </div>

      {/* Install Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Install {selectedItem?.name}</DialogTitle>
            <DialogDescription>
              Configure the service before installation.
            </DialogDescription>
          </DialogHeader>

          {selectedItem && (
            <div className="grid gap-4 py-4">
              {selectedItem.config.envVars.map((ev) => (
                <div key={ev.name} className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor={ev.name} className="text-right">
                    {ev.name}
                  </Label>
                  <Input
                    id={ev.name}
                    type={ev.type === "password" ? "password" : "text"}
                    placeholder={ev.description}
                    className="col-span-3"
                    value={envValues[ev.name] || ""}
                    onChange={(e) => handleInputChange(ev.name, e.target.value)}
                  />
                </div>
              ))}
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleConfirmInstall} disabled={isInstalling}>
              {isInstalling ? "Installing..." : "Confirm Installation"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
