/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig, FilesystemUpstreamService } from "@/lib/client";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, Key } from "lucide-react";
import { SecretPicker } from "@/components/secrets/secret-picker";
import { Card, CardContent } from "@/components/ui/card";

interface FilesystemConfigProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

/**
 * Component for configuring filesystem upstream services.
 * Allows users to set root paths, read-only mode, and filesystem type specific settings.
 *
 * @param props - Component properties.
 * @param props.service - The service configuration to edit.
 * @param props.onChange - Callback when configuration changes.
 */
export function FilesystemConfig({ service, onChange }: FilesystemConfigProps) {
    const fsConfig = service.filesystemService || {
        rootPaths: {},
        readOnly: true,
        filesystemType: { os: {} },
        allowedPaths: [],
        deniedPaths: [],
        symlinkMode: 0 // UNSPECIFIED
    };

    const updateFs = (updates: Partial<FilesystemUpstreamService>) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const newService = { ...service, filesystemService: { ...fsConfig, ...updates } as any };
        onChange(newService);
    };

    const getFsType = () => {
        if (fsConfig.filesystemType?.os) return "os";
        if (fsConfig.filesystemType?.s3) return "s3";
        if (fsConfig.filesystemType?.gcs) return "gcs";
        if (fsConfig.filesystemType?.zip) return "zip";
        if (fsConfig.filesystemType?.sftp) return "sftp";
        if (fsConfig.filesystemType?.tmpfs) return "tmpfs";
        return "os";
    };

    const handleTypeChange = (type: string) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        let newType: any = {};
        if (type === "os") newType = { os: {} };
        if (type === "s3") newType = { s3: { region: "us-east-1" } };
        if (type === "gcs") newType = { gcs: {} };
        if (type === "zip") newType = { zip: {} };
        if (type === "sftp") newType = { sftp: {} };
        if (type === "tmpfs") newType = { tmpfs: {} };

        updateFs({ filesystemType: newType });
    };

    // Root Paths Management
    const addRootPath = () => {
        const newPaths = { ...fsConfig.rootPaths, ["/new-mount"]: "/local/path" };
        updateFs({ rootPaths: newPaths });
    };

    const removeRootPath = (key: string) => {
        const newPaths = { ...fsConfig.rootPaths };
        delete newPaths[key];
        updateFs({ rootPaths: newPaths });
    };

    const updateRootPathKey = (oldKey: string, newKey: string) => {
        const newPaths = { ...fsConfig.rootPaths };
        const val = newPaths[oldKey];
        delete newPaths[oldKey];
        newPaths[newKey] = val;
        updateFs({ rootPaths: newPaths });
    };

    const updateRootPathValue = (key: string, value: string) => {
        const newPaths = { ...fsConfig.rootPaths };
        newPaths[key] = value;
        updateFs({ rootPaths: newPaths });
    };

    return (
        <div className="space-y-6">
            <div className="space-y-4">
                <div className="space-y-2">
                    <Label htmlFor="fs-type">Filesystem Type</Label>
                    <Select value={getFsType()} onValueChange={handleTypeChange}>
                        <SelectTrigger id="fs-type">
                            <SelectValue placeholder="Select type" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="os">Local OS</SelectItem>
                            <SelectItem value="s3">AWS S3</SelectItem>
                            <SelectItem value="gcs">Google Cloud Storage</SelectItem>
                            <SelectItem value="zip">ZIP Archive</SelectItem>
                            <SelectItem value="sftp">SFTP</SelectItem>
                            <SelectItem value="tmpfs">In-Memory (TmpFS)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div className="flex items-center space-x-2">
                    <Switch
                        id="read-only"
                        checked={fsConfig.readOnly}
                        onCheckedChange={(checked) => updateFs({ readOnly: checked })}
                    />
                    <Label htmlFor="read-only">Read Only</Label>
                </div>

                <div className="space-y-2">
                    <Label>Symlink Policy</Label>
                    <Select
                        value={fsConfig.symlinkMode?.toString() || "0"}
                        onValueChange={(val) => updateFs({ symlinkMode: parseInt(val) })}
                    >
                        <SelectTrigger>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="0">Default</SelectItem>
                            <SelectItem value="1">Allow All</SelectItem>
                            <SelectItem value="2">Deny All</SelectItem>
                            <SelectItem value="3">Internal Only (Safe)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <div className="space-y-3">
                <div className="flex items-center justify-between">
                    <Label className="text-base">Root Paths (Mounts)</Label>
                    <Button size="sm" variant="outline" onClick={addRootPath}>
                        <Plus className="h-4 w-4 mr-1" /> Add Path
                    </Button>
                </div>
                <div className="text-sm text-muted-foreground">
                    Map virtual paths (seen by LLM) to actual paths (local/bucket).
                </div>
                {Object.entries(fsConfig.rootPaths || {}).map(([key, val]) => (
                    <div key={key} className="flex items-center gap-2">
                        <Input
                            placeholder="/virtual/path"
                            value={key}
                            onChange={(e) => updateRootPathKey(key, e.target.value)}
                            className="flex-1"
                        />
                        <span className="text-muted-foreground">→</span>
                        <Input
                            placeholder="/real/path"
                            value={val}
                            onChange={(e) => updateRootPathValue(key, e.target.value)}
                            className="flex-1"
                        />
                        <Button size="icon" variant="ghost" onClick={() => removeRootPath(key)}>
                            <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                    </div>
                ))}
                {Object.keys(fsConfig.rootPaths || {}).length === 0 && (
                    <div className="p-4 border border-dashed rounded-md text-center text-muted-foreground text-sm">
                        No paths configured. The filesystem will be empty.
                    </div>
                )}
            </div>

            {/* Type Specific Configs */}
            {getFsType() === 's3' && (
                <Card className="bg-muted/30">
                    <CardContent className="pt-6 space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label>Bucket</Label>
                                <Input
                                    value={fsConfig.filesystemType?.s3?.bucket || ""}
                                    onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, bucket: e.target.value } } })}
                                    placeholder="my-bucket"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label>Region</Label>
                                <Input
                                    value={fsConfig.filesystemType?.s3?.region || ""}
                                    onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, region: e.target.value } } })}
                                    placeholder="us-east-1"
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Label>Access Key ID</Label>
                            <div className="flex gap-2">
                                <Input
                                    type="password"
                                    value={fsConfig.filesystemType?.s3?.accessKeyId || ""}
                                    onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, accessKeyId: e.target.value } } })}
                                />
                                <SecretPicker onSelect={(k) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, accessKeyId: `\${${k}}` } } })}>
                                    <Button variant="outline" size="icon"><Key className="h-4 w-4" /></Button>
                                </SecretPicker>
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Label>Secret Access Key</Label>
                            <div className="flex gap-2">
                                <Input
                                    type="password"
                                    value={fsConfig.filesystemType?.s3?.secretAccessKey || ""}
                                    onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, secretAccessKey: e.target.value } } })}
                                />
                                <SecretPicker onSelect={(k) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, secretAccessKey: `\${${k}}` } } })}>
                                    <Button variant="outline" size="icon"><Key className="h-4 w-4" /></Button>
                                </SecretPicker>
                            </div>
                        </div>
                         <div className="space-y-2">
                            <Label>Endpoint (Optional)</Label>
                            <Input
                                value={fsConfig.filesystemType?.s3?.endpoint || ""}
                                onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, s3: { ...fsConfig.filesystemType?.s3, endpoint: e.target.value } } })}
                                placeholder="https://s3.custom-endpoint.com"
                            />
                        </div>
                    </CardContent>
                </Card>
            )}

            {getFsType() === 'gcs' && (
                <div className="space-y-2">
                    <Label>Bucket</Label>
                    <Input
                        value={fsConfig.filesystemType?.gcs?.bucket || ""}
                        onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, gcs: { ...fsConfig.filesystemType?.gcs, bucket: e.target.value } } })}
                        placeholder="my-gcs-bucket"
                    />
                </div>
            )}

            {getFsType() === 'zip' && (
                <div className="space-y-2">
                    <Label>ZIP File Path</Label>
                    <Input
                        value={fsConfig.filesystemType?.zip?.filePath || ""}
                        onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, zip: { ...fsConfig.filesystemType?.zip, filePath: e.target.value } } })}
                        placeholder="/path/to/archive.zip"
                    />
                </div>
            )}

            {getFsType() === 'sftp' && (
                <Card className="bg-muted/30">
                    <CardContent className="pt-6 space-y-4">
                        <div className="space-y-2">
                            <Label>Address</Label>
                            <Input
                                value={fsConfig.filesystemType?.sftp?.address || ""}
                                onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, sftp: { ...fsConfig.filesystemType?.sftp, address: e.target.value } } })}
                                placeholder="sftp.example.com:22"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Username</Label>
                            <Input
                                value={fsConfig.filesystemType?.sftp?.username || ""}
                                onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, sftp: { ...fsConfig.filesystemType?.sftp, username: e.target.value } } })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Password</Label>
                            <div className="flex gap-2">
                                <Input
                                    type="password"
                                    value={fsConfig.filesystemType?.sftp?.password || ""}
                                    onChange={(e) => updateFs({ filesystemType: { ...fsConfig.filesystemType, sftp: { ...fsConfig.filesystemType?.sftp, password: e.target.value } } })}
                                />
                                <SecretPicker onSelect={(k) => updateFs({ filesystemType: { ...fsConfig.filesystemType, sftp: { ...fsConfig.filesystemType?.sftp, password: `\${${k}}` } } })}>
                                    <Button variant="outline" size="icon"><Key className="h-4 w-4" /></Button>
                                </SecretPicker>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}
