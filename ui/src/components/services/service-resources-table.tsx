/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Search, Database, Eye } from "lucide-react";
import { ResourceDefinition } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import Link from "next/link";

interface ServiceResourcesTableProps {
  resources: ResourceDefinition[];
  serviceId: string;
}

export function ServiceResourcesTable({ resources, serviceId }: ServiceResourcesTableProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredResources = resources.filter(resource =>
    resource.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    resource.uri.toLowerCase().includes(searchQuery.toLowerCase()) ||
    resource.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (!resources || resources.length === 0) {
     return (
       <Card>
        <CardHeader>
          <CardTitle className="text-xl flex items-center gap-2"><Database className="h-5 w-5" />Resources</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">No resources configured for this service.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl flex items-center gap-2">
            <Database className="h-5 w-5" />
            Resources
            <Badge variant="secondary" className="ml-2">
                {resources.length}
            </Badge>
          </CardTitle>
          <div className="relative w-64">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search resources..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 h-9"
            />
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>URI</TableHead>
                <TableHead>MIME Type</TableHead>
                <TableHead>Description</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredResources.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                    No resources found matching "{searchQuery}"
                  </TableCell>
                </TableRow>
              ) : (
                filteredResources.map((resource) => (
                  <TableRow key={resource.uri}>
                    <TableCell className="font-medium">
                        <Link href={`/service/${encodeURIComponent(serviceId)}/resource/${encodeURIComponent(resource.name)}`} className="hover:underline flex items-center gap-2">
                            {resource.name}
                        </Link>
                    </TableCell>
                     <TableCell className="font-mono text-xs text-muted-foreground truncate max-w-[200px]" title={resource.uri}>
                        {resource.uri}
                    </TableCell>
                    <TableCell>
                         {resource.mimeType ? <Badge variant="outline">{resource.mimeType}</Badge> : "-"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">{resource.description || "-"}</TableCell>
                    <TableCell className="text-right">
                         <Button variant="ghost" size="sm" asChild>
                            <Link href={`/service/${encodeURIComponent(serviceId)}/resource/${encodeURIComponent(resource.name)}`}>
                                <Eye className="h-4 w-4 mr-1" /> View
                            </Link>
                         </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
