/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Plus, Trash, Edit, RefreshCw } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { ProfileEditor } from "@/components/profiles/profile-editor";
import { ProfileDefinition } from "@/types/profile";

/**
 * ProfilesPage component.
 * @returns The rendered component.
 */
export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState<ProfileDefinition | null>(null);

  const fetchProfiles = async () => {
      setIsLoading(true);
      try {
          const data = await apiClient.listProfiles();
          setProfiles(data);
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

  const handleEdit = (profile: ProfileDefinition) => {
      setSelectedProfile(profile);
      setIsSheetOpen(true);
  };

  const handleCreateNew = () => {
      setSelectedProfile(null);
      setIsSheetOpen(true);
  };

  const handleSave = async (profile: ProfileDefinition) => {
      try {
          if (selectedProfile) {
              await apiClient.updateProfile(profile);
              toast.success("Profile updated");
          } else {
              await apiClient.createProfile(profile);
              toast.success("Profile created");
          }
          setIsSheetOpen(false);
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

  const getType = (profile: ProfileDefinition) => {
      const tags = profile.selector?.tags || [];
      if (tags.includes("prod")) return "prod";
      if (tags.includes("debug")) return "debug";
      return "dev"; // default
  };

  const getServiceCount = (profile: ProfileDefinition) => {
      // This assumes we know how many services match tags, which we don't without querying backend logic.
      // But we can count overrides.
      // Just show override count + "auto"
      const overrides = profile.serviceConfig ? Object.keys(profile.serviceConfig).length : 0;
      return overrides > 0 ? `${overrides} Overrides` : "Auto-selected";
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <div>
             <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
             <p className="text-muted-foreground">Manage execution profiles and service visibility.</p>
        </div>

        <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={fetchProfiles}>
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
            <Button onClick={handleCreateNew}>
                <Plus className="mr-2 h-4 w-4" /> Create Profile
            </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {isLoading && profiles.length === 0 && <div className="col-span-3 text-center p-4">Loading profiles...</div>}
          {!isLoading && profiles.length === 0 && (
              <div className="col-span-3 flex flex-col items-center justify-center p-12 border border-dashed rounded-lg bg-muted/10">
                  <User className="h-12 w-12 text-muted-foreground mb-4 opacity-20" />
                  <h3 className="text-lg font-medium">No profiles found</h3>
                  <p className="text-sm text-muted-foreground mb-4">Create a profile to get started.</p>
                  <Button onClick={handleCreateNew}>Create Profile</Button>
              </div>
          )}
          {profiles.map(profile => {
              const type = getType(profile);
              return (
                <Card key={profile.name} className="backdrop-blur-sm bg-background/50 hover:shadow-md transition-all">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-xl font-bold">{profile.name}</CardTitle>
                        <User className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-wrap gap-2 mb-4 mt-2">
                            <Badge variant={type === 'prod' ? 'destructive' : type === 'debug' ? 'secondary' : 'default'}>
                                {type.toUpperCase()}
                            </Badge>
                            {profile.selector?.tags?.map(t => (
                                t !== type && <Badge key={t} variant="outline" className="text-[10px]">{t}</Badge>
                            ))}
                        </div>
                        <div className="text-xs text-muted-foreground mb-4 font-mono bg-muted/50 p-1 rounded px-2">
                            {getServiceCount(profile)}
                        </div>
                        <div className="flex justify-end gap-2">
                            <Button variant="ghost" size="sm" onClick={() => handleEdit(profile)}>
                                <Edit className="h-3 w-3 mr-1"/> Edit
                            </Button>
                            <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-100/10" onClick={() => handleDelete(profile.name)}>
                                <Trash className="h-3 w-3 mr-1"/> Delete
                            </Button>
                        </div>
                    </CardContent>
                </Card>
              );
          })}
      </div>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
          <SheetContent className="sm:max-w-xl w-full">
              <SheetHeader className="mb-4">
                  <SheetTitle>{selectedProfile ? "Edit Profile" : "Create Profile"}</SheetTitle>
                  <SheetDescription>
                      Configure which services are available for this profile.
                  </SheetDescription>
              </SheetHeader>
              <div className="h-[calc(100vh-140px)]">
                <ProfileEditor
                    profile={selectedProfile}
                    onSave={handleSave}
                    onCancel={() => setIsSheetOpen(false)}
                />
              </div>
          </SheetContent>
      </Sheet>
    </div>
  );
}
