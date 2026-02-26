/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, Users, AlertTriangle, Lock } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { useEffect, useState } from "react";

/**
 * ServiceProvenance displays the supply chain attestation details for a service.
 * It shows whether the service is verified, the signer identity, and cryptographic details.
 *
 * @param props - The component props.
 * @param props.service - The service configuration containing provenance data.
 * @returns The rendered component.
 */
export function ServiceProvenance({ service }: { service: UpstreamServiceConfig }) {
    const verified = service.provenance?.verified;
    const provenance = service.provenance;
    const [formattedDate, setFormattedDate] = useState<string | null>(null);

    useEffect(() => {
        if (provenance?.attestationTime) {
            setFormattedDate(new Date(provenance.attestationTime).toLocaleString());
        }
    }, [provenance?.attestationTime]);

    if (!verified || !provenance) {
        return (
            <div className="space-y-6">
                <Card className="border-l-4 border-l-destructive bg-destructive/5">
                    <CardHeader className="pb-2">
                        <div className="flex items-center gap-2 text-destructive">
                            <ShieldAlert className="h-6 w-6" />
                            <CardTitle>Unverified Service</CardTitle>
                        </div>
                        <CardDescription className="text-destructive/80">
                            This service does not have a valid supply chain attestation.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground mt-2">
                            The provenance of this service cannot be verified. Proceed with caution when using tools or resources from this service, as they may have been modified or compromised.
                        </p>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <Card className="border-l-4 border-l-green-500 bg-green-500/5">
                <CardHeader className="pb-2">
                    <div className="flex items-center gap-2 text-green-600">
                        <ShieldCheck className="h-6 w-6" />
                        <CardTitle>Verified Service</CardTitle>
                    </div>
                    <CardDescription className="text-green-700/80">
                        This service has a valid cryptographic signature and attestation.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                        <div className="flex items-start gap-3 p-3 bg-background rounded-lg border shadow-sm">
                            <Users className="h-5 w-5 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-sm font-medium">Signer Identity</p>
                                <p className="text-sm text-muted-foreground font-mono">{provenance.signerIdentity}</p>
                            </div>
                        </div>
                        <div className="flex items-start gap-3 p-3 bg-background rounded-lg border shadow-sm">
                            <Calendar className="h-5 w-5 text-muted-foreground mt-0.5" />
                            <div>
                                <p className="text-sm font-medium">Attestation Time</p>
                                <p className="text-sm text-muted-foreground font-mono">
                                    {formattedDate || "Loading..."}
                                </p>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle className="text-lg flex items-center gap-2">
                        <Fingerprint className="h-5 w-5" />
                        Cryptographic Details
                    </CardTitle>
                    <CardDescription>
                        Technical details of the signature and service identity.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-1">
                            <div className="flex items-center gap-2 text-muted-foreground mb-1">
                                <Lock className="h-3 w-3" />
                                <span className="text-sm font-medium">Signature Algorithm</span>
                            </div>
                            <code className="text-sm bg-muted px-2 py-1 rounded block w-full truncate border font-mono">
                                {provenance.signatureAlgorithm || "Unknown"}
                            </code>
                        </div>
                         <div className="space-y-1">
                            <div className="flex items-center gap-2 text-muted-foreground mb-1">
                                <Fingerprint className="h-3 w-3" />
                                <span className="text-sm font-medium">Service ID Hash</span>
                            </div>
                            <code className="text-sm bg-muted px-2 py-1 rounded block w-full truncate font-mono border">
                                {service.id}
                            </code>
                        </div>
                    </div>

                    <Separator />

                    <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted/20 p-3 rounded border border-dashed">
                        <AlertTriangle className="h-4 w-4 text-amber-500" />
                        <span>This attestation ensures the service configuration matches the signed manifest. Any modification to the service configuration will invalidate this attestation.</span>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
