/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import {
    Plus,
    Trash2,
    Edit2,
    Settings,
    Shield,
    Search,
    RefreshCw,
    X
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useToast } from "@/hooks/use-toast";
import { apiClient, ProfileDefinition } from "@/lib/client";
import { Textarea } from "@/components/ui/textarea";

/**
 * ProfileManager component.
 * @returns The rendered component.
 */
export function ProfileManager() {
    const [profiles, setProfiles] = useState<ProfileDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const { toast } = useToast();

    // Form state
    const [editingProfile, setEditingProfile] = useState<ProfileDefinition | null>(null);
    const [name, setName] = useState("");
    const [tags, setTags] = useState("");
    const [toolProperties, setToolProperties] = useState<{key: string, value: string}[]>([{key: "", value: ""}]);

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

    const handleOpenDialog = (profile?: ProfileDefinition) => {
        if (profile) {
            setEditingProfile(profile);
            setName(profile.name);
            setTags(profile.selector?.tags?.join(", ") || "");
            const props = profile.selector?.toolProperties || {};
            const propsArray = Object.entries(props).map(([k, v]) => ({ key: k, value: v }));
            if (propsArray.length === 0) propsArray.push({ key: "", value: "" });
            setToolProperties(propsArray);
        } else {
            setEditingProfile(null);
            setName("");
            setTags("");
            setToolProperties([{ key: "", value: "" }]);
        }
        setIsDialogOpen(true);
    };

    const handleSave = async () => {
        if (!name) {
            toast({
                title: "Validation Error",
                description: "Profile name is required.",
                variant: "destructive",
            });
            return;
        }

        const tagsArray = tags.split(",").map(t => t.trim()).filter(t => t);
        const propsRecord: Record<string, string> = {};
        toolProperties.forEach(p => {
            if (p.key) propsRecord[p.key] = p.value;
        });

        const profileData: ProfileDefinition = {
            name,
            selector: {
                tags: tagsArray,
                toolProperties: propsRecord
            }
        };

        try {
            if (editingProfile) {
                await apiClient.updateProfile(profileData);
                toast({ title: "Success", description: "Profile updated successfully." });
            } else {
                await apiClient.createProfile(profileData);
                toast({ title: "Success", description: "Profile created successfully." });
            }
            setIsDialogOpen(false);
            loadProfiles();
        } catch (error) {
            console.error(error);
            toast({
                title: "Error",
                description: "Failed to save profile.",
                variant: "destructive",
            });
        }
    };

    const handleDelete = async (profileName: string) => {
        if (!confirm(`Are you sure you want to delete profile "${profileName}"?`)) return;
        try {
            await apiClient.deleteProfile(profileName);
            toast({ title: "Success", description: "Profile deleted successfully." });
            loadProfiles();
        } catch (error) {
            toast({
                title: "Error",
                description: "Failed to delete profile.",
                variant: "destructive",
            });
        }
    };

    const addPropertyRow = () => {
        setToolProperties([...toolProperties, { key: "", value: "" }]);
    };

    const removePropertyRow = (index: number) => {
        const newProps = [...toolProperties];
        newProps.splice(index, 1);
        setToolProperties(newProps);
    };

    const updateProperty = (index: number, field: 'key' | 'value', value: string) => {
        const newProps = [...toolProperties];
        newProps[index][field] = value;
        setToolProperties(newProps);
    };

    const filteredProfiles = profiles.filter(p =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Execution Profiles</h3>
                    <p className="text-sm text-muted-foreground">
                        Manage tool selection profiles and policies.
                    </p>
                </div>
                <Button onClick={() => handleOpenDialog()}>
                    <Plus className="mr-2 h-4 w-4" /> New Profile
                </Button>
            </div>

             <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>{editingProfile ? "Edit Profile" : "Create Profile"}</DialogTitle>
                        <DialogDescription>
                            Configure tool selection criteria for this profile.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Profile Name</Label>
                            <Input
                                id="name"
                                placeholder="e.g. strict-mode"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                disabled={!!editingProfile} // Name is ID often, so disable edit if exists or handle rename carefully. Assuming ID=Name.
                            />
                            {editingProfile && <p className="text-xs text-muted-foreground">Profile names cannot be changed once created.</p>}
                        </div>

                        <div className="grid gap-2">
                            <Label htmlFor="tags">Selector Tags</Label>
                            <Input
                                id="tags"
                                placeholder="safe, verified, production (comma separated)"
                                value={tags}
                                onChange={(e) => setTags(e.target.value)}
                            />
                        </div>

                         <div className="grid gap-2">
                            <Label>Tool Properties Matcher</Label>
                            <div className="border rounded-md p-4 space-y-2 bg-muted/20">
                                {toolProperties.map((prop, index) => (
                                    <div key={index} className="flex gap-2 items-center">
                                        <Input
                                            placeholder="Property Key (e.g. read_only)"
                                            className="flex-1"
                                            value={prop.key}
                                            onChange={(e) => updateProperty(index, 'key', e.target.value)}
                                        />
                                        <span className="text-muted-foreground">=</span>
                                        <Input
                                            placeholder="Value (e.g. true)"
                                            className="flex-1"
                                            value={prop.value}
                                            onChange={(e) => updateProperty(index, 'value', e.target.value)}
                                        />
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => removePropertyRow(index)}
                                            disabled={toolProperties.length === 1 && index === 0 && !prop.key && !prop.value}
                                        >
                                            <X className="h-4 w-4" />
                                        </Button>
                                    </div>
                                ))}
                                <Button variant="outline" size="sm" onClick={addPropertyRow} className="mt-2">
                                    <Plus className="h-3 w-3 mr-1" /> Add Property
                                </Button>
                            </div>
                            <p className="text-xs text-muted-foreground">Tools must match ALL specified properties to be included.</p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleSave}>Save Profile</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

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
                                <Shield className="h-8 w-8 opacity-20" />
                                <p>No profiles found.</p>
                            </div>
                        ) : (
                             <div className="divide-y">
                                {filteredProfiles.map((profile) => (
                                    <div key={profile.name} className="flex items-center justify-between p-4 hover:bg-muted/30 transition-colors group">
                                        <div className="flex items-center gap-4">
                                            <div className="bg-primary/10 p-2 rounded-full text-primary">
                                                <Settings className="h-4 w-4" />
                                            </div>
                                            <div>
                                                <h4 className="font-medium text-sm">{profile.name}</h4>
                                                <div className="flex gap-2 mt-1 flex-wrap">
                                                    {profile.selector?.tags?.map(tag => (
                                                        <Badge key={tag} variant="secondary" className="text-[10px] h-5 px-1">
                                                            tag: {tag}
                                                        </Badge>
                                                    ))}
                                                    {Object.entries(profile.selector?.toolProperties || {}).map(([k, v]) => (
                                                        <Badge key={k} variant="outline" className="text-[10px] h-5 px-1 font-mono text-muted-foreground">
                                                            {k}={v}
                                                        </Badge>
                                                    ))}
                                                </div>
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Button variant="ghost" size="icon" onClick={() => handleOpenDialog(profile)}>
                                                <Edit2 className="h-4 w-4 text-muted-foreground" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="text-destructive/70 hover:text-destructive hover:bg-destructive/10" onClick={() => handleDelete(profile.name)}>
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </CardContent>
            </Card>
        </div>
    );
}
