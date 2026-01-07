
"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { GlassCard } from "@/components/layout/glass-card";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatusBadge } from "@/components/layout/status-badge";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export default function ProfilesPage() {
    const [profiles, setProfiles] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        apiClient.listProfiles().then(res => {
            setProfiles(res);
            setLoading(false);
        }).catch(err => {
            console.error("Failed to load profiles", err);
            setLoading(false);
        });
    }, []);

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Profiles</h2>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Create Profile
                </Button>
            </div>
            <GlassCard>
                <CardHeader>
                    <CardTitle>Execution Profiles</CardTitle>
                    <CardDescription>Manage environment configurations (dev, prod, debug).</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>Name</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Roles</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center">Loading...</TableCell>
                                </TableRow>
                            ) : profiles.map((profile) => (
                                <TableRow key={profile.id}>
                                    <TableCell className="font-mono">{profile.id}</TableCell>
                                    <TableCell className="font-medium">{profile.name}</TableCell>
                                    <TableCell>
                                        <StatusBadge status={profile.active ? "active" : "inactive"} />
                                    </TableCell>
                                    <TableCell>{profile.roles?.join(", ")}</TableCell>
                                    <TableCell>
                                        <Button variant="ghost" size="sm">Edit</Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </GlassCard>
        </div>
    );
}
