/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { ServicesTable } from "@/components/services/services-table";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export default function ServicesPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Services</h2>
        <div className="flex items-center space-x-2">
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Service
          </Button>
        </div>
      </div>
      <ServicesTable />
    </div>
  );
}
