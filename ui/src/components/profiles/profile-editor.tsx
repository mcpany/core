/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useMemo } from "react";
import { apiClient, UpstreamServiceConfig, ToolDefinition } from "@/lib/client";
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
import { Search, Loader2, ChevronDown, ChevronRight, Check } from "lucide-react";
import { toast } from "sonner";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";

export interface ProfileServiceConfig {
    enabled: boolean;
    allowedTools?: string[];
    blockedTools?: string[];
}

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
    /** Service configuration for this profile. */
    services: Record<string, ProfileServiceConfig>;
    /** The environment type (dev, prod, debug). */
    type: "dev" | "prod" | "debug";
    /** Optional secrets associated with the profile. */
    secrets?: Record<string, string>;
}

interface ProfileData {
    name: string;
    selector: { tags: string[] };
    serviceConfig: Record<string, any>;
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
    const [services, setServices] = useState<Record<string, ProfileServiceConfig>>({});
    const [availableServices, setAvailableServices] = useState<UpstreamServiceConfig[]>([]);
    const [tools, setTools] = useState<ToolDefinition[]>([]);
    const [expandedServices, setExpandedServices] = useState<Set<string>>(new Set());
    const [searchQuery, setSearchQuery] = useState("");
    const [isLoadingServices, setIsLoadingServices] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    // Load initial data when profile changes or opens
    useEffect(() => {
        if (open) {
            fetchServicesAndTools();
            if (profile) {
                setName(profile.name);
                setDescription(profile.description || "");
                setType(profile.type);
                setServices(profile.services || {});
            } else {
                // Reset for new profile
                setName("");
                setDescription("");
                setType("dev");
                setServices({});
            }
        }
    }, [open, profile]);

