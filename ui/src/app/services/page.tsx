/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useState, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Settings, Trash2, Plus } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

interface Service {
  id: string;
  name: string;
  version: string;
  disable: boolean;
  service_config?: any; // Simplified for now
}

export default function ServicesPage() {
  const [services, setServices] = useState<Service[]>([]);
  const [selectedService, setSelectedService] = useState<Service | null>(null);
  const [isNewService, setIsNewService] = useState(false);

  useEffect(() => {
    fetchServices();
  }, []);

  const fetchServices = async () => {
    const res = await fetch("/api/services");
    if (res.ok) {
      setServices(await res.json());
    }
  };

  const toggleService = async (id: string, currentStatus: boolean) => {
    // Optimistic update
    setServices(services.map(s => s.id === id ? { ...s, disable: !currentStatus } : s));

    try {
        await fetch("/api/services", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id, disable: !currentStatus })
        });
    } catch (e) {
        console.error("Failed to toggle service", e);
        fetchServices(); // Revert
    }
  };

  const handleSaveService = async (e: React.FormEvent) => {
      e.preventDefault();
      // Logic to save service (new or existing) would go here
      // For now, just close the sheet
      setSelectedService(null);
      setIsNewService(false);
  }

  const handleAddService = () => {
      setSelectedService({
          id: "",
          name: "",
          version: "",
          disable: false,
          service_config: {}
      });
      setIsNewService(true);
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Services</h2>
        <Sheet open={!!selectedService} onOpenChange={(open) => { if(!open) setSelectedService(null); }}>
             <SheetTrigger asChild>
                <Button onClick={handleAddService}><Plus className="mr-2 h-4 w-4" /> Add Service</Button>
            </SheetTrigger>
            <SheetContent className="w-[400px] sm:w-[540px]">
                <SheetHeader>
                    <SheetTitle>{isNewService ? "Add New Service" : "Edit Service"}</SheetTitle>
                    <SheetDescription>
                        {isNewService ? "Configure a new upstream service." : "Make changes to your service configuration here."}
                    </SheetDescription>
                </SheetHeader>
                {selectedService && (
                    <form onSubmit={handleSaveService} className="grid gap-4 py-4">
                        <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="name" className="text-right">
                                Name
                            </Label>
                            <Input
                                id="name"
                                value={selectedService.name}
                                onChange={(e) => setSelectedService({...selectedService, name: e.target.value})}
                                className="col-span-3"
                            />
                        </div>
                        <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="version" className="text-right">
                                Version
                            </Label>
                            <Input
                                id="version"
                                value={selectedService.version}
                                onChange={(e) => setSelectedService({...selectedService, version: e.target.value})}
                                className="col-span-3"
                            />
                        </div>
                         <div className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor="type" className="text-right">
                                Type
                            </Label>
                            <Select
                                defaultValue={selectedService.service_config?.http_service ? "http" : "grpc"}
                                onValueChange={(val) => {
                                    // simple mock update
                                }}
                            >
                                <SelectTrigger className="col-span-3">
                                    <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="http">HTTP</SelectItem>
                                    <SelectItem value="grpc">gRPC</SelectItem>
                                    <SelectItem value="mcp">MCP</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="flex justify-end pt-4">
                            <Button type="submit">Save changes</Button>
                        </div>
                    </form>
                )}
            </SheetContent>
        </Sheet>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 border-muted/20">
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
                      {service.service_config?.http_service ? (
                          <Badge variant="secondary">HTTP</Badge>
                      ) : service.service_config?.grpc_service ? (
                          <Badge variant="secondary">gRPC</Badge>
                      ) : (
                          <Badge variant="outline">Other</Badge>
                      )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{service.version}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={() => toggleService(service.id, service.disable)}
                        />
                        <span className="text-sm text-muted-foreground min-w-[60px]">
                            {!service.disable ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                      <Button variant="ghost" size="icon" onClick={() => { setIsNewService(false); setSelectedService(service); }}>
                          <Settings className="h-4 w-4" />
                      </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
