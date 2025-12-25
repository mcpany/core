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
import { MoreHorizontal, Settings, Trash2, Search, Server } from "lucide-react";
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
  const [search, setSearch] = useState("");

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

  const filteredServices = services.filter(service =>
    service.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Services</h2>
        <div className="flex items-center space-x-2">
            <div className="relative w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search services..."
                    className="pl-8"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                />
            </div>
            <Button>Add Service</Button>
        </div>
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
              {filteredServices.map((service) => (
                <TableRow key={service.id}>
                  <TableCell className="font-medium flex items-center gap-2">
                      <Server className="h-4 w-4 text-muted-foreground" />
                      {service.name}
                  </TableCell>
                  <TableCell>
                      {service.service_config?.http_service ? <Badge variant="outline">HTTP</Badge> :
                       service.service_config?.grpc_service ? <Badge variant="outline">gRPC</Badge> : "Other"}
                  </TableCell>
                  <TableCell>{service.version}</TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!service.disable}
                            onCheckedChange={() => toggleService(service.id, service.disable)}
                        />
                        <span className="text-sm text-muted-foreground w-16">
                            {!service.disable ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <Sheet>
                        <SheetTrigger asChild>
                             <Button variant="ghost" size="icon" onClick={() => setSelectedService(service)}>
                                <Settings className="h-4 w-4" />
                            </Button>
                        </SheetTrigger>
                        <SheetContent className="w-[400px] sm:w-[540px]">
                            <SheetHeader>
                                <SheetTitle>Edit Service</SheetTitle>
                                <SheetDescription>
                                    Make changes to your service configuration here. Click save when you're done.
                                </SheetDescription>
                            </SheetHeader>
                            {selectedService && (
                                <div className="grid gap-4 py-4">
                                    <div className="grid grid-cols-4 items-center gap-4">
                                        <Label htmlFor="name" className="text-right">
                                            Name
                                        </Label>
                                        <Input id="name" defaultValue={selectedService.name} className="col-span-3" />
                                    </div>
                                    <div className="grid grid-cols-4 items-center gap-4">
                                        <Label htmlFor="version" className="text-right">
                                            Version
                                        </Label>
                                        <Input id="version" defaultValue={selectedService.version} className="col-span-3" />
                                    </div>
                                     {/* More complex forms would go here */}
                                </div>
                            )}
                             <div className="flex justify-end">
                                <Button type="submit">Save changes</Button>
                            </div>
                        </SheetContent>
                    </Sheet>
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
