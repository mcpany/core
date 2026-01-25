/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetTitle
} from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Search, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

/**
 * Represents a user profile configuration in the UI.
 */
export interface Profile {
    /** The unique identifier for the profile (often same as name). */
    id: string;
    /** The display name of the profile. */
    name: string;
    /** An optional description of the profile's purpose. */
    description?: string;
    /** A list of service names enabled for this profile. */
    services: string[];
    /** The environment type associated with this profile. */
    type: "dev" | "prod" | "debug";
    /** Optional key-value pairs of secrets associated with the profile. */
    secrets?: Record<string, string>;
}

interface ProfileData {
    name: string;
    selector: { tags: string[] };
    serviceConfig: Record<string, unknown>;
    secrets: Record<string, string>;
}

interface ProfileEditorProps {
    profile: Profile | null; // Null means new profile
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSave: (profileData: ProfileData) => Promise<void>;
}

/**
 * ProfileEditor component.
 * @param props - The component props.
 * @param props.profile - The profile property.
 * @param props.open - Whether the component is open.
 * @param props.onOpenChange - Whether the component is open.
 * @param props.onSave - The onSave property.
 * @returns The rendered component.
 */
export function ProfileEditor({ profile, open, onOpenChange, onSave }: ProfileEditorProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [type, setType] = useState<"dev" | "prod" | "debug">("dev");
    const [selectedServices, setSelectedServices] = useState<Set<string>>(new Set());
    const [availableServices, setAvailableServices] = useState<UpstreamServiceConfig[]>([]);
    const [searchQuery, setSearchQuery] = useState("");
    const [isLoadingServices, setIsLoadingServices] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    // Load initial data when profile changes or opens
    useEffect(() => {
        if (open) {
            fetchServices();
            if (profile) {
                setName(profile.name);
                setDescription(profile.description || "");
                setType(profile.type);
                setSelectedServices(new Set(profile.services));
            } else {
                // Reset for new profile
                setName("");
                setDescription("");
                setType("dev");
                setSelectedServices(new Set());
            }
        }
    }, [open, profile]);

    const fetchServices = async () => {
        setIsLoadingServices(true);
        try {
            const data = await apiClient.listServices();
            setAvailableServices(data);
        } catch (error) {
            console.error("Failed to load services", error);
            toast.error("Failed to load available services");
        } finally {
            setIsLoadingServices(false);
        }
    };

    const handleSave = async () => {
        if (!name.trim()) {
            toast.error("Profile name is required");
            return;
        }

        setIsSaving(true);
        try {
            // Construct backend ProfileDefinition
            const serviceConfig: Record<string, any> = {};
            selectedServices.forEach(svc => {
                serviceConfig[svc] = {}; // Default config
            });

            const profileData = {
                name: name,
                selector: {
                    tags: [type] // Use type as tag for now
                },
                serviceConfig: serviceConfig,
                secrets: profile?.secrets || {} // Preserve secrets if any
            };

            await onSave(profileData);
            // toast success handled by parent or here? Parent handles it better for context
        } catch (error) {
            console.error(error);
            // Parent should handle error toast usually
        } finally {
            setIsSaving(false);
        }
    };

    const filteredServices = useMemo(() => {
        if (!searchQuery) return availableServices;
        return availableServices.filter(s =>
            s.name.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [availableServices, searchQuery]);

    const toggleService = (svcName: string, checked: boolean) => {
        const newSet = new Set(selectedServices);
        if (checked) {
            newSet.add(svcName);
        } else {
            newSet.delete(svcName);
        }
        setSelectedServices(newSet);
    };

    const handleSelectAll = () => {
        const newSet = new Set(selectedServices);
        filteredServices.forEach(s => newSet.add(s.name));
        setSelectedServices(newSet);
    };

    const handleDeselectAll = () => {
        const newSet = new Set(selectedServices);
        filteredServices.forEach(s => newSet.delete(s.name));
        setSelectedServices(newSet);
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="w-[400px] sm:w-[540px] flex flex-col h-full">
                <SheetHeader>
                    <SheetTitle>{profile ? `Edit Profile: ${profile.name}` : "Create New Profile"}</SheetTitle>
                    <SheetDescription>
                        Configure profile details and service access levels.
                    </SheetDescription>
                </SheetHeader>

                <div className="flex-1 py-6 space-y-6 overflow-y-auto pr-2">
                    <div className="space-y-4">
                        <div className="grid gap-2">
                            <Label htmlFor="name">Profile Name</Label>
                            <Input
                                id="name"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                disabled={!!profile} // Read-only if editing
                                placeholder="e.g. staging-user"
                            />
                            {profile && <p className="text-[10px] text-muted-foreground">Profile ID cannot be changed once created.</p>}
                        </div>

                        <div className="grid gap-2">
                            <Label htmlFor="type">Environment Type</Label>
                            <Select value={type} onValueChange={(v: string) => setType(v as "dev" | "prod" | "debug")}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="dev">Development</SelectItem>
                                    <SelectItem value="prod">Production</SelectItem>
                                    <SelectItem value="debug">Debug / Test</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="grid gap-2">
                             <Label htmlFor="desc">Description</Label>
                             <Input
                                id="desc"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                placeholder="Optional description"
                            />
                        </div>
                    </div>

                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label>Service Access</Label>
                            <Badge variant="outline">{selectedServices.size} Selected</Badge>
                        </div>
                        <div className="bg-muted/30 p-3 rounded-lg border space-y-3">
                            <div className="flex items-center gap-2">
                                <Search className="h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search services..."
                                    className="h-8 bg-background"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                />
                            </div>
                            <div className="flex justify-end gap-2 text-xs">
                                <button onClick={handleSelectAll} className="text-primary hover:underline">Select All</button>
                                <button onClick={handleDeselectAll} className="text-muted-foreground hover:underline">None</button>
                            </div>

                            <ScrollArea className="h-[300px] pr-3">
                                {isLoadingServices ? (
                                    <div className="flex justify-center py-8">
                                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        {filteredServices.length === 0 && (
                                            <div className="text-sm text-muted-foreground text-center py-4">No services found.</div>
                                        )}
                                        {filteredServices.map(svc => (
                                            <div key={svc.name} className="flex items-start space-x-3 p-2 hover:bg-muted/50 rounded transition-colors">
                                                <Checkbox
                                                    id={`svc-${svc.name}`}
                                                    checked={selectedServices.has(svc.name)}
                                                    onCheckedChange={(c) => toggleService(svc.name, !!c)}
                                                />
                                                <div className="grid gap-1.5 leading-none">
                                                    <label
                                                        htmlFor={`svc-${svc.name}`}
                                                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                                                    >
                                                        {svc.name}
                                                    </label>
                                                    <p className="text-xs text-muted-foreground">
                                                        {svc.commandLineService ? "Command" : svc.httpService ? "HTTP" : "Remote"} • v{svc.version || "1.0.0"}
                                                    </p>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </ScrollArea>
                        </div>
                    </div>
                </div>

                <SheetFooter>
                     <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                     <Button onClick={handleSave} disabled={isSaving}>
                         {isSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                         Save Profile
                     </Button>
                </SheetFooter>
            </SheetContent>
        </Sheet>
    );
}
