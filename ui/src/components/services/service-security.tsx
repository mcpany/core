/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Lock, Unlock, Network, FileText, Terminal, Database, Clock } from "lucide-react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";

interface ServiceSecurityProps {
    service: UpstreamServiceConfig;
}

/**
 * ServiceSecurity component displays the security posture of an upstream service.
 * It visualizes:
 * - Provenance & Attestation (Signed by who?)
 * - Export Policies (What is exposed?)
 * - Network Rules (Call Policies)
 *
 * @param props - The component props.
 * @param props.service - The service configuration to analyze.
 * @returns The rendered security dashboard.
 */
export function ServiceSecurity({ service }: ServiceSecurityProps) {
    const provenance = service.provenance;
    const isVerified = provenance?.verified;

    const renderPolicyAction = (action: number) => {
        // 1 = EXPORT, 2 = UNEXPORT
        if (action === 1) return <Badge variant="outline" className="text-green-500 border-green-500/30 bg-green-500/10">Export</Badge>;
        if (action === 2) return <Badge variant="outline" className="text-red-500 border-red-500/30 bg-red-500/10">Block</Badge>;
        return <Badge variant="outline" className="text-muted-foreground">Unspecified</Badge>;
    };

    const renderCallAction = (action: number) => {
        // 0 = ALLOW, 1 = DENY
        if (action === 0) return <Badge variant="outline" className="text-green-500 border-green-500/30 bg-green-500/10">Allow</Badge>;
        if (action === 1) return <Badge variant="outline" className="text-red-500 border-red-500/30 bg-red-500/10">Deny</Badge>;
        return <Badge variant="outline">Unknown</Badge>;
    };

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            {/* 1. Provenance Card */}
            <Card className={`border-l-4 ${isVerified ? 'border-l-green-500' : 'border-l-amber-500'}`}>
                <CardHeader className="pb-2">
                    <div className="flex justify-between items-start">
                        <div>
                            <CardTitle className="flex items-center gap-2 text-lg">
                                {isVerified ? (
                                    <ShieldCheck className="h-5 w-5 text-green-500" />
                                ) : (
                                    <ShieldAlert className="h-5 w-5 text-amber-500" />
                                )}
                                Supply Chain Attestation
                            </CardTitle>
                            <CardDescription>
                                Verify the origin and integrity of this service configuration.
                            </CardDescription>
                        </div>
                        {isVerified ? (
                            <Badge className="bg-green-500 hover:bg-green-600">Verified</Badge>
                        ) : (
                            <Badge variant="secondary" className="text-amber-600 bg-amber-100">Unverified</Badge>
                        )}
                    </div>
                </CardHeader>
                <CardContent>
                    {isVerified ? (
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-sm">
                            <div className="space-y-1">
                                <span className="text-muted-foreground text-xs uppercase tracking-wider font-semibold">Signer Identity</span>
                                <div className="font-medium flex items-center gap-2">
                                    <Badge variant="outline" className="font-mono text-xs">{provenance?.signerIdentity}</Badge>
                                </div>
                            </div>
                            <div className="space-y-1">
                                <span className="text-muted-foreground text-xs uppercase tracking-wider font-semibold">Attestation Time</span>
                                <div className="font-medium flex items-center gap-2">
                                    <Clock className="h-3 w-3 text-muted-foreground" />
                                    {provenance?.attestationTime ? new Date(provenance.attestationTime as any).toLocaleString() : "Unknown"}
                                </div>
                            </div>
                            <div className="space-y-1">
                                <span className="text-muted-foreground text-xs uppercase tracking-wider font-semibold">Algorithm</span>
                                <div className="font-mono text-xs text-muted-foreground">
                                    {provenance?.signatureAlgorithm || "ECDSA-SHA256"}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="text-sm text-muted-foreground bg-muted/30 p-4 rounded-md border border-dashed flex items-center gap-3">
                            <ShieldAlert className="h-8 w-8 text-muted-foreground opacity-50" />
                            <div>
                                <p className="font-medium text-foreground">No cryptographic signature found.</p>
                                <p>This configuration has not been signed by a trusted authority. Proceed with caution.</p>
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 2. Export Policies */}
                <Card className="flex flex-col">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <Lock className="h-4 w-4 text-primary" />
                            Data Loss Prevention (DLP)
                        </CardTitle>
                        <CardDescription>
                            Control which internal capabilities are exposed to the AI model.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1 p-0">
                        <ScrollArea className="h-[300px]">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-[40%]">Type</TableHead>
                                        <TableHead>Rule</TableHead>
                                        <TableHead className="text-right">Action</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {/* Tool Policy */}
                                    <TableRow className="bg-muted/5">
                                        <TableCell className="font-medium flex items-center gap-2">
                                            <Terminal className="h-3 w-3" /> Tools
                                        </TableCell>
                                        <TableCell className="text-xs text-muted-foreground">Default</TableCell>
                                        <TableCell className="text-right">
                                            {renderPolicyAction(service.toolExportPolicy?.defaultAction || 0)}
                                        </TableCell>
                                    </TableRow>
                                    {service.toolExportPolicy?.rules?.map((rule, i) => (
                                        <TableRow key={`tool-${i}`}>
                                            <TableCell className="pl-8 text-xs text-muted-foreground">Tool Match</TableCell>
                                            <TableCell className="font-mono text-xs">{rule.nameRegex}</TableCell>
                                            <TableCell className="text-right">{renderPolicyAction(rule.action)}</TableCell>
                                        </TableRow>
                                    ))}

                                    {/* Resource Policy */}
                                    <TableRow className="bg-muted/5">
                                        <TableCell className="font-medium flex items-center gap-2">
                                            <Database className="h-3 w-3" /> Resources
                                        </TableCell>
                                        <TableCell className="text-xs text-muted-foreground">Default</TableCell>
                                        <TableCell className="text-right">
                                            {renderPolicyAction(service.resourceExportPolicy?.defaultAction || 0)}
                                        </TableCell>
                                    </TableRow>
                                    {service.resourceExportPolicy?.rules?.map((rule, i) => (
                                        <TableRow key={`res-${i}`}>
                                            <TableCell className="pl-8 text-xs text-muted-foreground">Resource Match</TableCell>
                                            <TableCell className="font-mono text-xs">{rule.nameRegex}</TableCell>
                                            <TableCell className="text-right">{renderPolicyAction(rule.action)}</TableCell>
                                        </TableRow>
                                    ))}

                                    {/* Prompt Policy */}
                                    <TableRow className="bg-muted/5">
                                        <TableCell className="font-medium flex items-center gap-2">
                                            <FileText className="h-3 w-3" /> Prompts
                                        </TableCell>
                                        <TableCell className="text-xs text-muted-foreground">Default</TableCell>
                                        <TableCell className="text-right">
                                            {renderPolicyAction(service.promptExportPolicy?.defaultAction || 0)}
                                        </TableCell>
                                    </TableRow>
                                </TableBody>
                            </Table>
                        </ScrollArea>
                    </CardContent>
                </Card>

                {/* 3. Network Rules */}
                <Card className="flex flex-col">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <Network className="h-4 w-4 text-primary" />
                            Network & Call Policies
                        </CardTitle>
                        <CardDescription>
                            Restrict specific API calls based on arguments or regex.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1 p-0">
                        <ScrollArea className="h-[300px]">
                            {(!service.callPolicies || service.callPolicies.length === 0) ? (
                                <div className="flex flex-col items-center justify-center h-full p-8 text-muted-foreground text-sm border-dashed">
                                    <Unlock className="h-8 w-8 mb-2 opacity-20" />
                                    <p>No active call policies.</p>
                                    <p className="text-xs">All calls are allowed by default.</p>
                                </div>
                            ) : (
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Target</TableHead>
                                            <TableHead>Condition</TableHead>
                                            <TableHead className="text-right">Effect</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {service.callPolicies.map((policy, pIdx) => (
                                            <>
                                                <TableRow key={`p-${pIdx}`} className="bg-muted/5">
                                                    <TableCell colSpan={2} className="font-medium text-xs">
                                                        Policy Set #{pIdx + 1}
                                                    </TableCell>
                                                    <TableCell className="text-right">
                                                        <span className="text-[10px] text-muted-foreground mr-1">Default:</span>
                                                        {renderCallAction(policy.defaultAction)}
                                                    </TableCell>
                                                </TableRow>
                                                {policy.rules?.map((rule, rIdx) => (
                                                    <TableRow key={`p-${pIdx}-r-${rIdx}`}>
                                                        <TableCell className="font-mono text-xs">
                                                            {rule.nameRegex || "*"}
                                                        </TableCell>
                                                        <TableCell className="text-xs text-muted-foreground max-w-[150px] truncate">
                                                            {rule.argumentRegex && <span className="block" title="Args Regex">Args: {rule.argumentRegex}</span>}
                                                            {rule.urlRegex && <span className="block" title="URL Regex">URL: {rule.urlRegex}</span>}
                                                        </TableCell>
                                                        <TableCell className="text-right">
                                                            {renderCallAction(rule.action)}
                                                        </TableCell>
                                                    </TableRow>
                                                ))}
                                            </>
                                        ))}
                                    </TableBody>
                                </Table>
                            )}
                        </ScrollArea>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
