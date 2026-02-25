/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Calendar, User, Lock, FileSignature } from "lucide-react";
import { format } from "date-fns";

interface ServiceProvenanceProps {
    service: UpstreamServiceConfig;
}

/**
 * ServiceProvenance component.
 * Displays the supply chain attestation and verification status of a service.
 *
 * @param props - The component props.
 * @param props.service - The upstream service configuration.
 * @returns The rendered component.
 */
export function ServiceProvenance({ service }: ServiceProvenanceProps) {
    const provenance = service.provenance;
    const isVerified = provenance?.verified || false;

    // Helper to parse date safely
    const parseDate = (date: any) => {
        if (!date) return null;
        try {
            return new Date(date);
        } catch {
            return null;
        }
    };

    const attestationDate = parseDate(provenance?.attestationTime);

    return (
        <div className="space-y-6 p-6">
            {/* Status Banner */}
            <div className={`rounded-lg border p-6 flex items-start gap-4 ${isVerified ? "bg-green-50/50 border-green-200 dark:bg-green-900/10 dark:border-green-900" : "bg-yellow-50/50 border-yellow-200 dark:bg-yellow-900/10 dark:border-yellow-900"}`}>
                {isVerified ? (
                    <ShieldCheck className="h-8 w-8 text-green-600 mt-1" />
                ) : (
                    <ShieldAlert className="h-8 w-8 text-yellow-600 mt-1" />
                )}
                <div>
                    <h3 className={`text-lg font-semibold ${isVerified ? "text-green-700 dark:text-green-400" : "text-yellow-700 dark:text-yellow-400"}`}>
                        {isVerified ? "Verified Service" : "Unverified Service"}
                    </h3>
                    <p className="text-sm text-muted-foreground mt-1">
                        {isVerified
                            ? "This service has been cryptographically verified and attested by a trusted authority."
                            : "This service does not have a valid supply chain attestation. Exercise caution when using tools from this source."}
                    </p>
                </div>
            </div>

            {/* Details Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <User className="h-4 w-4" /> Signer Identity
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        {isVerified ? (
                            <div className="flex flex-col gap-1">
                                <span className="text-2xl font-bold">{provenance?.signerIdentity}</span>
                                <span className="text-xs text-muted-foreground">Trusted Authority</span>
                            </div>
                        ) : (
                            <span className="text-muted-foreground italic">Unknown Identity</span>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Calendar className="h-4 w-4" /> Attestation Time
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        {isVerified && attestationDate ? (
                            <div className="flex flex-col gap-1">
                                <span className="text-xl font-mono">
                                    {format(attestationDate, "yyyy-MM-dd HH:mm:ss")}
                                </span>
                                <span className="text-xs text-muted-foreground">
                                    {format(attestationDate, "zzzz")}
                                </span>
                            </div>
                        ) : (
                            <span className="text-muted-foreground italic">No timestamp available</span>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Lock className="h-4 w-4" /> Signature Algorithm
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        {isVerified ? (
                            <Badge variant="outline" className="font-mono">
                                {provenance?.signatureAlgorithm || "ECDSA-SHA256"}
                            </Badge>
                        ) : (
                            <span className="text-muted-foreground italic">N/A</span>
                        )}
                    </CardContent>
                </Card>

                 <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <FileSignature className="h-4 w-4" /> Provenance Source
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                         {isVerified ? (
                            <div className="flex items-center gap-2">
                                <Badge variant="secondary">mcp-registry</Badge>
                                <span className="text-xs text-muted-foreground">v1.0.0</span>
                            </div>
                         ) : (
                            <span className="text-muted-foreground italic">Local / Manual</span>
                         )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
