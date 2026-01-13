/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";

import { UpstreamServiceConfig } from "@/lib/client";
import { ServiceList } from "@/components/services/service-list";
import { EnvVarEditor } from "@/components/services/env-var-editor";

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchServices = useCallback(async () => {
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
  }, [toast]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  const toggleService = useCallback(async (name: string, enabled: boolean) => {
    // Optimistic update
    setServices(prev => prev.map(s => s.name === name ? { ...s, disable: !enabled } : s));

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
  }, [fetchServices, toast]);

  const deleteService = useCallback(async (name: string) => {
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
  }, [fetchServices, toast]);

  const openEdit = useCallback((service: UpstreamServiceConfig) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  }, []);

  const openNew = () => {
      setSelectedService({ id: "", name: "", version: "1.0.0", disable: false, priority: 0, loadBalancingStrategy: 0, httpService: { address: "" } } as any);

      setIsSheetOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
      e.preventDefault();
      if (!selectedService) return;

      try {
          if (selectedService.id) {
               // Update
               await apiClient.updateService(selectedService as any);
               toast({ title: "Service Updated", description: "Service configuration saved." });
          } else {
              // Create
              await apiClient.registerService(selectedService as any);
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
            <Plus className="mr-2 h-4 w-4" /> Add Service
        </Button>
      </div>

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
                                value={selectedService.httpService ? "http" : selectedService.grpcService ? "grpc" : selectedService.commandLineService ? "cmd" : "mcp"}
                                onValueChange={(val) => {
                                    const newService = { ...selectedService };
                                    delete newService.httpService;
                                    delete newService.grpcService;
                                    delete newService.commandLineService;
                                    delete newService.mcpService;


                                    if (val === 'http') newService.httpService = { address: "", tools: [], calls: {}, resources: [], prompts: [] };
                                    if (val === 'grpc') newService.grpcService = { address: "", useReflection: false, tools: [], resources: [], calls: {}, prompts: [], protoDefinitions: [], protoCollection: [] };
                                    if (val === 'cmd') newService.commandLineService = { command: "", workingDirectory: "", local: false, env: {}, tools: [], resources: [], prompts: [], communicationProtocol: 0, calls: {} };
                                    if (val === 'mcp') newService.mcpService = { toolAutoDiscovery: true, tools: [], resources: [], calls: {}, prompts: [] };

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

                    {selectedService.httpService && (
                         <div className="grid grid-cols-4 items-center gap-4">
                             <Label htmlFor="endpoint" className="text-right">Endpoint</Label>
                             <Input
                                id="endpoint"
                                value={selectedService.httpService.address}
                                onChange={(e) => setSelectedService({...selectedService, httpService: { ...selectedService.httpService, address: e.target.value } as any})}
                                placeholder="http://localhost:8080"
                                className="col-span-3"
                            />
                        </div>
                    )}
                     {selectedService.grpcService && (
                         <div className="grid grid-cols-4 items-center gap-4">
                             <Label htmlFor="grpc-endpoint" className="text-right">Address</Label>
                             <Input
                                id="grpc-endpoint"
                                value={selectedService.grpcService.address}
                                onChange={(e) => setSelectedService({...selectedService, grpcService: { ...selectedService.grpcService, address: e.target.value } as any})}
                                placeholder="localhost:9090"
                                className="col-span-3"
                            />
                        </div>
                    )}
                     {selectedService.commandLineService && (
                         <>
                             <div className="grid grid-cols-4 items-center gap-4">
                                 <Label htmlFor="command" className="text-right">Command</Label>
                                 <Input
                                    id="command"
                                    value={selectedService.commandLineService.command}
                                    onChange={(e) => setSelectedService({...selectedService, commandLineService: { ...selectedService.commandLineService, command: e.target.value } as any})}
                                    placeholder="docker run ..."
                                    className="col-span-3"
                                />
                            </div>
                             <div className="grid grid-cols-4 items-center gap-4">
                                 <Label htmlFor="workingDirectory" className="text-right">Working Dir</Label>
                                 <Input
                                    id="workingDirectory"
                                    value={selectedService.commandLineService.workingDirectory || ""}
                                    onChange={(e) => setSelectedService({...selectedService, commandLineService: { ...selectedService.commandLineService, workingDirectory: e.target.value } as any})}
                                    placeholder="/app"
                                    className="col-span-3"
                                />
                            </div>
                            <div className="col-span-4 border-t pt-4 mt-4">
                                <EnvVarEditor
                                    initialEnv={selectedService.commandLineService.env as any}
                                    onChange={(newEnv) => setSelectedService({...selectedService, commandLineService: { ...selectedService.commandLineService, env: newEnv } as any})}
                                />
                            </div>
                        </>
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
