/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, Settings, Trash } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { mockServices } from "@/lib/mock-data";
import { useState } from "react";
import Link from "next/link";

export function ServicesTable() {
  const [services, setServices] = useState(mockServices);

  const toggleService = (id: string) => {
    setServices(services.map(s =>
      s.id === id ? { ...s, disable: !s.disable } : s
    ));
  };

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Version</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {services.map((service) => (
            <TableRow key={service.id}>
              <TableCell className="font-medium">
                <Link href={`/services/${service.id}`} className="hover:underline">
                    {service.name}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant="outline">
                    {service.serviceConfig?.case?.replace('Service', '') || 'Unknown'}
                </Badge>
              </TableCell>
              <TableCell>{service.version}</TableCell>
              <TableCell>
                <div className="flex items-center space-x-2">
                    <Switch
                        checked={!service.disable}
                        onCheckedChange={() => toggleService(service.id)}
                    />
                    <span className="text-sm text-muted-foreground">
                        {!service.disable ? 'Enabled' : 'Disabled'}
                    </span>
                </div>
              </TableCell>
              <TableCell>{service.priority}</TableCell>
              <TableCell className="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" className="h-8 w-8 p-0">
                      <span className="sr-only">Open menu</span>
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                    <DropdownMenuItem>
                      <Settings className="mr-2 h-4 w-4" /> Configure
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-600">
                      <Trash className="mr-2 h-4 w-4" /> Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
