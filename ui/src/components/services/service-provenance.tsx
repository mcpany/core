/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ServiceProvenance } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, User, FileSignature } from "lucide-react";

interface ServiceProvenanceProps {
  provenance?: ServiceProvenance;
}

/**
 * ServiceProvenanceViewer displays the security attestation details for a service.
 */
export function ServiceProvenanceViewer({ provenance }: ServiceProvenanceProps) {
  if (!provenance || !provenance.verified) {
    return (
      <Card className="border-l-4 border-l-destructive/50 bg-destructive/5">
        <CardHeader>
          <div className="flex items-center gap-2">
            <ShieldAlert className="h-6 w-6 text-destructive" />
            <CardTitle className="text-destructive">Unverified Service</CardTitle>
          </div>
          <CardDescription>
            This service does not have a valid supply chain attestation. Use with caution.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-muted-foreground">
            <p>
              We could not verify the authenticity of this service configuration. It may have been modified or come from an untrusted source.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Format timestamp
  // protobuf timestamp is usually ISO string in JSON
  const attestationDate = provenance.attestationTime
    ? new Date(provenance.attestationTime as unknown as string).toLocaleString()
    : "Unknown";

  return (
    <div className="space-y-6">
      <Card className="border-l-4 border-l-green-500 bg-green-50/10">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldCheck className="h-6 w-6 text-green-500" />
              <CardTitle className="text-green-600 dark:text-green-400">Verified Service</CardTitle>
            </div>
            <Badge variant="outline" className="border-green-500 text-green-600 bg-green-50 dark:bg-green-900/20">
              Trusted
            </Badge>
          </div>
          <CardDescription>
            This service has a valid cryptographic signature and verified provenance.
          </CardDescription>
        </CardHeader>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base font-medium flex items-center gap-2">
              <User className="h-4 w-4 text-primary" /> Signer Identity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold tracking-tight">
              {provenance.signerIdentity || "Unknown Signer"}
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              The entity that signed this configuration.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base font-medium flex items-center gap-2">
              <Calendar className="h-4 w-4 text-primary" /> Attestation Time
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-mono">
              {attestationDate}
            </div>
             <p className="text-xs text-muted-foreground mt-1">
              When the signature was generated.
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-medium flex items-center gap-2">
            <Fingerprint className="h-4 w-4 text-primary" /> Cryptographic Details
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase flex items-center gap-1">
                  <FileSignature className="h-3 w-3" /> Algorithm
              </span>
              <div className="font-mono text-sm border p-2 rounded bg-muted/50">
                {provenance.signatureAlgorithm || "Unknown"}
              </div>
            </div>
             {/* Future: Add Signature Hash visualization if available in proto */}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
