/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { UpstreamServiceConfig } from "@/lib/client";
import { Search } from "lucide-react";

export interface ProfileEditorProps {
    profile?: any; // The existing profile (if editing)
    services: UpstreamServiceConfig[];
    onSave: (profile: any) => void;
    onCancel: () => void;
}

export function ProfileEditor({ profile, services, onSave, onCancel }: ProfileEditorProps) {
    const [name, setName] = useState(profile?.name || "");
    const [tags, setTags] = useState<string>(profile?.selector?.tags?.join(", ") || "dev");
    const [enabledServices, setEnabledServices] = useState<Set<string>>(new Set());
    const [searchQuery, setSearchQuery] = useState("");

    // Initialize state from profile
    useEffect(() => {
        if (profile) {
            setName(profile.name);
            setTags(profile.selector?.tags?.join(", ") || "");

            const initialEnabled = new Set<string>();
            if (profile.serviceConfig) {
                Object.entries(profile.serviceConfig).forEach(([svcName, config]: [string, any]) => {
                    if (config.enabled) {
                        initialEnabled.add(svcName);
                    }
                });
            } else if (profile.services && Array.isArray(profile.services)) {
                // Fallback if we passed the UI 'Profile' type which has .services array
                profile.services.forEach((s: string) => initialEnabled.add(s));
            }
            setEnabledServices(initialEnabled);
        } else {
            // New profile defaults
            setName("");
            setTags("dev");
            setEnabledServices(new Set());
        }
    }, [profile]);

    const handleToggleService = (serviceName: string, checked: boolean) => {
        const next = new Set(enabledServices);
        if (checked) {
            next.add(serviceName);
        } else {
            next.delete(serviceName);
        }
        setEnabledServices(next);
    };

    const handleSave = () => {
        // Construct ProfileDefinition object
        const tagList = tags.split(",").map(t => t.trim()).filter(t => t.length > 0);

        const serviceConfig: Record<string, { enabled: boolean }> = {};
        enabledServices.forEach(svcName => {
            serviceConfig[svcName] = { enabled: true };
        });

        const profileData = {
            name,
            selector: {
                tags: tagList
            },
            serviceConfig
        };

        onSave(profileData);
    };

    const filteredServices = services.filter(s =>
        s.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col h-full gap-4 py-4">
            <div className="grid gap-2">
                <Label htmlFor="name">Profile Name</Label>
                <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g. production-client"
                    disabled={!!profile} // ID/Name often immutable after creation
                />
                {profile && <p className="text-xs text-muted-foreground">Profile name cannot be changed.</p>}
            </div>

            <div className="grid gap-2">
                <Label htmlFor="tags">Tags (Comma separated)</Label>
                <Input
                    id="tags"
                    value={tags}
                    onChange={(e) => setTags(e.target.value)}
                    placeholder="e.g. dev, mobile, experimental"
                />
            </div>

            <div className="flex-1 flex flex-col gap-2 min-h-0">
                <div className="flex items-center justify-between">
                    <Label>Enabled Services ({enabledServices.size})</Label>
                    <div className="flex items-center gap-2">
                        <Button variant="ghost" size="xs" onClick={() => setEnabledServices(new Set(services.map(s => s.name)))} className="h-6 text-xs">All</Button>
                        <Button variant="ghost" size="xs" onClick={() => setEnabledServices(new Set())} className="h-6 text-xs">None</Button>
                    </div>
                </div>

                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-3 w-3 text-muted-foreground" />
                    <Input
                        className="pl-8 h-8 text-xs"
                        placeholder="Search services..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>

                <ScrollArea className="flex-1 border rounded-md p-2">
                    <div className="space-y-2">
                        {filteredServices.map(service => (
                            <div key={service.name} className="flex items-start space-x-3 p-2 hover:bg-muted/50 rounded-md transition-colors">
                                <Checkbox
                                    id={`svc-${service.name}`}
                                    checked={enabledServices.has(service.name)}
                                    onCheckedChange={(c) => handleToggleService(service.name, c as boolean)}
                                />
                                <div className="grid gap-1.5 leading-none">
                                    <label
                                        htmlFor={`svc-${service.name}`}
                                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                                    >
                                        {service.name}
                                    </label>
                                    <p className="text-xs text-muted-foreground">
                                        {service.mcpService ? "MCP" : service.httpService ? "HTTP" : "Service"} • {service.version}
                                    </p>
                                </div>
                                {service.tags && service.tags.length > 0 && (
                                    <div className="ml-auto flex gap-1">
                                        {service.tags.slice(0, 2).map(tag => (
                                            <Badge key={tag} variant="outline" className="text-[10px] h-4 px-1">{tag}</Badge>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ))}
                        {filteredServices.length === 0 && (
                            <p className="text-center text-sm text-muted-foreground py-4">No services found.</p>
                        )}
                    </div>
                </ScrollArea>
            </div>

            <div className="flex justify-end gap-2 pt-2 border-t mt-auto">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleSave} disabled={!name}>Save Profile</Button>
            </div>
        </div>
    );
}
