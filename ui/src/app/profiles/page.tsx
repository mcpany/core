/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Plus, Trash, Edit } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";

interface Profile {
    id: string;
    name: string;
    description: string;
    services: string[];
    type: "dev" | "prod" | "debug";
}

/**
 * ProfilesPage component.
 * @returns The rendered component.
 */
export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newProfileName, setNewProfileName] = useState("");
  const [newProfileDesc, setNewProfileDesc] = useState("");

  const fetchProfiles = async () => {
      setIsLoading(true);
      try {
          const data = await apiClient.listProfiles();
          // Map backend ProfileDefinition to UI Profile
          const mapped: Profile[] = data.map((p: any) => ({
              id: p.name, // Use name as ID
              name: p.name,
              description: "", // Not supported in backend yet
              services: p.serviceConfig ? Object.keys(p.serviceConfig) : [],
              type: (p.selector?.tags?.find((t: string) => ["dev", "prod", "debug"].includes(t)) as "dev" | "prod" | "debug") || "dev"
          }));
          setProfiles(mapped);
      } catch (error) {
          console.error("Failed to fetch profiles", error);
          toast.error("Failed to fetch profiles");
      } finally {
          setIsLoading(false);
      }
  };

  useEffect(() => {
      fetchProfiles();
  }, []);

  const handleCreate = async () => {
      try {
          // Backend expects ProfileDefinition
          const newProfile = {
              name: newProfileName,
              selector: {
                  tags: ["dev"] // Default to dev
              },
              serviceConfig: {}, // Empty initially
              secrets: {}
          };
          await apiClient.createProfile(newProfile);
          toast.success("Profile created");
          setIsDialogOpen(false);
          setNewProfileName("");
          setNewProfileDesc("");
          fetchProfiles();
      } catch (error) {
          console.error("Failed to create profile", error);
          toast.error("Failed to create profile");
      }
  };

  const handleDelete = async (name: string) => {
      try {
          await apiClient.deleteProfile(name);
          toast.success("Profile deleted");
          fetchProfiles();
      } catch (error) {
           console.error("Failed to delete profile", error);
           toast.error("Failed to delete profile");
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Create Profile
                </Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Create Profile</DialogTitle>
                    <DialogDescription>Add a new profile to manage service access.</DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">Name</Label>
                        <Input id="name" value={newProfileName} onChange={(e) => setNewProfileName(e.target.value)} className="col-span-3" />
                    </div>
                    {/* Description is not saved to backend currently, but UI prompts for it. Maybe remove? Or keep for future. */}
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="description" className="text-right">Description</Label>
                        <Input id="description" value={newProfileDesc} onChange={(e) => setNewProfileDesc(e.target.value)} className="col-span-3" />
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleCreate}>Save</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {isLoading && <div className="col-span-3 text-center p-4">Loading profiles...</div>}
          {!isLoading && profiles.length === 0 && <div className="col-span-3 text-center p-4 text-muted-foreground">No profiles found. Create one to get started.</div>}
          {profiles.map(profile => (
              <Card key={profile.id} className="backdrop-blur-sm bg-background/50 hover:shadow-md transition-all">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-xl font-bold">{profile.name}</CardTitle>
                      <User className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                      <div className="text-sm text-muted-foreground mb-4">{profile.description || "No description"}</div>
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
                          <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600" onClick={() => handleDelete(profile.name)}><Trash className="h-3 w-3 mr-1"/> Delete</Button>
                      </div>
                  </CardContent>
              </Card>
          ))}
      </div>
    </div>
  );
}
