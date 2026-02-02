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
import { Search, Book, Play } from "lucide-react";
import { PromptDefinition } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import Link from "next/link";

interface ServicePromptsTableProps {
  prompts: PromptDefinition[];
  serviceId: string;
}

export function ServicePromptsTable({ prompts, serviceId }: ServicePromptsTableProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredPrompts = prompts.filter(prompt =>
    prompt.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    prompt.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (!prompts || prompts.length === 0) {
     return (
       <Card>
        <CardHeader>
          <CardTitle className="text-xl flex items-center gap-2"><Book className="h-5 w-5" />Prompts</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">No prompts configured for this service.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl flex items-center gap-2">
            <Book className="h-5 w-5" />
            Prompts
            <Badge variant="secondary" className="ml-2">
                {prompts.length}
            </Badge>
          </CardTitle>
          <div className="relative w-64">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search prompts..."
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
                <TableHead>Description</TableHead>
                <TableHead>Arguments</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredPrompts.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                    No prompts found matching "{searchQuery}"
                  </TableCell>
                </TableRow>
              ) : (
                filteredPrompts.map((prompt) => (
                  <TableRow key={prompt.name}>
                    <TableCell className="font-medium">
                        <Link href={`/service/${encodeURIComponent(serviceId)}/prompt/${encodeURIComponent(prompt.name)}`} className="hover:underline flex items-center gap-2">
                            {prompt.name}
                        </Link>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{prompt.description || "-"}</TableCell>
                    <TableCell>
                        {prompt.inputSchema?.properties ? (
                            <div className="flex flex-wrap gap-1">
                                {Object.keys(prompt.inputSchema.properties).map(name => (
                                    <Badge key={name} variant="outline" className="text-xs">
                                        {name}
                                        {prompt.inputSchema?.required?.includes(name) && <span className="text-red-500 ml-0.5">*</span>}
                                    </Badge>
                                ))}
                            </div>
                        ) : (
                            <span className="text-muted-foreground text-xs italic">No arguments</span>
                        )}
                    </TableCell>
                    <TableCell className="text-right">
                         <Button variant="ghost" size="sm" asChild>
                            <Link href={`/service/${encodeURIComponent(serviceId)}/prompt/${encodeURIComponent(prompt.name)}`}>
                                <Play className="h-4 w-4 mr-1" /> View
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
