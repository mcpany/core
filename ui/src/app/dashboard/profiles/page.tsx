/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface Profile {
  id: string;
  name: string;
  roles: string[];
  tags: string[];
}

export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<Profile[]>([
      { id: "1", name: "Default Profile", roles: ["admin", "user"], tags: ["prod"] },
      { id: "2", name: "Dev Profile", roles: ["developer"], tags: ["dev", "test"] }
  ]);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Profile
        </Button>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Execution Profiles</CardTitle>
          <CardDescription>Manage user and execution profiles.</CardDescription>
        </CardHeader>
        <CardContent>
           <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Roles</TableHead>
                <TableHead>Tags</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {profiles.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>
                      <div className="flex gap-1">
                          {p.roles.map(r => <Badge key={r} variant="outline">{r}</Badge>)}
                      </div>
                  </TableCell>
                  <TableCell>
                      <div className="flex gap-1">
                          {p.tags.map(t => <Badge key={t} variant="secondary">{t}</Badge>)}
                      </div>
                  </TableCell>
                  <TableCell className="text-right">
                      <Button variant="ghost" size="sm">Edit</Button>
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
