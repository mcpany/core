/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { UpstreamServiceConfig, FilesystemUpstreamService, FilesystemUpstreamService_SymlinkMode, S3Fs, GcsFs, SftpFs } from "@proto/config/v1/upstream_service";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2, Folder, Shield, Cloud } from "lucide-react";
import { SecretPicker } from "@/components/secrets/secret-picker";
import { Key } from "lucide-react";

interface FilesystemConfigProps {
    service: UpstreamServiceConfig;
    onChange: (service: UpstreamServiceConfig) => void;
}

export function FilesystemConfig({ service, onChange }: FilesystemConfigProps) {
    const [fsConfig, setFsConfig] = useState<FilesystemUpstreamService>(
        service.filesystemService || {
            rootPaths: {},
            readOnly: false,
            allowedPaths: [],
            deniedPaths: [],
            symlinkMode: FilesystemUpstreamService_SymlinkMode.SYMLINK_MODE_UNSPECIFIED,
            tools: [],
            resources: [],
            prompts: [],
            os: {}, // Default to OS
        }
    );

    useEffect(() => {
        if (!service.filesystemService) {
            updateParent(fsConfig);
        }
    }, []);

    const updateParent = (config: FilesystemUpstreamService) => {
        onChange({ ...service, filesystemService: config });
    };

    const handleConfigChange = (updates: Partial<FilesystemUpstreamService>) => {
        const newConfig = { ...fsConfig, ...updates };
        setFsConfig(newConfig);
        updateParent(newConfig);
    };

    const handleBackendTypeChange = (type: string) => {
        const newConfig = { ...fsConfig };
        // Clear previous backends
        delete newConfig.os;
        delete newConfig.s3;
        delete newConfig.gcs;
        delete newConfig.sftp;
        delete newConfig.tmpfs;
        delete newConfig.http;
        delete newConfig.zip;

        if (type === 'os') newConfig.os = {};
        if (type === 's3') newConfig.s3 = { bucket: "", region: "", accessKeyId: "", secretAccessKey: "", sessionToken: "", endpoint: "" };
        if (type === 'gcs') newConfig.gcs = { bucket: "" };
        if (type === 'sftp') newConfig.sftp = { address: "", username: "", password: "", keyPath: "" };

        setFsConfig(newConfig);
        updateParent(newConfig);
    };

    const getBackendType = () => {
        if (fsConfig.s3) return 's3';
        if (fsConfig.gcs) return 'gcs';
        if (fsConfig.sftp) return 'sftp';
        return 'os'; // Default
    };

    return (
        <div className="space-y-6">
            <Tabs defaultValue="general">
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="general">
                        <Folder className="mr-2 h-4 w-4" /> Mounts & General
                    </TabsTrigger>
                    <TabsTrigger value="backend">
                        <Cloud className="mr-2 h-4 w-4" /> Backend Storage
                    </TabsTrigger>
                    <TabsTrigger value="security">
                        <Shield className="mr-2 h-4 w-4" /> Security
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="general" className="space-y-4 pt-4">
                    <div className="flex items-center justify-between space-x-2 border p-4 rounded-md">
                        <div className="space-y-0.5">
                            <Label className="text-base">Read Only Mode</Label>
                            <p className="text-sm text-muted-foreground">
                                Prevent any write operations (create, update, delete) to the filesystem.
                            </p>
                        </div>
                        <Switch
                            checked={fsConfig.readOnly}
                            onCheckedChange={(checked) => handleConfigChange({ readOnly: checked })}
                        />
                    </div>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-base">Root Paths (Mounts)</CardTitle>
                            <CardDescription>
                                Map virtual paths (seen by LLM) to actual paths (on backend/storage).
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <KeyValueEditor
                                items={fsConfig.rootPaths || {}}
                                keyLabel="Virtual Path (e.g. /workspace)"
                                valueLabel="Physical Path (e.g. /home/user/data)"
                                onChange={(paths) => handleConfigChange({ rootPaths: paths })}
                            />
                        </CardContent>
                    </Card>

                    <div className="space-y-2">
                        <Label>Symlink Policy</Label>
                        <Select
                            value={fsConfig.symlinkMode.toString()}
                            onValueChange={(val) => handleConfigChange({ symlinkMode: parseInt(val) })}
                        >
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="0">Unspecified (Default)</SelectItem>
                                <SelectItem value="1">Allow All</SelectItem>
                                <SelectItem value="2">Deny All</SelectItem>
                                <SelectItem value="3">Internal Only (Same Root)</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </TabsContent>

                <TabsContent value="backend" className="space-y-4 pt-4">
                    <div className="space-y-2">
                        <Label>Storage Backend</Label>
                        <Select value={getBackendType()} onValueChange={handleBackendTypeChange}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select Backend" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="os">Local Filesystem (OS)</SelectItem>
                                <SelectItem value="s3">Amazon S3 / MinIO</SelectItem>
                                <SelectItem value="gcs">Google Cloud Storage</SelectItem>
                                <SelectItem value="sftp">SFTP / SSH</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    {fsConfig.os && (
                        <div className="p-4 border rounded-md bg-muted/50 text-sm text-muted-foreground">
                            Uses the local server filesystem. Ensure the server process has appropriate permissions.
                        </div>
                    )}

                    {fsConfig.s3 && (
                        <S3Config
                            config={fsConfig.s3}
                            onChange={(s3) => handleConfigChange({ s3 })}
                        />
                    )}

                    {fsConfig.gcs && (
                        <div className="space-y-2 border p-4 rounded-md">
                            <Label>Bucket Name</Label>
                            <Input
                                value={fsConfig.gcs.bucket}
                                onChange={(e) => handleConfigChange({ gcs: { bucket: e.target.value } })}
                                placeholder="my-gcs-bucket"
                            />
                        </div>
                    )}

                    {fsConfig.sftp && (
                        <div className="space-y-4 border p-4 rounded-md">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>Address</Label>
                                    <Input
                                        value={fsConfig.sftp.address}
                                        onChange={(e) => handleConfigChange({ sftp: { ...fsConfig.sftp!, address: e.target.value } })}
                                        placeholder="sftp.example.com:22"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Username</Label>
                                    <Input
                                        value={fsConfig.sftp.username}
                                        onChange={(e) => handleConfigChange({ sftp: { ...fsConfig.sftp!, username: e.target.value } })}
                                        placeholder="user"
                                    />
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label>Password (Optional)</Label>
                                <Input
                                    type="password"
                                    value={fsConfig.sftp.password}
                                    onChange={(e) => handleConfigChange({ sftp: { ...fsConfig.sftp!, password: e.target.value } })}
                                    placeholder="password"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label>Private Key Path (Optional)</Label>
                                <Input
                                    value={fsConfig.sftp.keyPath}
                                    onChange={(e) => handleConfigChange({ sftp: { ...fsConfig.sftp!, keyPath: e.target.value } })}
                                    placeholder="/path/to/id_rsa"
                                />
                            </div>
                        </div>
                    )}
                </TabsContent>

                <TabsContent value="security" className="space-y-4 pt-4">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-base">Allowed Paths (Glob Patterns)</CardTitle>
                            <CardDescription>
                                Whitelist specific paths. If empty, all paths under roots are allowed.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <StringListEditor
                                items={fsConfig.allowedPaths || []}
                                onChange={(paths) => handleConfigChange({ allowedPaths: paths })}
                                placeholder="e.g. *.txt, src/*"
                            />
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="text-base">Denied Paths (Glob Patterns)</CardTitle>
                            <CardDescription>
                                Blacklist specific paths. Checked after allowed paths.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <StringListEditor
                                items={fsConfig.deniedPaths || []}
                                onChange={(paths) => handleConfigChange({ deniedPaths: paths })}
                                placeholder="e.g. *.env, secrets/*"
                            />
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}

function S3Config({ config, onChange }: { config: S3Fs, onChange: (c: S3Fs) => void }) {
    const update = (field: keyof S3Fs, value: string) => onChange({ ...config, [field]: value });

    return (
        <div className="space-y-4 border p-4 rounded-md">
            <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                    <Label>Bucket</Label>
                    <Input value={config.bucket} onChange={(e) => update('bucket', e.target.value)} placeholder="my-bucket" />
                </div>
                <div className="space-y-2">
                    <Label>Region</Label>
                    <Input value={config.region} onChange={(e) => update('region', e.target.value)} placeholder="us-east-1" />
                </div>
            </div>
            <div className="space-y-2">
                <Label>Endpoint (Optional)</Label>
                <Input value={config.endpoint} onChange={(e) => update('endpoint', e.target.value)} placeholder="https://minio.example.com" />
            </div>
            <div className="space-y-2">
                <Label>Access Key ID</Label>
                <div className="flex gap-2">
                    <Input value={config.accessKeyId} onChange={(e) => update('accessKeyId', e.target.value)} placeholder="AKIA..." />
                    <SecretPicker onSelect={(key) => update('accessKeyId', `\${${key}}`)}>
                        <Button variant="outline" size="icon"><Key className="h-4 w-4" /></Button>
                    </SecretPicker>
                </div>
            </div>
            <div className="space-y-2">
                <Label>Secret Access Key</Label>
                <div className="flex gap-2">
                    <Input type="password" value={config.secretAccessKey} onChange={(e) => update('secretAccessKey', e.target.value)} placeholder="Secret..." />
                    <SecretPicker onSelect={(key) => update('secretAccessKey', `\${${key}}`)}>
                        <Button variant="outline" size="icon"><Key className="h-4 w-4" /></Button>
                    </SecretPicker>
                </div>
            </div>
        </div>
    );
}

function KeyValueEditor({ items, onChange, keyLabel, valueLabel }: {
    items: Record<string, string>,
    onChange: (items: Record<string, string>) => void,
    keyLabel?: string,
    valueLabel?: string
}) {
    const [entries, setEntries] = useState(Object.entries(items));

    useEffect(() => {
        setEntries(Object.entries(items));
    }, [items]);

    const update = (newEntries: [string, string][]) => {
        setEntries(newEntries);
        onChange(Object.fromEntries(newEntries));
    };

    return (
        <div className="space-y-2">
            <div className="grid grid-cols-[1fr,1fr,auto] gap-2 mb-2">
                <Label className="text-xs text-muted-foreground">{keyLabel || "Key"}</Label>
                <Label className="text-xs text-muted-foreground">{valueLabel || "Value"}</Label>
                <span className="w-8"></span>
            </div>
            {entries.map((entry, idx) => (
                <div key={idx} className="grid grid-cols-[1fr,1fr,auto] gap-2 items-center">
                    <Input
                        value={entry[0]}
                        onChange={(e) => {
                            const next = [...entries];
                            next[idx][0] = e.target.value;
                            update(next);
                        }}
                        placeholder="Key"
                    />
                    <Input
                        value={entry[1]}
                        onChange={(e) => {
                            const next = [...entries];
                            next[idx][1] = e.target.value;
                            update(next);
                        }}
                        placeholder="Value"
                    />
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => update(entries.filter((_, i) => i !== idx))}
                        className="text-muted-foreground hover:text-destructive"
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            <Button variant="outline" size="sm" onClick={() => update([...entries, ["", ""]])} className="w-full mt-2">
                <Plus className="mr-2 h-4 w-4" /> Add Path
            </Button>
        </div>
    );
}

function StringListEditor({ items, onChange, placeholder }: { items: string[], onChange: (items: string[]) => void, placeholder?: string }) {
    return (
        <div className="space-y-2">
            {items.map((item, idx) => (
                <div key={idx} className="flex gap-2 items-center">
                    <Input
                        value={item}
                        onChange={(e) => {
                            const next = [...items];
                            next[idx] = e.target.value;
                            onChange(next);
                        }}
                        placeholder={placeholder}
                    />
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onChange(items.filter((_, i) => i !== idx))}
                        className="text-muted-foreground hover:text-destructive"
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            ))}
            <Button variant="outline" size="sm" onClick={() => onChange([...items, ""])} className="w-full mt-2">
                <Plus className="mr-2 h-4 w-4" /> Add Pattern
            </Button>
        </div>
    );
}
