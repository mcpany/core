/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { User, Plus, Trash, Edit, RefreshCw } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription
} from "@/components/ui/sheet";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { toast } from "sonner";
import { ProfileEditor } from "@/components/profiles/profile-editor";

interface ProfileUI {
    id: string;
    name: string;
    description: string;
    services: string[];
    type: "dev" | "prod" | "debug";
    raw: any; // Store the full backend object
}

/**
 * ProfilesPage component.
 * @returns The rendered component.
 */
export default function ProfilesPage() {
  const [profiles, setProfiles] = useState<ProfileUI[]>([]);
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState<any | null>(null);

  const fetchData = async () => {
      setIsLoading(true);
      try {
          const [profilesData, servicesData] = await Promise.all([
              apiClient.listProfiles(),
              apiClient.listServices()
          ]);

          // Map backend ProfileDefinition to UI Profile
          const mapped: ProfileUI[] = profilesData.map((p: any) => ({
              id: p.name, // Use name as ID
              name: p.name,
              description: "",
              services: p.serviceConfig ? Object.keys(p.serviceConfig).filter(k => p.serviceConfig[k].enabled) : [],
              type: (p.selector?.tags?.find((t: string) => ["dev", "prod", "debug"].includes(t)) as "dev" | "prod" | "debug") || "dev",
              raw: p
          }));
          setProfiles(mapped);
          setServices(Array.isArray(servicesData) ? servicesData : (servicesData as any).services || []);
      } catch (error) {
          console.error("Failed to fetch data", error);
          toast.error("Failed to fetch profiles or services");
      } finally {
          setIsLoading(false);
      }
  };

  useEffect(() => {
      fetchData();
  }, []);

  const handleCreateOpen = () => {
      setSelectedProfile(null);
      setIsSheetOpen(true);
  };

  const handleEditOpen = (profile: ProfileUI) => {
      setSelectedProfile(profile.raw);
      setIsSheetOpen(true);
  };

  const handleSave = async (profileData: any) => {
      try {
          if (selectedProfile) {
              // Update
              await apiClient.updateProfile(profileData);
              toast.success("Profile updated successfully");
          } else {
              // Create
              await apiClient.createProfile(profileData);
              toast.success("Profile created successfully");
          }
          setIsSheetOpen(false);
          fetchData();
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
          fetchData();
      } catch (error) {
           console.error("Failed to delete profile", error);
           toast.error("Failed to delete profile");
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] overflow-hidden flex flex-col">
      <div className="flex items-center justify-between">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
            <p className="text-muted-foreground">Manage access control and service visibility.</p>
        </div>
        <div className="flex gap-2">
            <Button variant="ghost" size="icon" onClick={fetchData}>
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
            <Button onClick={handleCreateOpen}>
                <Plus className="mr-2 h-4 w-4" /> Create Profile
            </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 pb-8">
              {!isLoading && profiles.length === 0 && (
                  <div className="col-span-3 text-center p-12 border rounded-lg border-dashed bg-muted/20">
                      <User className="h-12 w-12 mx-auto text-muted-foreground opacity-20 mb-4" />
                      <h3 className="text-lg font-medium">No profiles found</h3>
                      <p className="text-sm text-muted-foreground mb-4">Create a profile to control which services are exposed.</p>
                      <Button onClick={handleCreateOpen}>Create Profile</Button>
                  </div>
              )}
              {profiles.map(profile => (
                  <Card key={profile.id} className="backdrop-blur-sm bg-background/50 hover:shadow-md transition-all flex flex-col">
                      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                          <CardTitle className="text-xl font-bold truncate pr-4" title={profile.name}>{profile.name}</CardTitle>
                          <User className="h-4 w-4 text-muted-foreground shrink-0" />
                      </CardHeader>
                      <CardContent className="flex-1 flex flex-col">
                          <div className="flex flex-wrap gap-2 mb-4">
                              <Badge variant={profile.type === 'prod' ? 'destructive' : profile.type === 'debug' ? 'secondary' : 'default'}>
                                  {profile.type.toUpperCase()}
                              </Badge>
                          </div>
                          <div className="text-sm text-muted-foreground mb-6 flex-1">
                              {profile.services.length === 0 ? (
                                  <span className="text-yellow-600 dark:text-yellow-500">No services enabled</span>
                              ) : (
                                  <div className="flex flex-wrap gap-1">
                                      {profile.services.slice(0, 5).map(s => (
                                          <Badge key={s} variant="outline" className="font-normal text-xs">{s}</Badge>
                                      ))}
                                      {profile.services.length > 5 && (
                                          <Badge variant="outline" className="font-normal text-xs">+{profile.services.length - 5}</Badge>
                                      )}
                                  </div>
                              )}
                          </div>
                          <div className="flex justify-end gap-2 mt-auto">
                              <Button variant="ghost" size="sm" onClick={() => handleEditOpen(profile)}>
                                  <Edit className="h-3 w-3 mr-1"/> Edit
                              </Button>
                              <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/50" onClick={() => handleDelete(profile.name)}>
                                  <Trash className="h-3 w-3 mr-1"/> Delete
                              </Button>
                          </div>
                      </CardContent>
                  </Card>
              ))}
          </div>
      </div>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
          <SheetContent className="sm:max-w-md w-full">
              <SheetHeader>
                  <SheetTitle>{selectedProfile ? "Edit Profile" : "New Profile"}</SheetTitle>
                  <SheetDescription>
                      Configure which services are accessible via this profile.
                  </SheetDescription>
              </SheetHeader>
              <div className="h-[calc(100vh-120px)] mt-4">
                  <ProfileEditor
                      profile={selectedProfile}
                      services={services}
                      onSave={handleSave}
                      onCancel={() => setIsSheetOpen(false)}
                  />
              </div>
          </SheetContent>
      </Sheet>
    </div>
  );
}
