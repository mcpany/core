/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Copy, Eye, EyeOff, Server, Shield, Terminal, Globe, Database, Lock, Settings } from "lucide-react";
import { useState } from "react";
import { useToast } from "@/hooks/use-toast";

interface ServiceConfigViewerProps {
    service: UpstreamServiceConfig;
}

function CopyButton({ value, label }: { value: string; label: string }) {
    const { toast } = useToast();
    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 ml-2"
            onClick={() => {
                navigator.clipboard.writeText(value);
                toast({ title: "Copied", description: `${label} copied to clipboard.` });
            }}
            title={`Copy ${label}`}
        >
            <Copy className="h-3 w-3" />
        </Button>
    );
}

function SecretViewer({ value, label }: { value: string; label: string }) {
    const [revealed, setRevealed] = useState(false);
    const isSecretRef = value.startsWith("${") && value.endsWith("}");

    if (!value) return <span className="text-muted-foreground italic">Not set</span>;

    return (
        <div className="flex items-center gap-2">
            <code className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono break-all">
                {isSecretRef ? (
                    <span className="text-yellow-600 dark:text-yellow-400">{value}</span>
                ) : revealed ? (
                    value
                ) : (
                    "••••••••••••••••"
                )}
            </code>
            {!isSecretRef && (
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => setRevealed(!revealed)}
                    title={revealed ? "Hide" : "Show"}
                >
                    {revealed ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                </Button>
            )}
            <CopyButton value={value} label={label} />
        </div>
    );
}

function ConfigRow({ label, value, secret = false }: { label: string; value: React.ReactNode; secret?: boolean }) {
    return (
        <div className="grid grid-cols-1 sm:grid-cols-3 py-3 border-b last:border-0">
            <dt className="font-medium text-sm text-muted-foreground self-center">{label}</dt>
            <dd className="sm:col-span-2 text-sm flex items-center">
                {secret && typeof value === "string" ? (
                    <SecretViewer value={value} label={label} />
                ) : (
                    value
                )}
            </dd>
        </div>
    );
}

