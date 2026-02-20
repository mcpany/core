/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { UpstreamServiceConfig } from "@/lib/client";
import {
    FilesystemUpstreamService,
    FilesystemUpstreamService_SymlinkMode
} from "@proto/config/v1/upstream_service";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue
} from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Plus, Trash2, Folder, HardDrive, Key, Globe, Archive } from "lucide-react";
import { SecretPicker } from "@/components/secrets/secret-picker";

interface FilesystemConfigProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

interface RootPathEntry {
    id: string;
    virtual: string;
    physical: string;
}

export function FilesystemConfig({ service, onChange }: FilesystemConfigProps) {
    // Ensure filesystemService exists
    const fsConfig = service.filesystemService || {
        rootPaths: {},
        readOnly: true,
        allowedPaths: [],
        deniedPaths: [],
        symlinkMode: FilesystemUpstreamService_SymlinkMode.SYMLINK_MODE_UNSPECIFIED,
        os: {}
    };

    // Use local state for array-based editing of the map to prevent reordering issues
    const [rootPathEntries, setRootPathEntries] = useState<RootPathEntry[]>(() => {
        return Object.entries(fsConfig.rootPaths || {}).map(([k, v]) => ({
            id: crypto.randomUUID(),
            virtual: k,
            physical: v
        }));
    });

    const updateFsConfig = (updates: Partial<FilesystemUpstreamService>) => {
        onChange({
            ...service,
            filesystemService: { ...fsConfig, ...updates }
        });
    };

    // Helper to sync local entries to parent config
    const updateRootPaths = (entries: RootPathEntry[]) => {
        const newRoots: Record<string, string> = {};
        entries.forEach(e => {
            // Only add if virtual path is not empty to avoid empty keys
            if (e.virtual.trim()) {
                newRoots[e.virtual] = e.physical;
            }
        });
        updateFsConfig({ rootPaths: newRoots });
    };

    const handleEntryChange = (id: string, field: 'virtual' | 'physical', value: string) => {
        const newEntries = rootPathEntries.map(e => e.id === id ? { ...e, [field]: value } : e);
        setRootPathEntries(newEntries);
        updateRootPaths(newEntries);
    };

    const addEntry = () => {
        const newEntries = [...rootPathEntries, { id: crypto.randomUUID(), virtual: "", physical: "" }];
        setRootPathEntries(newEntries);
        // We don't sync to parent immediately if key is empty, to avoid pollution
    };

    const removeEntry = (id: string) => {
        const newEntries = rootPathEntries.filter(e => e.id !== id);
        setRootPathEntries(newEntries);
        updateRootPaths(newEntries);
    };

    const getFsType = () => {
        if (fsConfig.s3) return 's3';
        if (fsConfig.gcs) return 'gcs';
        if (fsConfig.sftp) return 'sftp';
        if (fsConfig.zip) return 'zip';
        if (fsConfig.http) return 'http';
        return 'os';
    };

    const handleTypeChange = (type: string) => {
        const newConfig = { ...fsConfig };
        // Reset specialized fields
        delete newConfig.os;
        delete newConfig.s3;
        delete newConfig.gcs;
        delete newConfig.sftp;
        delete newConfig.zip;
        delete newConfig.http;
        delete newConfig.tmpfs;

        switch (type) {
            case 'os':
                newConfig.os = {};
                break;
            case 's3':
                newConfig.s3 = { bucket: "", region: "", accessKeyId: "", secretAccessKey: "", sessionToken: "", endpoint: "" };
                break;
            case 'gcs':
                newConfig.gcs = { bucket: "" };
                break;
            case 'sftp':
                newConfig.sftp = { address: "", username: "", password: "", keyPath: "" };
                break;
            case 'zip':
                newConfig.zip = { filePath: "" };
                break;
            default:
                newConfig.os = {};
        }
        updateFsConfig(newConfig);
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <HardDrive className="h-5 w-5" />
                        Backend Storage
                    </CardTitle>
                    <CardDescription>
                        Select the type of filesystem or object store to expose.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label>Storage Type</Label>
                        <Select value={getFsType()} onValueChange={handleTypeChange}>
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="os">Local Filesystem (OS)</SelectItem>
                                <SelectItem value="s3">Amazon S3 / Compatible</SelectItem>
                                <SelectItem value="gcs">Google Cloud Storage</SelectItem>
                                <SelectItem value="sftp">SFTP Server</SelectItem>
                                <SelectItem value="zip">Zip Archive</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    <Separator />

                    {/* S3 Config */}
                    {fsConfig.s3 && (
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>Bucket Name</Label>
                                    <Input
                                        value={fsConfig.s3.bucket}
                                        onChange={(e) => updateFsConfig({ s3: { ...fsConfig.s3!, bucket: e.target.value } })}
                                        placeholder="my-bucket"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Region</Label>
                                    <Input
                                        value={fsConfig.s3.region}
                                        onChange={(e) => updateFsConfig({ s3: { ...fsConfig.s3!, region: e.target.value } })}
                                        placeholder="us-east-1"
                                    />
                                </div>
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>Access Key ID</Label>
                                    <Input
                                        value={fsConfig.s3.accessKeyId}
                                        onChange={(e) => updateFsConfig({ s3: { ...fsConfig.s3!, accessKeyId: e.target.value } })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Secret Access Key</Label>
                                    <div className="flex gap-2">
                                        <Input
                                            type="password"
                                            value={fsConfig.s3.secretAccessKey}
                                            onChange={(e) => updateFsConfig({ s3: { ...fsConfig.s3!, secretAccessKey: e.target.value } })}
                                        />
                                        <SecretPicker
                                            onSelect={(key) => updateFsConfig({ s3: { ...fsConfig.s3!, secretAccessKey: `\${${key}}` } })}
                                        >
                                            <Button variant="outline" size="icon">
                                                <Key className="h-4 w-4" />
                                            </Button>
                                        </SecretPicker>
                                    </div>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label>Endpoint (Optional)</Label>
                                <Input
                                    value={fsConfig.s3.endpoint}
                                    onChange={(e) => updateFsConfig({ s3: { ...fsConfig.s3!, endpoint: e.target.value } })}
                                    placeholder="https://minio.example.com"
                                />
                                <p className="text-xs text-muted-foreground">Leave empty for AWS S3.</p>
                            </div>
                        </div>
                    )}

                    {/* GCS Config */}
                    {fsConfig.gcs && (
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <Label>Bucket Name</Label>
                                <Input
                                    value={fsConfig.gcs.bucket}
                                    onChange={(e) => updateFsConfig({ gcs: { ...fsConfig.gcs!, bucket: e.target.value } })}
                                    placeholder="my-gcs-bucket"
                                />
                            </div>
                        </div>
                    )}

                    {/* SFTP Config */}
                    {fsConfig.sftp && (
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <Label>Address</Label>
                                <Input
                                    value={fsConfig.sftp.address}
                                    onChange={(e) => updateFsConfig({ sftp: { ...fsConfig.sftp!, address: e.target.value } })}
                                    placeholder="sftp.example.com:22"
                                />
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>Username</Label>
                                    <Input
                                        value={fsConfig.sftp.username}
                                        onChange={(e) => updateFsConfig({ sftp: { ...fsConfig.sftp!, username: e.target.value } })}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Password</Label>
                                    <div className="flex gap-2">
                                        <Input
                                            type="password"
                                            value={fsConfig.sftp.password}
                                            onChange={(e) => updateFsConfig({ sftp: { ...fsConfig.sftp!, password: e.target.value } })}
                                        />
                                        <SecretPicker
                                            onSelect={(key) => updateFsConfig({ sftp: { ...fsConfig.sftp!, password: `\${${key}}` } })}
                                        >
                                            <Button variant="outline" size="icon">
                                                <Key className="h-4 w-4" />
                                            </Button>
                                        </SecretPicker>
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Zip Config */}
                    {fsConfig.zip && (
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <Label>File Path</Label>
                                <div className="flex gap-2">
                                    <Input
                                        value={fsConfig.zip.filePath}
                                        onChange={(e) => updateFsConfig({ zip: { ...fsConfig.zip!, filePath: e.target.value } })}
                                        placeholder="/path/to/archive.zip"
                                    />
                                    {/* Removed dead browse button */}
                                </div>
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Folder className="h-5 w-5" />
                        Mount Points
                    </CardTitle>
                    <CardDescription>
                        Map virtual paths (seen by the Agent) to physical paths on the backend.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-12 gap-2 font-medium text-sm text-muted-foreground mb-2">
                        <div className="col-span-5">Virtual Path</div>
                        <div className="col-span-1 text-center">→</div>
                        <div className="col-span-5">Target Path / Prefix</div>
                        <div className="col-span-1"></div>
                    </div>

                    {rootPathEntries.map((entry) => (
                        <div key={entry.id} className="grid grid-cols-12 gap-2 items-center">
                            <div className="col-span-5">
                                <Input
                                    value={entry.virtual}
                                    onChange={(e) => handleEntryChange(entry.id, 'virtual', e.target.value)}
                                    placeholder="/workspace"
                                />
                            </div>
                            <div className="col-span-1 text-center text-muted-foreground">→</div>
                            <div className="col-span-5">
                                <Input
                                    value={entry.physical}
                                    onChange={(e) => handleEntryChange(entry.id, 'physical', e.target.value)}
                                    placeholder={getFsType() === 'os' ? "/home/user/projects" : "folder/prefix"}
                                />
                            </div>
                            <div className="col-span-1 text-right">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="text-destructive hover:bg-destructive/10"
                                    onClick={() => removeEntry(entry.id)}
                                >
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    ))}

                    <Button
                        variant="outline"
                        size="sm"
                        className="w-full mt-2 border-dashed"
                        onClick={addEntry}
                    >
                        <Plus className="mr-2 h-4 w-4" /> Add Mount Point
                    </Button>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Globe className="h-5 w-5" />
                        Security & Access
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div className="space-y-0.5">
                            <Label>Read Only Mode</Label>
                            <p className="text-xs text-muted-foreground">
                                Prevent any write operations (upload, edit, delete).
                            </p>
                        </div>
                        <Switch
                            checked={fsConfig.readOnly}
                            onCheckedChange={(checked) => updateFsConfig({ readOnly: checked })}
                        />
                    </div>

                    <Separator />

                    <div className="space-y-2">
                        <Label>Symlink Policy</Label>
                        <Select
                            value={(fsConfig.symlinkMode ?? 0).toString()}
                            onValueChange={(val) => updateFsConfig({ symlinkMode: parseInt(val) })}
                        >
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value={FilesystemUpstreamService_SymlinkMode.SYMLINK_MODE_UNSPECIFIED.toString()}>Default (Unspecified)</SelectItem>
                                <SelectItem value={FilesystemUpstreamService_SymlinkMode.ALLOW.toString()}>Allow All</SelectItem>
                                <SelectItem value={FilesystemUpstreamService_SymlinkMode.INTERNAL_ONLY.toString()}>Internal Only (Safe)</SelectItem>
                                <SelectItem value={FilesystemUpstreamService_SymlinkMode.DENY.toString()}>Deny All</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
