/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus, LayoutGrid, List } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/components/ui/tabs"
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";

import { UpstreamServiceConfig } from "@/lib/client";
import { ServiceList } from "@/components/services/service-list";
import { ServiceMarketplace } from "@/components/services/service-marketplace";

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    fetchServices();
  }, []);

  const fetchServices = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listServices();
      // Handle both array and object response formats for robustness
      if (Array.isArray(res)) {
          setServices(res);
      } else {
          setServices(res.services || []);
      }
    } catch (e) {
      console.error("Failed to fetch services", e);
      toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load services."
      });
    } finally {
        setLoading(false);
    }
  };

  const toggleService = async (name: string, enabled: boolean) => {
    // Optimistic update
    setServices(services.map(s => s.name === name ? { ...s, disable: !enabled } : s));

    try {
        await apiClient.setServiceStatus(name, !enabled);
        toast({
            title: enabled ? "Service Enabled" : "Service Disabled",
            description: `Service ${name} has been ${enabled ? "enabled" : "disabled"}.`
        });
    } catch (e) {
        console.error("Failed to toggle service", e);
        fetchServices(); // Revert
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update service status."
        });
    }
  };

  const deleteService = async (name: string) => {
    if (!confirm(`Are you sure you want to delete service "${name}"?`)) return;

    try {
        await apiClient.unregisterService(name);
        setServices(prev => prev.filter(s => s.name !== name));
        toast({
            title: "Service Deleted",
            description: `Service ${name} has been removed.`
        });
    } catch (e) {
         console.error("Failed to delete service", e);
         fetchServices();
         toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to delete service."
        });
    }
  }

  const openEdit = (service: UpstreamServiceConfig) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  };

  const openNew = () => {
      setSelectedService({ id: "", name: "", version: "1.0.0", disable: false, http_service: { address: "" } });
      setIsSheetOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
      e.preventDefault();
      if (!selectedService) return;

      try {
          if (selectedService.id) {
               // Update
               await apiClient.updateService(selectedService);
               toast({ title: "Service Updated", description: "Service configuration saved." });
          } else {
              // Create
              await apiClient.registerService(selectedService);
              toast({ title: "Service Created", description: "New service registered successfully." });
          }
          setIsSheetOpen(false);
          fetchServices();
      } catch (err) {
          console.error("Failed to save service", err);
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to save service configuration."
          });
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Services</h2>
        <Button onClick={openNew}>
            <Plus className="mr-2 h-4 w-4" /> Custom Service
        </Button>
      </div>

      <Tabs defaultValue="installed" className="space-y-4">
        <TabsList>
            <TabsTrigger value="installed" className="flex items-center gap-2">
                <List className="h-4 w-4" /> Installed
            </TabsTrigger>
            <TabsTrigger value="marketplace" className="flex items-center gap-2">
                <LayoutGrid className="h-4 w-4" /> Marketplace
            </TabsTrigger>
        </TabsList>

        <TabsContent value="installed" className="space-y-4">
            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Upstream Services</CardTitle>
                    <CardDescription>Manage your connected upstream services.</CardDescription>
                </CardHeader>
                <CardContent>
                    <ServiceList
                        services={services}
                        isLoading={loading}
                        onToggle={toggleService}
                        onEdit={openEdit}
                        onDelete={deleteService}
                    />
                </CardContent>
            </Card>
        </TabsContent>

        <TabsContent value="marketplace" className="space-y-4">
             <Card className="backdrop-blur-sm bg-background/50 border-dashed border-2">
                <CardHeader>
                    <CardTitle>Service Marketplace</CardTitle>
                    <CardDescription>Discover and install 1-click MCP servers.</CardDescription>
                </CardHeader>
                <CardContent>
                    <ServiceMarketplace onInstallComplete={() => {
                        fetchServices();
                        toast({ title: "Installation Complete", description: "The service is now available in your list." });
                    }} />
                </CardContent>
            </Card>
        </TabsContent>
      </Tabs>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent className="w-[400px] sm:w-[540px]">
            <SheetHeader>
                <SheetTitle>{selectedService?.id ? "Edit Service" : "New Service"}</SheetTitle>
                <SheetDescription>
                    Configure your upstream service details.
                </SheetDescription>
            </SheetHeader>
            {selectedService && (
                <form onSubmit={handleSave} className="grid gap-6 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">Name</Label>
                        <Input
                            id="name"
                            value={selectedService.name}
                            onChange={(e) => setSelectedService({...selectedService, name: e.target.value})}
                            className="col-span-3"
                        />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="type" className="text-right">Type</Label>
                        <div className="col-span-3">
                             <Select
                                value={selectedService.http_service ? "http" : selectedService.grpc_service ? "grpc" : selectedService.command_line_service ? "cmd" : "mcp"}
                                onValueChange={(val) => {
                                    const newService = { ...selectedService };
                                    delete newService.http_service;
                                    delete newService.grpc_service;
                                    delete newService.command_line_service;
                                    delete newService.mcp_service;

                                    if (val === 'http') newService.http_service = { address: "" };
                                    if (val === 'grpc') newService.grpc_service = { address: "" };
                                    if (val === 'cmd') newService.command_line_service = { command: "" };
                                    if (val === 'mcp') newService.mcp_service = { http_connection: { http_address: "" } };
                                    setSelectedService(newService);
                                }}
                             >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="http">HTTP</SelectItem>
                                    <SelectItem value="grpc">gRPC</SelectItem>
                                    <SelectItem value="cmd">Command Line</SelectItem>
                                    <SelectItem value="mcp">MCP Proxy</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    {selectedService.http_service && (
                         <div className="grid grid-cols-4 items-center gap-4">
                             <Label htmlFor="endpoint" className="text-right">Endpoint</Label>
                             <Input
                                id="endpoint"
                                value={selectedService.http_service.address}
                                onChange={(e) => setSelectedService({...selectedService, http_service: { ...selectedService.http_service, address: e.target.value }})}
                                placeholder="http://localhost:8080"
                                className="col-span-3"
                            />
                        </div>
                    )}
                     {selectedService.grpc_service && (
                         <div className="grid grid-cols-4 items-center gap-4">
                             <Label htmlFor="grpc-endpoint" className="text-right">Address</Label>
                             <Input
                                id="grpc-endpoint"
                                value={selectedService.grpc_service.address}
                                onChange={(e) => setSelectedService({...selectedService, grpc_service: { ...selectedService.grpc_service, address: e.target.value }})}
                                placeholder="localhost:9090"
                                className="col-span-3"
                            />
                        </div>
                    )}
                     {selectedService.command_line_service && (
                         <div className="grid grid-cols-4 items-center gap-4">
                             <Label htmlFor="command" className="text-right">Command</Label>
                             <Input
                                id="command"
                                value={selectedService.command_line_service.command}
                                onChange={(e) => setSelectedService({...selectedService, command_line_service: { ...selectedService.command_line_service, command: e.target.value }})}
                                placeholder="docker run ..."
                                className="col-span-3"
                            />
                        </div>
                    )}

                    <div className="flex justify-end gap-2 pt-4">
                        <Button type="button" variant="outline" onClick={() => setIsSheetOpen(false)}>Cancel</Button>
                        <Button type="submit">Save Changes</Button>
                    </div>
                </form>
            )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