export function ServiceConfigViewer({ service }: ServiceConfigViewerProps) {
    const type = service.httpService ? "HTTP" :
                 service.grpcService ? "gRPC" :
                 service.commandLineService ? "Command Line" :
                 service.mcpService ? "MCP Proxy" :
                 service.openapiService ? "OpenAPI" : "Unknown";

    return (
        <div className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
                {/* Identity & General Info */}
                <Card>
                    <CardHeader className="pb-3">
                        <div className="flex items-center gap-2">
                            <Settings className="h-5 w-5 text-primary" />
                            <CardTitle className="text-lg">Service Identity</CardTitle>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <dl>
                            <ConfigRow label="Name" value={service.name} />
                            <ConfigRow label="ID" value={<code className="bg-muted px-1 rounded">{service.id}</code>} />
                            <ConfigRow label="Version" value={service.version} />
                            <ConfigRow label="Priority" value={service.priority} />
                            <ConfigRow label="Status" value={
                                <Badge variant={service.disable ? "secondary" : "default"}>
                                    {service.disable ? "Disabled" : "Enabled"}
                                </Badge>
                            } />
                            <ConfigRow label="Tags" value={
                                <div className="flex gap-1 flex-wrap">
                                    {service.tags?.map(t => <Badge key={t} variant="outline" className="text-xs">{t}</Badge>) || "-"}
                                </div>
                            } />
                        </dl>
                    </CardContent>
                </Card>

                {/* Connection Details */}
                <Card>
                    <CardHeader className="pb-3">
                        <div className="flex items-center gap-2">
                            <Server className="h-5 w-5 text-primary" />
                            <CardTitle className="text-lg">Connection Details</CardTitle>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <dl>
                            <ConfigRow label="Service Type" value={<Badge>{type}</Badge>} />

                            {service.httpService && (
                                <ConfigRow label="Base URL" value={
                                    <div className="flex items-center">
                                        <a href={service.httpService.address} target="_blank" rel="noreferrer" className="text-primary hover:underline font-mono break-all">
                                            {service.httpService.address}
                                        </a>
                                        <CopyButton value={service.httpService.address} label="URL" />
                                    </div>
                                } />
                            )}

                            {service.grpcService && (
                                <>
                                    <ConfigRow label="Address" value={<code className="font-mono">{service.grpcService.address}</code>} />
                                    <ConfigRow label="Reflection" value={service.grpcService.useReflection ? "Enabled" : "Disabled"} />
                                </>
                            )}

                            {service.commandLineService && (
                                <>
                                    <ConfigRow label="Command" value={<code className="font-mono bg-muted p-1 rounded break-all">{service.commandLineService.command}</code>} />
                                    <ConfigRow label="Working Dir" value={<code className="font-mono">{service.commandLineService.workingDirectory || "(Default)"}</code>} />
                                </>
                            )}

                            {service.openapiService && (
                                <>
                                    <ConfigRow label="Base URL" value={service.openapiService.address} />
                                    <ConfigRow label="Spec URL" value={
                                        <a href={service.openapiService.specUrl} target="_blank" rel="noreferrer" className="text-primary hover:underline break-all">
                                            {service.openapiService.specUrl}
                                        </a>
                                    } />
                                </>
                            )}

                            {service.mcpService && (
                                <>
                                    <ConfigRow label="Auto-Discovery" value={service.mcpService.toolAutoDiscovery ? "Enabled" : "Disabled"} />
                                    {service.mcpService.httpConnection && (
                                        <ConfigRow label="MCP URL" value={service.mcpService.httpConnection.httpAddress} />
                                    )}
                                    {service.mcpService.stdioConnection && (
                                        <ConfigRow label="MCP Command" value={service.mcpService.stdioConnection.command} />
                                    )}
                                </>
                            )}
                        </dl>
                    </CardContent>
                </Card>
            </div>

            {/* Authentication */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center gap-2">
                        <Lock className="h-5 w-5 text-primary" />
                        <CardTitle className="text-lg">Authentication</CardTitle>
                    </div>
                    <CardDescription>Credentials used to authenticate with the upstream service.</CardDescription>
                </CardHeader>
                <CardContent>
                    {!service.upstreamAuth ? (
                        <div className="text-muted-foreground italic p-4 text-center bg-muted/20 rounded">
                            No authentication configured.
                        </div>
                    ) : (
                        <dl>
                            {service.upstreamAuth.apiKey && (
                                <>
                                    <ConfigRow label="Type" value={<Badge variant="outline">API Key</Badge>} />
                                    <ConfigRow label="Parameter Name" value={<code className="font-mono">{service.upstreamAuth.apiKey.paramName}</code>} />
                                    <ConfigRow label="Location" value={
                                        service.upstreamAuth.apiKey.in === 0 ? "Header" :
                                        service.upstreamAuth.apiKey.in === 1 ? "Query Param" : "Cookie"
                                    } />
                                    <ConfigRow label="Value" value={service.upstreamAuth.apiKey.value?.plainText || ""} secret />
                                </>
                            )}
                            {service.upstreamAuth.bearerToken && (
                                <>
                                    <ConfigRow label="Type" value={<Badge variant="outline">Bearer Token</Badge>} />
                                    <ConfigRow label="Token" value={service.upstreamAuth.bearerToken.token?.plainText || ""} secret />
                                </>
                            )}
                            {service.upstreamAuth.oauth2 && (
                                <>
                                    <ConfigRow label="Type" value={<Badge variant="outline">OAuth 2.0</Badge>} />
                                    <ConfigRow label="Client ID" value={service.upstreamAuth.oauth2.clientId?.plainText || ""} />
                                    <ConfigRow label="Client Secret" value={service.upstreamAuth.oauth2.clientSecret?.plainText || ""} secret />
                                    <ConfigRow label="Token URL" value={service.upstreamAuth.oauth2.tokenUrl} />
                                    <ConfigRow label="Scopes" value={service.upstreamAuth.oauth2.scopes} />
                                </>
                            )}
                        </dl>
                    )}
                </CardContent>
            </Card>

            {/* Environment Variables (CLI Only) */}
            {service.commandLineService && service.commandLineService.env && Object.keys(service.commandLineService.env).length > 0 && (
                <Card>
                    <CardHeader className="pb-3">
                        <div className="flex items-center gap-2">
                            <Terminal className="h-5 w-5 text-primary" />
                            <CardTitle className="text-lg">Environment Variables</CardTitle>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[30%]">Key</TableHead>
                                    <TableHead>Value</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {Object.entries(service.commandLineService.env).map(([key, value]) => (
                                    <TableRow key={key}>
                                        <TableCell className="font-mono font-medium">{key}</TableCell>
                                        <TableCell>
                                            <SecretViewer value={value as string} label={key} />
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {/* Policies */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center gap-2">
                        <Shield className="h-5 w-5 text-primary" />
                        <CardTitle className="text-lg">Policies</CardTitle>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="grid gap-4 md:grid-cols-3">
                        {[
                            { title: "Tool Export", policy: service.toolExportPolicy },
                            { title: "Prompt Export", policy: service.promptExportPolicy },
                            { title: "Resource Export", policy: service.resourceExportPolicy }
                        ].map(({ title, policy }) => (
                            <div key={title} className="border rounded p-3">
                                <h4 className="font-medium text-sm mb-2">{title}</h4>
                                {policy ? (
                                    <div className="space-y-1">
                                        <div className="text-xs text-muted-foreground">Allowed: <span className="text-foreground">{policy.allowList?.join(", ") || "None"}</span></div>
                                        <div className="text-xs text-muted-foreground">Blocked: <span className="text-foreground">{policy.blockList?.join(", ") || "None"}</span></div>
                                    </div>
                                ) : (
                                    <span className="text-xs text-muted-foreground italic">Allow All</span>
                                )}
                            </div>
                        ))}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
