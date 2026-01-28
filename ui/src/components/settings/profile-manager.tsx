/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import {
    Plus,
    Trash2,
    Pencil,
    Search,
    RefreshCw,
    User,
    Tag,
    Shield
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useToast } from "@/hooks/use-toast";
import { apiClient, ProfileDefinition } from "@/lib/client";

/**
 * ProfileManager component.
 * @returns The rendered component.
 */
export function ProfileManager() {
    const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
    const [editingProfile, setEditingProfile] = useState<ProfileDefinition | null>(null);
    const { toast } = useToast();

    // Form state
    const [profileName, setProfileName] = useState("");
    const [requiredRoles, setRequiredRoles] = useState("");
    const [tags, setTags] = useState("");
    const [parentProfiles, setParentProfiles] = useState("");

    useEffect(() => {
        loadProfiles();
    }, []);

    const loadProfiles = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listProfiles();
            setProfiles(data);
        } catch (error) {
            console.error("Failed to load profiles", error);
            toast({
                title: "Error",
                description: "Failed to load profiles.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSaveProfile = async () => {
        if (!profileName) {
            toast({
                title: "Validation Error",
                description: "Profile name is required.",
                variant: "destructive",
            });
            return;
        }

        try {
            const profile: ProfileDefinition = {
                name: profileName,
                requiredRoles: requiredRoles.split(",").map(s => s.trim()).filter(Boolean),
                parentProfileIds: parentProfiles.split(",").map(s => s.trim()).filter(Boolean),
                selector: {
                    tags: tags.split(",").map(s => s.trim()).filter(Boolean),
                    toolProperties: {}
                },
                serviceConfig: {},
                secrets: {}
            };

            if (editingProfile) {
                // Update
                // Note: The API might expect the name in the path to identify the resource,
                // and the body to contain the updates.
                // If renaming is allowed, we might need to handle ID/Name consistency.
                // For now, assuming name is the ID.
                await apiClient.updateProfile(profile);
                toast({ title: "Success", description: "Profile updated successfully." });
            } else {
                // Create
                await apiClient.createProfile(profile);
                toast({ title: "Success", description: "Profile created successfully." });
            }

            setIsAddDialogOpen(false);
            setIsEditDialogOpen(false);
            resetForm();
            loadProfiles();
        } catch (error) {
            console.error("Failed to save profile", error);
            toast({
                title: "Error",
                description: "Failed to save profile.",
                variant: "destructive",
            });
        }
    };

    const handleDeleteProfile = async (name: string) => {
        if (!confirm(`Are you sure you want to delete profile "${name}"?`)) return;
        try {
            await apiClient.deleteProfile(name);
            toast({
                title: "Success",
                description: "Profile deleted successfully.",
            });
            loadProfiles();
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to delete profile.",
                variant: "destructive",
            });
        }
    };

    const openEdit = (profile: ProfileDefinition) => {
        setEditingProfile(profile);
        setProfileName(profile.name);
        setRequiredRoles(profile.requiredRoles?.join(", ") || "");
        setTags(profile.selector?.tags?.join(", ") || "");
        setParentProfiles(profile.parentProfileIds?.join(", ") || "");
        setIsEditDialogOpen(true);
    };

    const openAdd = () => {
        setEditingProfile(null);
        resetForm();
        setIsAddDialogOpen(true);
    };

    const resetForm = () => {
        setProfileName("");
        setRequiredRoles("");
        setTags("");
        setParentProfiles("");
    };

    const filteredProfiles = profiles.filter(p =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const renderDialogContent = (title: string, desc: string) => (
        <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
                <DialogTitle>{title}</DialogTitle>
                <DialogDescription>{desc}</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                    <Label htmlFor="name">Profile Name</Label>
                    <Input
                        id="name"
                        placeholder="e.g. development"
                        value={profileName}
                        onChange={(e) => setProfileName(e.target.value)}
                        disabled={!!editingProfile} // Name is ID usually
                    />
                </div>
                <div className="grid gap-2">
                    <Label htmlFor="roles">Required Roles (comma separated)</Label>
                    <Input
                        id="roles"
                        placeholder="e.g. admin, developer"
                        value={requiredRoles}
                        onChange={(e) => setRequiredRoles(e.target.value)}
                    />
                </div>
                <div className="grid gap-2">
                    <Label htmlFor="tags">Selector Tags (comma separated)</Label>
                    <Input
                        id="tags"
                        placeholder="e.g. env:dev, team:backend"
                        value={tags}
                        onChange={(e) => setTags(e.target.value)}
                    />
                </div>
                <div className="grid gap-2">
                    <Label htmlFor="parents">Parent Profiles (comma separated)</Label>
                    <Input
                        id="parents"
                        placeholder="e.g. base-profile"
                        value={parentProfiles}
                        onChange={(e) => setParentProfiles(e.target.value)}
                    />
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" onClick={() => { setIsAddDialogOpen(false); setIsEditDialogOpen(false); }}>Cancel</Button>
                <Button onClick={handleSaveProfile}>Save Profile</Button>
            </DialogFooter>
        </DialogContent>
    );

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Execution Profiles</h3>
                    <p className="text-sm text-muted-foreground">
                        Manage configuration profiles and access controls.
                    </p>
                </div>
                <Button onClick={openAdd}>
                    <Plus className="mr-2 h-4 w-4" /> Create Profile
                </Button>
            </div>

            <Card className="flex-1 flex flex-col overflow-hidden bg-background/50 backdrop-blur-sm border-muted/50">
                <CardHeader className="p-4 border-b bg-muted/20">
                     <div className="relative">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search profiles..."
                            className="pl-8 bg-background max-w-sm"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                </CardHeader>
                <CardContent className="p-0 flex-1 overflow-hidden">
                    <ScrollArea className="h-full">
                        {loading ? (
                            <div className="flex items-center justify-center h-40 text-muted-foreground gap-2">
                                <RefreshCw className="h-4 w-4 animate-spin" /> Loading profiles...
                            </div>
                        ) : filteredProfiles.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-40 text-muted-foreground gap-2">
                                <User className="h-8 w-8 opacity-20" />
                                <p>No profiles found.</p>
                            </div>
                        ) : (
                            <div className="divide-y">
                                {filteredProfiles.map((profile) => (
                                    <ProfileItem key={profile.name} profile={profile} onEdit={openEdit} onDelete={handleDeleteProfile} />
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </CardContent>
            </Card>

            <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
                {renderDialogContent("Create Profile", "Define a new execution profile.")}
            </Dialog>

            <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
                {renderDialogContent("Edit Profile", "Update profile configuration.")}
            </Dialog>
        </div>
    );
}

function ProfileItem({ profile, onEdit, onDelete }: { profile: ProfileDefinition; onEdit: (p: ProfileDefinition) => void; onDelete: (name: string) => void }) {
    return (
        <div className="flex items-center justify-between p-4 hover:bg-muted/30 transition-colors group">
            <div className="flex items-center gap-4">
                <div className="bg-primary/10 p-2 rounded-full text-primary">
                    <User className="h-4 w-4" />
                </div>
                <div>
                    <div className="flex items-center gap-2">
                        <h4 className="font-medium text-sm">{profile.name}</h4>
                        {profile.requiredRoles?.map(role => (
                            <Badge key={role} variant="outline" className="text-[10px] h-5 font-mono flex gap-1">
                                <Shield className="h-3 w-3" /> {role}
                            </Badge>
                        ))}
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                        {profile.selector?.tags?.map(tag => (
                            <span key={tag} className="text-xs text-muted-foreground bg-muted px-1.5 rounded-sm flex items-center gap-1">
                                <Tag className="h-3 w-3" /> {tag}
                            </span>
                        ))}
                        {(!profile.selector?.tags || profile.selector.tags.length === 0) && (
                            <span className="text-xs text-muted-foreground italic">No tags</span>
                        )}
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-2">
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onEdit(profile)}>
                    <Pencil className="h-4 w-4 text-muted-foreground" />
                </Button>
                 <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive/70 hover:text-destructive hover:bg-destructive/10" onClick={() => onDelete(profile.name)}>
                    <Trash2 className="h-4 w-4" />
                </Button>
            </div>
        </div>
    );
}
