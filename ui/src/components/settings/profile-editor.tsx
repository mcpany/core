/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { ProfileDefinition } from "@/lib/client";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

interface ProfileEditorProps {
    profile?: ProfileDefinition | null;
    onSubmit: (profile: ProfileDefinition) => void;
    onCancel: () => void;
}

export function ProfileEditor({ profile, onSubmit, onCancel }: ProfileEditorProps) {
    const [name, setName] = useState("");
    const [tags, setTags] = useState("");
    const [roles, setRoles] = useState("");
    const [toolProperties, setToolProperties] = useState("");

    useEffect(() => {
        if (profile) {
            setName(profile.name);
            setTags(profile.selector?.tags?.join(", ") || "");
            setRoles(profile.requiredRoles?.join(", ") || "");
            setToolProperties(JSON.stringify(profile.selector?.toolProperties || {}, null, 2));
        } else {
            setName("");
            setTags("");
            setRoles("");
            setToolProperties("{}");
        }
    }, [profile]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        let parsedToolProps = {};
        try {
            parsedToolProps = JSON.parse(toolProperties);
        } catch (e) {
            alert("Invalid JSON in Tool Properties");
            return;
        }

        const newProfile: ProfileDefinition = {
            name,
            selector: {
                tags: tags.split(",").map(t => t.trim()).filter(Boolean),
                toolProperties: parsedToolProps,
            },
            requiredRoles: roles.split(",").map(r => r.trim()).filter(Boolean),
            serviceConfig: profile?.serviceConfig || {},
            secrets: profile?.secrets || {},
            parentProfileIds: profile?.parentProfileIds || [],
        };

        onSubmit(newProfile);
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
                <Label htmlFor="name">Profile Name</Label>
                <Input
                    id="name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g., development"
                    disabled={!!profile} // Name usually immutable for ID?
                    required
                />
                {profile && <p className="text-[10px] text-muted-foreground">Profile name cannot be changed once created.</p>}
            </div>

            <div className="space-y-2">
                <Label htmlFor="tags">Selector Tags</Label>
                <Input
                    id="tags"
                    value={tags}
                    onChange={(e) => setTags(e.target.value)}
                    placeholder="e.g., dev, web, internal (comma separated)"
                />
            </div>

            <div className="space-y-2">
                <Label htmlFor="roles">Required Roles</Label>
                <Input
                    id="roles"
                    value={roles}
                    onChange={(e) => setRoles(e.target.value)}
                    placeholder="e.g., admin, developer (comma separated)"
                />
            </div>

            <div className="space-y-2">
                <Label htmlFor="tool-props">Tool Properties (JSON)</Label>
                <Textarea
                    id="tool-props"
                    value={toolProperties}
                    onChange={(e) => setToolProperties(e.target.value)}
                    className="font-mono text-xs h-32"
                    placeholder="{}"
                />
            </div>

            <div className="flex justify-end gap-2 pt-4">
                <Button type="button" variant="outline" onClick={onCancel}>
                    Cancel
                </Button>
                <Button type="submit">
                    {profile ? "Save Changes" : "Create Profile"}
                </Button>
            </div>
        </form>
    );
}
