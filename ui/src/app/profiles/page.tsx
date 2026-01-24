/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Plus, Trash, Edit } from "lucide-react";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { ProfileEditor, Profile } from "@/components/profiles/profile-editor";

/**
 * ProfilesPage component.
 * @returns The rendered component.
 */
export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Editor State
  const [editingProfile, setEditingProfile] = useState<Profile | null>(null);
  const [isEditorOpen, setIsEditorOpen] = useState(false);

  const fetchProfiles = async () => {
      setIsLoading(true);
      try {
          const data = await apiClient.listProfiles();
          // Map backend ProfileDefinition to UI Profile
          const mapped: Profile[] = data.map((p: any) => ({
              id: p.name, // Use name as ID
              name: p.name,
              description: "", // Not supported in backend yet, but we can display placeholder
              services: p.serviceConfig ? Object.keys(p.serviceConfig) : [],
              type: (p.selector?.tags?.find((t: string) => ["dev", "prod", "debug"].includes(t)) as "dev" | "prod" | "debug") || "dev",
              secrets: p.secrets
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

  const handleSaveProfile = async (profileData: any) => {
      try {
          if (editingProfile) {
              await apiClient.updateProfile(profileData);
              toast.success("Profile updated");
          } else {
              await apiClient.createProfile(profileData);
              toast.success("Profile created");
          }
          setIsEditorOpen(false);
          setEditingProfile(null);
          fetchProfiles();
      } catch (error) {
          console.error("Failed to save profile", error);
          toast.error("Failed to save profile");
      }
  };

  const handleDelete = async (name: string) => {
      if (!confirm(`Are you sure you want to delete profile "${name}"?`)) return;
      try {
          await apiClient.deleteProfile(name);
          toast.success("Profile deleted");
          fetchProfiles();
      } catch (error) {
           console.error("Failed to delete profile", error);
           toast.error("Failed to delete profile");
      }
  };

  const openNew = () => {
      setEditingProfile(null);
      setIsEditorOpen(true);
  };

  const openEdit = (p: Profile) => {
      setEditingProfile(p);
      setIsEditorOpen(true);
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
        <Button onClick={openNew}>
            <Plus className="mr-2 h-4 w-4" /> Create Profile
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {isLoading && <div className="col-span-3 text-center p-4">Loading profiles...</div>}
          {!isLoading && profiles.length === 0 && <div className="col-span-3 text-center p-4 text-muted-foreground">No profiles found. Create one to get started.</div>}
          {profiles.map(profile => (
              <Card key={profile.id} className="backdrop-blur-sm bg-background/50 hover:shadow-md transition-all cursor-pointer group" onClick={() => openEdit(profile)}>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-xl font-bold group-hover:text-primary transition-colors">{profile.name}</CardTitle>
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
                      <div className="flex justify-end gap-2 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity">
                          <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); openEdit(profile); }}>
                              <Edit className="h-3 w-3 mr-1"/> Edit
                          </Button>
                          <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600" onClick={(e) => { e.stopPropagation(); handleDelete(profile.name); }}>
                              <Trash className="h-3 w-3 mr-1"/> Delete
                          </Button>
                      </div>
                  </CardContent>
              </Card>
          ))}
      </div>

      <ProfileEditor
        open={isEditorOpen}
        onOpenChange={setIsEditorOpen}
        profile={editingProfile}
        onSave={handleSaveProfile}
      />
    </div>
  );
}
