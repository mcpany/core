/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Loader2, Save } from "lucide-react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { ProfileDefinition } from "@/types/profile";
import { toast } from "sonner";

interface ProfileEditorProps {
    profile?: ProfileDefinition | null;
    onSave: (profile: ProfileDefinition) => Promise<void>;
    onCancel: () => void;
}

export function ProfileEditor({ profile, onSave, onCancel }: ProfileEditorProps) {
    const [name, setName] = useState(profile?.name || "");
    const [tags, setTags] = useState(profile?.selector?.tags?.join(", ") || "dev");
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [serviceOverrides, setServiceOverrides] = useState<Record<string, "auto" | "enabled" | "disabled">>({});

    useEffect(() => {
        const loadServices = async () => {
            setLoading(true);
            try {
                const data = await apiClient.listServices();
                const list = Array.isArray(data) ? data : (data.services || []);
                setServices(list);

                // Initialize overrides from profile
                const overrides: Record<string, "auto" | "enabled" | "disabled"> = {};
                if (profile?.serviceConfig) {
                    Object.entries(profile.serviceConfig).forEach(([svcName, config]) => {
                        if (config.enabled === true) overrides[svcName] = "enabled";
                        else if (config.enabled === false) overrides[svcName] = "disabled";
                    });
                }
                setServiceOverrides(overrides);

            } catch (error) {
                console.error("Failed to load services", error);
                toast.error("Failed to load services");
            } finally {
                setLoading(false);
            }
        };

        loadServices();
    }, [profile]);

    const handleSave = async () => {
        if (!name.trim()) {
            toast.error("Profile name is required");
            return;
        }

        const tagList = tags.split(",").map(t => t.trim()).filter(Boolean);

        const newServiceConfig: Record<string, { enabled?: boolean }> = {};
        Object.entries(serviceOverrides).forEach(([svcName, state]) => {
            if (state === "enabled") newServiceConfig[svcName] = { enabled: true };
            if (state === "disabled") newServiceConfig[svcName] = { enabled: false };
        });

        const newProfile: ProfileDefinition = {
            name: name,
            selector: {
                tags: tagList
            },
            serviceConfig: newServiceConfig,
            // Preserve other fields if editing
            requiredRoles: profile?.requiredRoles,
            parentProfileIds: profile?.parentProfileIds,
            secrets: profile?.secrets
        };

        await onSave(newProfile);
    };

    const getOverrideState = (svcName: string) => serviceOverrides[svcName] || "auto";

    const setOverrideState = (svcName: string, state: "auto" | "enabled" | "disabled") => {
        setServiceOverrides(prev => {
            const next = { ...prev };
            if (state === "auto") delete next[svcName];
            else next[svcName] = state;
            return next;
        });
    };

    return (
        <div className="flex flex-col h-full gap-6 p-1">
            <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                    <Label htmlFor="name">Profile Name</Label>
                    <Input
                        id="name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        placeholder="e.g. production-api"
                        disabled={!!profile} // Usually ID/Name is immutable after creation, but strictly speaking we could allow renaming if backend supports it (it doesn't usually).
                        // Wait, updateProfile uses name in URL. So changing name = creating new?
                        // For now disable if editing.
                    />
                </div>
                <div className="grid gap-2">
                    <Label htmlFor="tags">Selector Tags</Label>
                    <Input
                        id="tags"
                        value={tags}
                        onChange={(e) => setTags(e.target.value)}
                        placeholder="e.g. dev, public"
                    />
                    <p className="text-xs text-muted-foreground">
                        Services with these tags will be automatically enabled (unless disabled below).
                    </p>
                </div>
            </div>

            <div className="flex-1 flex flex-col min-h-0 border rounded-md">
                <div className="p-3 border-b bg-muted/20 font-medium text-sm flex items-center justify-between">
                    <span>Service Configuration</span>
                    <span className="text-xs text-muted-foreground font-normal">
                        Override default selection logic
                    </span>
                </div>
                <ScrollArea className="flex-1">
                    {loading ? (
                        <div className="flex items-center justify-center h-40">
                            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Service</TableHead>
                                    <TableHead>Tags</TableHead>
                                    <TableHead className="w-[180px]">Status</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {services.map((svc) => (
                                    <TableRow key={svc.name}>
                                        <TableCell className="font-medium">{svc.name}</TableCell>
                                        <TableCell>
                                            <div className="flex flex-wrap gap-1">
                                                {svc.tags?.map(t => (
                                                    <Badge key={t} variant="secondary" className="text-[10px] h-5 px-1">{t}</Badge>
                                                ))}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Select
                                                value={getOverrideState(svc.name)}
                                                onValueChange={(val: any) => setOverrideState(svc.name, val)}
                                            >
                                                <SelectTrigger className="h-8 text-xs">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="auto">Default (Auto)</SelectItem>
                                                    <SelectItem value="enabled">Force Enable</SelectItem>
                                                    <SelectItem value="disabled">Force Disable</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </ScrollArea>
            </div>

            <div className="flex justify-end gap-3 pt-2">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleSave}>
                    <Save className="mr-2 h-4 w-4" />
                    Save Profile
                </Button>
            </div>
        </div>
    );
}
