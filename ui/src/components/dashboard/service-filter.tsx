/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDashboard } from "@/components/dashboard/dashboard-context";
import { apiClient } from "@/lib/client";
import { Filter } from "lucide-react";

export function ServiceFilter() {
  const { serviceId, setServiceId } = useDashboard();
  const [services, setServices] = useState<{ id: string; name: string }[]>([]);

  useEffect(() => {
    async function loadServices() {
      try {
        const list = await apiClient.listServices();
        // Use name as the ID for filtering since most configs use name as the service identifier
        setServices(list.map((s: any) => ({ id: s.name, name: s.name })));
      } catch (error) {
        console.error("Failed to load services for filter", error);
      }
    }
    loadServices();
  }, []);

  return (
    <div className="flex items-center space-x-2">
      <Filter className="h-4 w-4 text-muted-foreground" />
      <Select
        value={serviceId || "all"}
        onValueChange={(value) => setServiceId(value === "all" ? undefined : value)}
      >
        <SelectTrigger className="w-[200px] h-8">
          <SelectValue placeholder="All Services" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Services</SelectItem>
          {services.map((svc) => (
            <SelectItem key={svc.id} value={svc.id}>
              {svc.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
