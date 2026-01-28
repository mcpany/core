/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient, ProfileDefinition } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2, Plus, UserCircle, Tag, Shield, Trash2, Download, MoreHorizontal, Settings } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogDescription
} from "@/components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ProfileEditor } from "@/components/settings/profile-editor";

/**
 * ProfilesManager component.
 * Allows managing execution profiles.
 */
export function ProfilesManager() {
    const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const { toast } = useToast();
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [selectedProfile, setSelectedProfile] = useState<ProfileDefinition | null>(null);

    const fetchProfiles = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiClient.listProfiles();
            // Ensure array
            const list = Array.isArray(data) ? data : [];
            setProfiles(list);
        } catch (e) {
            console.error("Failed to fetch profiles", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to load profiles."
            });
        } finally {
            setLoading(false);
        }
    }, [toast]);

    useEffect(() => {
        fetchProfiles();
    }, [fetchProfiles]);

    const handleDelete = async (name: string) => {
        if (!confirm(`Are you sure you want to delete profile "${name}"?`)) return;
        try {
            await apiClient.deleteProfile(name);
            toast({
                title: "Profile Deleted",
                description: `Profile ${name} has been removed.`
            });
            fetchProfiles();
        } catch (e) {
            console.error("Failed to delete profile", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to delete profile."
            });
        }
    };

    const handleExport = (profile: ProfileDefinition) => {
        const data = JSON.stringify(profile, null, 2);
        const blob = new Blob([data], { type: "application/json" });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = `${profile.name}-profile.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        toast({
            title: "Profile Exported",
            description: `Profile ${profile.name} has been downloaded.`
        });
    };

    const handleEdit = (profile: ProfileDefinition) => {
        setSelectedProfile(profile);
        setIsEditorOpen(true);
    };

    const handleCreate = () => {
        setSelectedProfile(null);
        setIsEditorOpen(true);
    };

    const handleEditorSubmit = async (profile: ProfileDefinition) => {
        try {
            if (selectedProfile) {
                // Update
                await apiClient.updateProfile(profile);
                toast({ title: "Profile Updated", description: "Profile configuration saved." });
            } else {
                // Create
                await apiClient.createProfile(profile);
                toast({ title: "Profile Created", description: "New profile created successfully." });
            }
            setIsEditorOpen(false);
            fetchProfiles();
        } catch (e) {
            console.error("Failed to save profile", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to save profile configuration."
            });
        }
    };

    if (loading) {
        return (
             <div className="flex items-center justify-center h-48">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                     <h3 className="text-lg font-medium">Execution Profiles</h3>
                     <p className="text-sm text-muted-foreground">
                        Define sets of configuration, secrets, and tools for different environments.
                     </p>
                </div>
                <Button onClick={handleCreate}>
                    <Plus className="mr-2 h-4 w-4" /> New Profile
                </Button>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {profiles.length === 0 && (
                    <div className="col-span-full text-center py-12 border rounded-lg border-dashed text-muted-foreground">
                        No profiles found. Create one to get started.
                    </div>
                )}
                {profiles.map((profile) => (
                    <Card key={profile.name} className="relative group overflow-hidden transition-all hover:shadow-md">
                        <CardHeader className="pb-2">
                             <div className="flex items-center justify-between">
                                <CardTitle className="text-base font-medium flex items-center gap-2">
                                    <UserCircle className="h-4 w-4 text-primary" />
                                    {profile.name}
                                </CardTitle>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <span className="sr-only">Open menu</span>
                                            <MoreHorizontal className="h-4 w-4" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                        <DropdownMenuItem onClick={() => handleEdit(profile)}>
                                            <Settings className="mr-2 h-4 w-4" /> Edit
                                        </DropdownMenuItem>
                                        <DropdownMenuItem onClick={() => handleExport(profile)}>
                                            <Download className="mr-2 h-4 w-4" /> Export
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => handleDelete(profile.name)} className="text-destructive">
                                            <Trash2 className="mr-2 h-4 w-4" /> Delete
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                             </div>
                             <CardDescription>
                                {Object.keys(profile.serviceConfig || {}).length} services configured
                             </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            {/* Tags */}
                            {profile.selector?.tags && profile.selector.tags.length > 0 && (
                                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <Tag className="h-3 w-3" />
                                    <div className="flex flex-wrap gap-1">
                                        {profile.selector.tags.map(tag => (
                                            <Badge key={tag} variant="secondary" className="px-1 py-0 h-4 text-[10px]">
                                                {tag}
                                            </Badge>
                                        ))}
                                    </div>
                                </div>
                            )}

                             {/* Roles */}
                             {profile.requiredRoles && profile.requiredRoles.length > 0 && (
                                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <Shield className="h-3 w-3" />
                                    <div className="flex flex-wrap gap-1">
                                        {profile.requiredRoles.map(role => (
                                            <Badge key={role} variant="outline" className="px-1 py-0 h-4 text-[10px]">
                                                {role}
                                            </Badge>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                ))}
            </div>

            <Dialog open={isEditorOpen} onOpenChange={setIsEditorOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{selectedProfile ? `Edit ${selectedProfile.name}` : "New Profile"}</DialogTitle>
                        <DialogDescription>
                            Configure execution profile settings.
                        </DialogDescription>
                    </DialogHeader>
                    <ProfileEditor
                        profile={selectedProfile}
                        onCancel={() => setIsEditorOpen(false)}
                        onSubmit={handleEditorSubmit}
                    />
                </DialogContent>
            </Dialog>
        </div>
    );
}
