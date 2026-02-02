/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { Virtuoso } from "react-virtuoso";
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
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Search, Loader2, X, Plus, ChevronRight, ChevronDown } from "lucide-react";
import { toast } from "sonner";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ToolDefinition } from "@/lib/client";

/**
 * Represents a user profile configuration in the UI.
 */
export interface Profile {
    /** Unique identifier for the profile (usually same as name). */
    id: string; // name
    /** Display name of the profile. */
    name: string;
    /** Optional description of the profile's purpose. */
    description?: string;
    /** List of service names enabled for this profile. */
    services: string[]; // List of enabled service names
    /** The environment type (dev, prod, debug). */
    type: "dev" | "prod" | "debug";
    /** Additional tags for service selection. */
    additionalTags: string[];
    /** Optional secrets associated with the profile. */
    secrets?: Record<string, string>;
    /** Map of serviceName -> List of disabled tool names. */
    disabledTools?: Record<string, string[]>;
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
 * A sheet component for creating or editing a user profile.
 * Allows configuring profile details and selecting accessible services.
 *
 * @param props - The component props.
 * @param props.profile - The profile to edit, or null to create a new one.
 * @param props.open - Whether the editor sheet is open.
 * @param props.onOpenChange - Callback to toggle the sheet's open state.
 * @param props.onSave - Callback invoked when the profile is saved.
 * @returns The rendered profile editor component.
 */
export function ProfileEditor({ profile, open, onOpenChange, onSave }: ProfileEditorProps) {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [type, setType] = useState<"dev" | "prod" | "debug">("dev");
    const [additionalTags, setAdditionalTags] = useState<string[]>([]);
    const [newTagInput, setNewTagInput] = useState("");
    const [selectedServices, setSelectedServices] = useState<Set<string>>(new Set());
    const [availableServices, setAvailableServices] = useState<UpstreamServiceConfig[]>([]);
    const [searchQuery, setSearchQuery] = useState("");
    const [isLoadingServices, setIsLoadingServices] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const [expandedServices, setExpandedServices] = useState<Set<string>>(new Set());
    const [serviceTools, setServiceTools] = useState<Record<string, ToolDefinition[]>>({});
    const [disabledTools, setDisabledTools] = useState<Record<string, Set<string>>>({});

    // Load initial data when profile changes or opens
    useEffect(() => {
        if (open) {
            fetchServices();
            if (profile) {
                setName(profile.name);
                setDescription(profile.description || "");
                setType(profile.type);
                setAdditionalTags(profile.additionalTags || []);
                setSelectedServices(new Set(profile.services));

                // Load disabled tools
                const dt: Record<string, Set<string>> = {};
                if (profile.disabledTools) {
                    Object.entries(profile.disabledTools).forEach(([svc, tools]) => {
                        dt[svc] = new Set(tools);
                    });
                }
                setDisabledTools(dt);
            } else {
                // Reset for new profile
                setName("");
                setDescription("");
                setType("dev");
                setAdditionalTags([]);
                setSelectedServices(new Set());
                setDisabledTools({});
            }
            // Reset expanded state
            setExpandedServices(new Set());
            setServiceTools({});
        }
    }, [open, profile]);

    const toggleExpandService = async (svcName: string) => {
        const newExpanded = new Set(expandedServices);
        if (newExpanded.has(svcName)) {
            newExpanded.delete(svcName);
        } else {
            newExpanded.add(svcName);
            // Fetch tools if not already fetched
            if (!serviceTools[svcName]) {
                try {
                    const res = await apiClient.listTools();
                    // Filter by service ID (svcName matches ID usually)
                    const tools = res.tools.filter((t: ToolDefinition) => t.serviceId === svcName);
                    setServiceTools(prev => ({ ...prev, [svcName]: tools }));
                } catch (e) {
                    console.error("Failed to fetch tools", e);
                }
            }
        }
        setExpandedServices(newExpanded);
    };

    const toggleTool = (svcName: string, toolName: string, checked: boolean) => {
        const currentDisabled = disabledTools[svcName] || new Set();
        const newDisabled = new Set(currentDisabled);
        if (checked) {
            // Enabled -> Remove from disabled set
            newDisabled.delete(toolName);
        } else {
            // Disabled -> Add to disabled set
            newDisabled.add(toolName);
        }
        setDisabledTools(prev => ({ ...prev, [svcName]: newDisabled }));
    };

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
                const toolsConfig: Record<string, any> = {};
                if (disabledTools[svc]) {
                    disabledTools[svc].forEach(tName => {
                        toolsConfig[tName] = { disabled: true };
                    });
                }

                serviceConfig[svc] = {
                    enabled: true,
                    tools: toolsConfig
                };
            });

            const profileData = {
                name: name,
                selector: {
                    tags: [type, ...additionalTags] // Combine type and additional tags
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

    const addTag = () => {
        if (!newTagInput.trim()) return;
        if (additionalTags.includes(newTagInput.trim())) {
            setNewTagInput("");
            return;
        }
        setAdditionalTags([...additionalTags, newTagInput.trim()]);
        setNewTagInput("");
    };

    const removeTag = (tag: string) => {
        setAdditionalTags(additionalTags.filter(t => t !== tag));
    };

    const filteredServices = useMemo(() => {
        if (!searchQuery) return availableServices;
        return availableServices.filter(s =>
            s.name.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [availableServices, searchQuery]);

    const implicitlySelectedServices = useMemo(() => {
        const allProfileTags = new Set([type, ...additionalTags]);
        const implicitSet = new Set<string>();

        availableServices.forEach(svc => {
            if (svc.tags && svc.tags.some(t => allProfileTags.has(t))) {
                implicitSet.add(svc.name);
            }
        });
        return implicitSet;
    }, [availableServices, type, additionalTags]);

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
                            <Label>Access Tags</Label>
                            <div className="flex gap-2">
                                <Input
                                    value={newTagInput}
                                    onChange={(e) => setNewTagInput(e.target.value)}
                                    placeholder="Add tag (e.g. finance, hr)"
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter') {
                                            e.preventDefault();
                                            addTag();
                                        }
                                    }}
                                />
                                <Button type="button" size="icon" variant="outline" onClick={addTag}>
                                    <Plus className="h-4 w-4" />
                                </Button>
                            </div>
                            {additionalTags.length > 0 && (
                                <div className="flex flex-wrap gap-2 mt-2 p-2 bg-muted/20 rounded-md border">
                                    {additionalTags.map(tag => (
                                        <Badge key={tag} variant="secondary" className="pl-2 pr-1 py-1">
                                            {tag}
                                            <button
                                                onClick={() => removeTag(tag)}
                                                className="ml-1 hover:bg-muted rounded-full p-0.5"
                                            >
                                                <X className="h-3 w-3" />
                                            </button>
                                        </Badge>
                                    ))}
                                </div>
                            )}
                            <p className="text-[10px] text-muted-foreground">
                                Services matching these tags will be automatically available to this profile.
                            </p>
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

                            {isLoadingServices ? (
                                <div className="flex justify-center py-8">
                                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                </div>
                            ) : (
                                filteredServices.length === 0 ? (
                                    <div className="text-sm text-muted-foreground text-center py-4">No services found.</div>
                                ) : (
                                    // ⚡ BOLT: Randomized Selection from Top 5 High-Impact Targets
                                    // Optimized list rendering with virtualization for performance
                                    <Virtuoso
                                        style={{ height: "300px" }}
                                        data={filteredServices}
                                        itemContent={(_, svc) => {
                                            const isImplicit = implicitlySelectedServices.has(svc.name);
                                            const isExplicit = selectedServices.has(svc.name);
                                            const isSelected = isImplicit || isExplicit;

                                            return (
                                                <div key={svc.name} className="flex flex-col p-2 hover:bg-muted/50 rounded transition-colors pr-3">
                                                    <div className="flex items-start space-x-3">
                                                        <Checkbox
                                                            id={`svc-${svc.name}`}
                                                            checked={isSelected}
                                                            disabled={isImplicit}
                                                            onCheckedChange={(c) => toggleService(svc.name, !!c)}
                                                        />
                                                        <div className="grid gap-1.5 leading-none flex-1">
                                                            <div className="flex items-center gap-2">
                                                                <label
                                                                    htmlFor={`svc-${svc.name}`}
                                                                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                                                                >
                                                                    {svc.name}
                                                                </label>
                                                                {isImplicit && <Badge variant="secondary" className="text-[10px] h-4 px-1">Auto</Badge>}
                                                                {isSelected && (
                                                                    <Button
                                                                        variant="ghost"
                                                                        size="icon"
                                                                        className="h-4 w-4 ml-auto"
                                                                        onClick={() => toggleExpandService(svc.name)}
                                                                    >
                                                                        {expandedServices.has(svc.name) ? (
                                                                            <ChevronDown className="h-3 w-3" />
                                                                        ) : (
                                                                            <ChevronRight className="h-3 w-3" />
                                                                        )}
                                                                    </Button>
                                                                )}
                                                            </div>
                                                            <div className="flex flex-wrap gap-1 mt-1">
                                                                {svc.tags && svc.tags.map(tag => (
                                                                    <Badge key={tag} variant="outline" className="text-[9px] h-3 px-1">{tag}</Badge>
                                                                ))}
                                                            </div>
                                                            <p className="text-xs text-muted-foreground mt-0.5">
                                                                {svc.commandLineService ? "Command" : svc.httpService ? "HTTP" : "Remote"} • v{svc.version || "1.0.0"}
                                                            </p>
                                                        </div>
                                                    </div>

                                                    {/* Tool List */}
                                                    {isSelected && expandedServices.has(svc.name) && (
                                                        <div className="ml-8 mt-2 space-y-2 border-l pl-3">
                                                            {!serviceTools[svc.name] ? (
                                                                <div className="flex items-center text-xs text-muted-foreground">
                                                                    <Loader2 className="mr-2 h-3 w-3 animate-spin" /> Loading tools...
                                                                </div>
                                                            ) : serviceTools[svc.name].length === 0 ? (
                                                                <div className="text-xs text-muted-foreground">No tools found.</div>
                                                            ) : (
                                                                serviceTools[svc.name].map(tool => {
                                                                    const isDisabled = disabledTools[svc.name]?.has(tool.name);
                                                                    return (
                                                                        <div key={tool.name} className="flex items-center space-x-2">
                                                                            <Checkbox
                                                                                id={`tool-${svc.name}-${tool.name}`}
                                                                                checked={!isDisabled}
                                                                                onCheckedChange={(c) => toggleTool(svc.name, tool.name, !!c)}
                                                                                className="h-3 w-3"
                                                                            />
                                                                            <label
                                                                                htmlFor={`tool-${svc.name}-${tool.name}`}
                                                                                className="text-xs leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                                                                            >
                                                                                {tool.name}
                                                                            </label>
                                                                        </div>
                                                                    );
                                                                })
                                                            )}
                                                        </div>
                                                    )}
                                                </div>
                                            );
                                        }}
                                    />
                                )
                            )}
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
