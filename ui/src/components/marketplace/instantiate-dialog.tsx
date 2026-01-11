/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { UpstreamServiceConfig, apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";

interface InstantiateDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    templateConfig?: UpstreamServiceConfig;
    onComplete: () => void;
}

export function InstantiateDialog({ open, onOpenChange, templateConfig, onComplete }: InstantiateDialogProps) {
    const { toast } = useToast();
    const router = useRouter();
    const [name, setName] = useState("");
    const [authId, setAuthId] = useState("none");
    const [loading, setLoading] = useState(false);
    const [credentials, setCredentials] = useState<any[]>([]);

    useEffect(() => {
        if (open && templateConfig) {
            setName(`${templateConfig.name}-copy`);
            setAuthId("none");
            apiClient.listCredentials().then(setCredentials).catch(console.error);
        }
    }, [open, templateConfig]);

    const handleInstantiate = async () => {
        if (!templateConfig) return;
        setLoading(true);
        try {
            const newConfig = { ...templateConfig };
            newConfig.name = name;
            newConfig.id = name; // ID is name for now

            if (authId !== 'none') {
                const cred = credentials.find(c => c.id === authId);
                if (cred && cred.authentication) {
                    newConfig.upstreamAuth = cred.authentication;
                }
            }

            await apiClient.registerService(newConfig);
            toast({ title: "Service Instantiated", description: `${name} is now running.` });
            onOpenChange(false);

            // Redirect to the new service page
            router.push(`/upstream-services/${name}`);
        } catch (e) {
            toast({ title: "Failed to instantiate", variant: "destructive", description: String(e) });
        } finally {
            setLoading(false);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Instantiate Service</DialogTitle>
                    <DialogDescription>
                        Create a running instance from this template.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid gap-2">
                        <Label htmlFor="service-name-input">Service Name</Label>
                        <Input id="service-name-input" value={name} onChange={e => setName(e.target.value)} />
                    </div>
                    <div className="grid gap-2">
                        <Label>Authentication</Label>
                        <Select value={authId} onValueChange={setAuthId}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select credential" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="none">None</SelectItem>
                                {credentials.map(c => (
                                    <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleInstantiate} disabled={loading}>
                        {loading ? "Creating..." : "Create Instance"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
