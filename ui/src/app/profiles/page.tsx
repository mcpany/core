
"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Edit } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { useToast } from "@/hooks/use-toast";

interface Profile {
    id: string;
    name: string;
    description: string;
    env: Record<string, string>;
    active: boolean;
}

export default function ProfilesPage() {
    const [profiles, setProfiles] = useState<Profile[]>([]);
    const [selectedProfile, setSelectedProfile] = useState<Profile | null>(null);
    const [isSheetOpen, setIsSheetOpen] = useState(false);
    const { toast } = useToast();

    useEffect(() => {
        fetchProfiles();
    }, []);

    const fetchProfiles = async () => {
        try {
            const res = await fetch('/api/profiles');
            if (res.ok) {
                const data = await res.json();
                setProfiles(data);
            }
        } catch (e) {
            console.error("Failed to fetch profiles", e);
        }
    };

    const openNew = () => {
        setSelectedProfile({ id: "", name: "", description: "", env: {}, active: false });
        setIsSheetOpen(true);
    };

    const openEdit = (profile: Profile) => {
        setSelectedProfile({ ...profile }); // copy
        setIsSheetOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Delete this profile?")) return;
        try {
            await fetch('/api/profiles', {
                method: 'POST',
                body: JSON.stringify({ action: 'delete', id })
            });
            setProfiles(profiles.filter(p => p.id !== id));
            toast({ title: "Profile Deleted" });
        } catch (e) {
            console.error(e);
            toast({ variant: "destructive", title: "Error deleting profile" });
        }
    };

    const handleSave = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!selectedProfile) return;

        try {
             const res = await fetch('/api/profiles', {
                method: 'POST',
                body: JSON.stringify(selectedProfile)
            });
            if (res.ok) {
                toast({ title: "Profile Saved" });
                setIsSheetOpen(false);
                fetchProfiles();
            }
        } catch (e) {
             toast({ variant: "destructive", title: "Error saving profile" });
        }
    };

    const updateEnv = (key: string, value: string) => {
        if (!selectedProfile) return;
        const newEnv = { ...selectedProfile.env };
        if (value === "") delete newEnv[key]; // crude delete
        else newEnv[key] = value;
        setSelectedProfile({ ...selectedProfile, env: newEnv });
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Execution Profiles</h2>
                <Button onClick={openNew}>
                    <Plus className="mr-2 h-4 w-4" /> Create Profile
                </Button>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {profiles.map(profile => (
                    <Card key={profile.id} className="relative group">
                        <CardHeader>
                            <CardTitle className="flex justify-between items-center">
                                {profile.name}
                                {profile.active && <Badge variant="default">Active</Badge>}
                            </CardTitle>
                            <CardDescription>{profile.description}</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-2">
                                <p className="text-sm font-medium">Environment Variables:</p>
                                <div className="bg-muted p-2 rounded text-xs font-mono max-h-32 overflow-y-auto">
                                    {Object.entries(profile.env).map(([k, v]) => (
                                        <div key={k}>{k}={v}</div>
                                    ))}
                                    {Object.keys(profile.env).length === 0 && <span className="text-muted-foreground italic">None</span>}
                                </div>
                            </div>
                            <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity flex space-x-1">
                                <Button variant="secondary" size="icon" onClick={() => openEdit(profile)}>
                                    <Edit className="h-4 w-4" />
                                </Button>
                                <Button variant="destructive" size="icon" onClick={() => handleDelete(profile.id)}>
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
                <SheetContent className="w-[400px] sm:w-[540px]">
                    <SheetHeader>
                        <SheetTitle>{selectedProfile?.id ? "Edit Profile" : "New Profile"}</SheetTitle>
                        <SheetDescription>Configure execution environment settings.</SheetDescription>
                    </SheetHeader>
                    {selectedProfile && (
                        <form onSubmit={handleSave} className="space-y-4 py-4">
                            <div className="space-y-2">
                                <Label htmlFor="name">Name</Label>
                                <Input id="name" value={selectedProfile.name} onChange={e => setSelectedProfile({...selectedProfile, name: e.target.value})} required />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="desc">Description</Label>
                                <Input id="desc" value={selectedProfile.description} onChange={e => setSelectedProfile({...selectedProfile, description: e.target.value})} />
                            </div>
                             <div className="space-y-2">
                                <Label>Environment Variables (JSON)</Label>
                                <Textarea
                                    className="font-mono text-sm"
                                    rows={10}
                                    value={JSON.stringify(selectedProfile.env, null, 2)}
                                    onChange={e => {
                                        try {
                                            const env = JSON.parse(e.target.value);
                                            setSelectedProfile({...selectedProfile, env});
                                        } catch(err) {
                                            // ignore parse errors while typing
                                        }
                                    }}
                                />
                                <p className="text-xs text-muted-foreground">Edit as JSON object.</p>
                            </div>
                            <div className="flex justify-end pt-4">
                                <Button type="submit">Save Profile</Button>
                            </div>
                        </form>
                    )}
                </SheetContent>
            </Sheet>
        </div>
    );
}
