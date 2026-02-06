/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Search, Filter } from "lucide-react";

interface TraceFilterProps {
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  statusFilter: string;
  setStatusFilter: (status: string) => void;
}

/**
 * TraceFilter component for filtering Inspector traces.
 *
 * @param props - The component props.
 * @param props.searchQuery - The current search query.
 * @param props.setSearchQuery - Callback to update search query.
 * @param props.statusFilter - The current status filter (all, success, error).
 * @param props.setStatusFilter - Callback to update status filter.
 * @returns The rendered component.
 */
export function TraceFilter({
  searchQuery,
  setSearchQuery,
  statusFilter,
  setStatusFilter,
}: TraceFilterProps) {
  return (
    <div className="flex items-center gap-2">
      <div className="relative w-[200px] md:w-[300px]">
        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Filter by name or ID..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-8 h-9"
        />
      </div>
      <Select value={statusFilter} onValueChange={setStatusFilter}>
        <SelectTrigger className="w-[130px] h-9">
          <Filter className="mr-2 h-3.5 w-3.5 text-muted-foreground" />
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Status</SelectItem>
          <SelectItem value="success">Success</SelectItem>
          <SelectItem value="error">Error</SelectItem>
          <SelectItem value="pending">Pending</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
}
