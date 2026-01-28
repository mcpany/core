/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, ProfileDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus, Edit2, Trash2, Loader2, RefreshCw } from "lucide-react";
import { ProfileDialog } from "./profile-dialog";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";

export function ProfilesList() {
  const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState<ProfileDefinition | null>(null);
  const { toast } = useToast();

  const fetchProfiles = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listProfiles();
      // listProfiles returns { profiles: [...] } based on proto
      setProfiles(res.profiles || []);
    } catch (e) {
      console.error("Failed to fetch profiles", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to fetch profiles.",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfiles();
  }, []);

  const handleEdit = (profile: ProfileDefinition) => {
    setSelectedProfile(profile);
    setDialogOpen(true);
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Are you sure you want to delete profile "${name}"?`)) return;
    try {
      await apiClient.deleteProfile(name);
      toast({ title: "Profile Deleted", description: `Profile ${name} has been deleted.` });
      fetchProfiles();
    } catch (e) {
      console.error("Failed to delete profile", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to delete profile.",
      });
    }
  };

  const handleSave = async () => {
    setDialogOpen(false);
    setSelectedProfile(null);
    fetchProfiles();
  };

  return (
    <div className="space-y-4 h-full">
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Execution Profiles</CardTitle>
              <CardDescription>Manage environment variables, secrets, and role-based access for execution contexts.</CardDescription>
            </div>
            <Button onClick={() => { setSelectedProfile(null); setDialogOpen(true); }}>
              <Plus className="mr-2 h-4 w-4" /> New Profile
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex justify-center p-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Required Roles</TableHead>
                  <TableHead>Parents</TableHead>
                  <TableHead>Secrets</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {profiles.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                      No profiles found. Create one to get started.
                    </TableCell>
                  </TableRow>
                )}
                {profiles.map((profile) => (
                  <TableRow key={profile.name}>
                    <TableCell className="font-medium">{profile.name}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {profile.requiredRoles?.map((role) => (
                          <Badge key={role} variant="secondary" className="text-xs">
                            {role}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {profile.parentProfileIds?.map((pid) => (
                          <Badge key={pid} variant="outline" className="text-xs">
                            {pid}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {Object.keys(profile.secrets || {}).length} variables
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="icon" onClick={() => handleEdit(profile)}>
                          <Edit2 className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => handleDelete(profile.name)}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ProfileDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        profile={selectedProfile}
        onSave={handleSave}
      />
    </div>
  );
}