    const fetchServicesAndTools = async () => {
        setIsLoadingServices(true);
        try {
            const [servicesData, toolsData] = await Promise.all([
                apiClient.listServices(),
                apiClient.listTools()
            ]);
            setAvailableServices(servicesData);
            setTools(toolsData.tools);
        } catch (error) {
            console.error("Failed to load data", error);
            toast.error("Failed to load services or tools");
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
            // Filter only enabled services or those with specific config
            const serviceConfig: Record<string, any> = {};
            Object.entries(services).forEach(([svcName, config]) => {
                if (config.enabled) {
                    serviceConfig[svcName] = {
                        enabled: true,
                        allowed_tools: config.allowedTools,
                        blocked_tools: config.blockedTools
                    };
                }
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
        setServices(prev => {
            const next = { ...prev };
            if (checked) {
                next[svcName] = { enabled: true };
            } else {
                delete next[svcName];
            }
            return next;
        });
    };

    const toggleTool = (svcName: string, toolName: string, checked: boolean) => {
        setServices(prev => {
            const next = { ...prev };
            const currentConfig = next[svcName] || { enabled: true };

            // Start with current allowed tools or ALL tools if undefined
            // If undefined, it means "All". If we uncheck one, we must populate the list with all OTHERS.

            let newAllowed: string[];

            if (!currentConfig.allowedTools) {
                // Currently "All". If checking, nothing changes.
                // If unchecking (checked=false), we need to populate with ALL EXCEPT this one.
                if (!checked) {
                    const allServiceTools = tools.filter(t => t.serviceId === svcName).map(t => t.name);
                    newAllowed = allServiceTools.filter(t => t !== toolName);
                } else {
                    return next; // Already allowed
                }
            } else {
                // Currently explicit list
                if (checked) {
                    newAllowed = [...currentConfig.allowedTools, toolName];
                } else {
                    newAllowed = currentConfig.allowedTools.filter(t => t !== toolName);
                }
            }

            next[svcName] = { ...currentConfig, allowedTools: newAllowed };
            return next;
        });
    };

    const getServiceTools = (svcName: string) => {
        return tools.filter(t => t.serviceId === svcName || t.name.startsWith(svcName + "."));
    };

    const handleSelectAll = () => {
        setServices(prev => {
            const next = { ...prev };
            filteredServices.forEach(s => {
                next[s.name] = { enabled: true };
            });
            return next;
        });
    };

    const handleDeselectAll = () => {
        setServices(prev => {
            const next = { ...prev };
            filteredServices.forEach(s => {
                delete next[s.name];
            });
            return next;
        });
    };

    const toggleExpand = (svcName: string) => {
        setExpandedServices(prev => {
            const next = new Set(prev);
            if (next.has(svcName)) next.delete(svcName);
            else next.add(svcName);
            return next;
        });
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="w-[400px] sm:w-[600px] flex flex-col h-full">
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
                            <Badge variant="outline">{Object.keys(services).length} Selected</Badge>
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

                            <ScrollArea className="h-[400px] pr-3">
                                {isLoadingServices ? (
                                    <div className="flex justify-center py-8">
                                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        {filteredServices.length === 0 && (
                                            <div className="text-sm text-muted-foreground text-center py-4">No services found.</div>
                                        )}
                                        {filteredServices.map(svc => {
                                            const isSelected = !!services[svc.name]?.enabled;
                                            const config = services[svc.name];
                                            const serviceTools = getServiceTools(svc.name);
                                            const isExpanded = expandedServices.has(svc.name);
                                            const allowedCount = config?.allowedTools?.length ?? serviceTools.length;

                                            return (
                                                <div key={svc.name} className="border rounded-md bg-background/50 overflow-hidden">
                                                    <div className="flex items-center space-x-3 p-3 hover:bg-muted/50 transition-colors">
                                                        <Checkbox
                                                            id={`svc-${svc.name}`}
                                                            checked={isSelected}
                                                            onCheckedChange={(c) => toggleService(svc.name, !!c)}
                                                        />
                                                        <div className="flex-1 grid gap-1 leading-none cursor-pointer" onClick={() => toggleExpand(svc.name)}>
                                                            <div className="flex items-center justify-between">
                                                                <label
                                                                    htmlFor={`svc-${svc.name}`}
                                                                    className="text-sm font-medium leading-none cursor-pointer"
                                                                    onClick={(e) => e.stopPropagation()} // Prevent expand on label click if it toggles checkbox
                                                                >
                                                                    {svc.name}
                                                                </label>
                                                                {isSelected && (
                                                                    <Badge variant="secondary" className="text-[10px] h-5">
                                                                        {allowedCount === serviceTools.length ? "All Tools" : `${allowedCount} Tools`}
                                                                    </Badge>
                                                                )}
                                                            </div>
                                                            <p className="text-xs text-muted-foreground">
                                                                {svc.commandLineService ? "Command" : svc.httpService ? "HTTP" : "Remote"}
                                                            </p>
                                                        </div>
                                                        {isSelected && (
                                                            <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => toggleExpand(svc.name)}>
                                                                {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                                                            </Button>
                                                        )}
                                                    </div>

                                                    {isSelected && isExpanded && (
                                                        <div className="border-t bg-muted/20 p-3 space-y-2 animate-in slide-in-from-top-1 duration-200">
                                                            <div className="flex items-center justify-between text-xs text-muted-foreground mb-2">
                                                                <span>Available Tools</span>
                                                                <div className="flex gap-2">
                                                                    <button
                                                                        className="hover:text-primary"
                                                                        onClick={() => setServices(prev => ({
                                                                            ...prev,
                                                                            [svc.name]: { ...prev[svc.name], allowedTools: undefined }
                                                                        }))}
                                                                    >Allow All</button>
                                                                    <button
                                                                        className="hover:text-primary"
                                                                        onClick={() => setServices(prev => ({
                                                                            ...prev,
                                                                            [svc.name]: { ...prev[svc.name], allowedTools: [] }
                                                                        }))}
                                                                    >None</button>
                                                                </div>
                                                            </div>
                                                            {serviceTools.length === 0 ? (
                                                                <div className="text-xs text-muted-foreground italic pl-6">No tools detected.</div>
                                                            ) : (
                                                                <div className="grid grid-cols-1 gap-2 pl-2">
                                                                    {serviceTools.map(tool => {
                                                                        const isAllowed = !config?.allowedTools || config.allowedTools.includes(tool.name);
                                                                        return (
                                                                            <div key={tool.name} className="flex items-center space-x-2">
                                                                                <Checkbox
                                                                                    id={`tool-${tool.name}`}
                                                                                    checked={isAllowed}
                                                                                    onCheckedChange={(c) => toggleTool(svc.name, tool.name, !!c)}
                                                                                    className="h-3 w-3"
                                                                                />
                                                                                <label
                                                                                    htmlFor={`tool-${tool.name}`}
                                                                                    className="text-xs cursor-pointer truncate"
                                                                                    title={tool.name}
                                                                                >
                                                                                    {tool.name}
                                                                                </label>
                                                                            </div>
                                                                        );
                                                                    })}
                                                                </div>
                                                            )}
                                                        </div>
                                                    )}
                                                </div>
                                            );
                                        })}
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
