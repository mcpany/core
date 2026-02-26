/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ServiceProvenance } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Fingerprint, Clock, Key, FileBadge, CheckCircle2, AlertTriangle } from "lucide-react";
import { format } from "date-fns";
import { Separator } from "@/components/ui/separator";

interface ServiceProvenanceProps {
  provenance?: ServiceProvenance;
  serviceName: string;
}

/**
 * ServiceProvenanceView component.
 * Displays the supply chain attestation details for a service.
 *
 * @param props - The component props.
 * @param props.provenance - The provenance data.
 * @param props.serviceName - The name of the service.
 * @returns The rendered component.
 */
export function ServiceProvenanceView({ provenance, serviceName }: ServiceProvenanceProps) {
  if (!provenance || !provenance.verified) {
    return (
      <div className="space-y-6">
        <Card className="border-l-4 border-l-destructive bg-destructive/5">
          <CardHeader>
            <div className="flex items-center gap-2 text-destructive">
              <ShieldAlert className="h-6 w-6" />
              <CardTitle>Unverified Service</CardTitle>
            </div>
            <CardDescription className="text-destructive/80">
              This service does not have a valid supply chain attestation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              The provenance of <strong>{serviceName}</strong> could not be verified.
              This means we cannot cryptographically guarantee its origin or integrity.
              Proceed with caution if this service handles sensitive data.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const attestationDate = provenance.attestationTime
    ? new Date(typeof provenance.attestationTime === 'string' ? provenance.attestationTime : Number(provenance.attestationTime))
    : null;

  return (
    <div className="space-y-6">
      <Card className="border-l-4 border-l-green-500 bg-green-50/50 dark:bg-green-900/10">
        <CardHeader>
          <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
            <ShieldCheck className="h-6 w-6" />
            <CardTitle>Verified Service</CardTitle>
          </div>
          <CardDescription className="text-green-700/80 dark:text-green-300/80">
            This service has a valid supply chain attestation.
          </CardDescription>
        </CardHeader>
        <CardContent>
            <div className="flex items-center gap-2">
                 <CheckCircle2 className="h-4 w-4 text-green-600 dark:text-green-400" />
                 <span className="text-sm font-medium">Identity verified by {provenance.signerIdentity}</span>
            </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <FileBadge className="h-4 w-4 text-muted-foreground" />
                    Signer Identity
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{provenance.signerIdentity}</div>
                <p className="text-xs text-muted-foreground mt-1">The entity that signed this service.</p>
            </CardContent>
        </Card>

        <Card>
            <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    Attestation Time
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">
                    {attestationDate ? format(attestationDate, "PP p") : "Unknown"}
                </div>
                <p className="text-xs text-muted-foreground mt-1">When the signature was generated.</p>
            </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
             <CardTitle className="flex items-center gap-2">
                 <Fingerprint className="h-5 w-5" />
                 Cryptographic Details
             </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase">Algorithm</span>
                    <div className="flex items-center gap-2 font-mono text-sm border rounded px-2 py-1 bg-muted/50 w-fit">
                        <Key className="h-3 w-3" />
                        {provenance.signatureAlgorithm || "Unknown"}
                    </div>
                </div>
                <div className="space-y-1 md:col-span-2">
                    <span className="text-xs font-medium text-muted-foreground uppercase">Status</span>
                    <div className="flex items-center gap-2">
                         <Badge variant="outline" className="border-green-500 text-green-600 bg-green-50 dark:bg-green-900/20">
                            Signature Valid
                         </Badge>
                         <Badge variant="outline">
                            Chain of Trust Intact
                         </Badge>
                    </div>
                </div>
            </div>

            <Separator />

            <div className="space-y-2">
                <span className="text-xs font-medium text-muted-foreground uppercase">Compliance</span>
                <div className="text-sm text-muted-foreground">
                    This service meets the <strong>MCP Supply Chain Security Level 1</strong> requirements.
                    The signature guarantees that the code has not been tampered with since it was published by the signer.
                </div>
            </div>
        </CardContent>
      </Card>
    </div>
  );
}
