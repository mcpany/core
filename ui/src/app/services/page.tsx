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

interface Service {
  id: string;
  name: string;
  version: string;
  disable: boolean;
  service_config?: any;
}

export default function ServicesPage() {
  const [services, setServices] = useState<Service[]>([]);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  useEffect(() => {
    fetchServices();
  }, []);

  const fetchServices = async () => {
    try {
      const res = await apiClient.listServices();
      setServices((res.services || []) as unknown as Service[]);
    } catch (e) {
      console.error("Failed to fetch services", e);
    }
  };

  const toggleService = async (id: string, currentStatus: boolean) => {
    // Optimistic update
    setServices(services.map(s => s.id === id ? { ...s, disable: !currentStatus } : s));

    try {
        await fetch("/api/v1/services", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id, disable: !currentStatus })
        });
    } catch (e) {
        console.error("Failed to toggle service", e);
        fetchServices(); // Revert
    }
  };

  const openEdit = (service: Service) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  };

  const openNew = () => {
      setSelectedService({ id: "", name: "", version: "1.0.0", disable: false, service_config: {} });
      setIsSheetOpen(true);
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
                <TableRow key={service.id}>
                  <TableCell className="font-medium">{service.name}</TableCell>
                  <TableCell>
                      <Badge variant="secondary">
                      {service.service_config?.http_service ? "HTTP" :
                       service.service_config?.grpc_service ? "gRPC" :
                       service.service_config?.command_line_service ? "CMD" : "Other"}
                      </Badge>
                  </TableCell>
                  <TableCell>{service.version}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={() => toggleService(service.id, service.disable)}
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
                <div className="grid gap-6 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">Name</Label>
                        <Input id="name" defaultValue={selectedService.name} className="col-span-3" />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="type" className="text-right">Type</Label>
                        <div className="col-span-3">
                             <Select defaultValue={selectedService.service_config?.http_service ? "http" : "grpc"}>
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
                    <div className="grid grid-cols-4 items-center gap-4">
                         <Label htmlFor="endpoint" className="text-right">Endpoint</Label>
                         <Input id="endpoint" placeholder="http://localhost:8080" className="col-span-3" />
                    </div>
                </div>
            )}
            <div className="flex justify-end gap-2">
                 <Button variant="outline" onClick={() => setIsSheetOpen(false)}>Cancel</Button>
                <Button type="submit">Save Changes</Button>
            </div>
        </SheetContent>
    </Sheet>
    </div>
  );
}
