/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Plus, Trash, Edit } from "lucide-react";

interface Profile {
    id: string;
    name: string;
    description: string;
    services: string[];
    type: "dev" | "prod" | "debug";
}

const mockProfiles: Profile[] = [
    { id: "p1", name: "Default Dev", description: "Standard development profile", services: ["weather-service", "local-files"], type: "dev" },
    { id: "p2", name: "Production", description: "Production environment with strict limits", services: ["weather-service"], type: "prod" },
    { id: "p3", name: "Debug All", description: "All services enabled with verbose logging", services: ["weather-service", "memory-store", "local-files"], type: "debug" },
];

export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<Profile[]>(mockProfiles);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Profile
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {profiles.map(profile => (
              <Card key={profile.id} className="backdrop-blur-sm bg-background/50 hover:shadow-md transition-all">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-xl font-bold">{profile.name}</CardTitle>
                      <User className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                      <div className="text-sm text-muted-foreground mb-4">{profile.description}</div>
                      <div className="flex flex-wrap gap-2 mb-4">
                          <Badge variant={profile.type === 'prod' ? 'destructive' : profile.type === 'debug' ? 'secondary' : 'default'}>
                              {profile.type.toUpperCase()}
                          </Badge>
                          <span className="text-xs text-muted-foreground flex items-center">
                              {profile.services.length} Services
                          </span>
                      </div>
                      <div className="flex justify-end gap-2">
                          <Button variant="ghost" size="sm"><Edit className="h-3 w-3 mr-1"/> Edit</Button>
                          <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600"><Trash className="h-3 w-3 mr-1"/> Delete</Button>
                      </div>
                  </CardContent>
              </Card>
          ))}
      </div>
    </div>
  );
}
