/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Settings, Plus } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/components/ui/sheet"
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { UpstreamServiceConfig } from "@/lib/client";

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  useEffect(() => {
    fetchServices();
  }, []);

  const fetchServices = async () => {
    try {
      const res = await apiClient.listServices();
      setServices(res.services || []);
    } catch (e) {
      console.error("Failed to fetch services", e);
    }
  };

  const toggleService = async (name: string, currentStatus: boolean) => {
    // Optimistic update
    setServices(services.map(s => s.name === name ? { ...s, disable: !currentStatus } : s));

    try {
        await apiClient.setServiceStatus(name, !currentStatus);
    } catch (e) {
        console.error("Failed to toggle service", e);
        fetchServices(); // Revert
    }
  };

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
          if (services.find(s => s.name === selectedService.name && s.id !== selectedService.id)) {
             // Logic for handling duplicate names or editing existing
          }

          if (selectedService.id) {
               // Update
               // Note: Real implementation would align fields carefully
               await apiClient.updateService(selectedService as any);
          } else {
              // Create
              await apiClient.registerService(selectedService as any);
          }
          setIsSheetOpen(false);
          fetchServices();
      } catch (err) {
          console.error("Failed to save service", err);
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
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Version</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {services.map((service) => (
                <TableRow key={service.name}>
                  <TableCell className="font-medium">{service.name}</TableCell>
                  <TableCell>
                      <Badge variant="secondary">
                      {service.http_service ? "HTTP" :
                       service.grpc_service ? "gRPC" :
                       service.command_line_service ? "CMD" :
                       service.mcp_service ? "MCP" : "Other"}
                      </Badge>
                  </TableCell>
                  <TableCell>{service.version}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={() => toggleService(service.name, !!service.disable)}
                        />
                        <span className="text-sm text-muted-foreground w-16">
                            {!service.disable ? "Active" : "Inactive"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                        <Button variant="ghost" size="icon" onClick={() => openEdit(service)}>
                            <Settings className="h-4 w-4" />
                        </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
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
