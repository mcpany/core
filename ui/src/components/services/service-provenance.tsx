/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, User, Hash } from "lucide-react";

interface ServiceProvenanceProps {
  service: UpstreamServiceConfig;
}

/**
 * ServiceProvenance component.
 * Displays the supply chain attestation details for a service.
 *
 * @param props - The component props.
 * @param props.service - The service configuration.
 * @returns The rendered component.
 */
export function ServiceProvenance({ service }: ServiceProvenanceProps) {
  const provenance = service.provenance;
  const isVerified = provenance?.verified;

  if (!provenance) {
    return (
      <div className="space-y-6 p-6">
        <Alert variant="destructive">
          <ShieldAlert className="h-4 w-4" />
          <AlertTitle>Unverified Service</AlertTitle>
          <AlertDescription>
            This service has no provenance information. Its origin and integrity cannot be verified.
            Proceed with caution.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      <div className="grid gap-6 md:grid-cols-2">
        <Card className={isVerified ? "border-green-200 bg-green-50/50 dark:border-green-900 dark:bg-green-900/10" : "border-yellow-200 bg-yellow-50/50 dark:border-yellow-900 dark:bg-yellow-900/10"}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <CardTitle className="flex items-center gap-2">
                  {isVerified ? (
                    <>
                      <ShieldCheck className="h-5 w-5 text-green-600" />
                      <span className="text-green-700 dark:text-green-400">Verified Service</span>
                    </>
                  ) : (
                    <>
                      <ShieldAlert className="h-5 w-5 text-yellow-600" />
                      <span className="text-yellow-700 dark:text-yellow-400">Unverified Service</span>
                    </>
                  )}
                </CardTitle>
                <CardDescription>
                  {isVerified
                    ? "This service has been cryptographically signed and verified."
                    : "The signature for this service could not be verified."}
                </CardDescription>
              </div>
              <Badge variant={isVerified ? "default" : "secondary"} className={isVerified ? "bg-green-600 hover:bg-green-700" : ""}>
                {isVerified ? "Trusted" : "Untrusted"}
              </Badge>
            </div>
          </CardHeader>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <Fingerprint className="h-5 w-5 text-primary" />
              Attestation Details
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <User className="h-4 w-4" />
                Signer Identity
              </div>
              <span className="font-medium">{provenance.signerIdentity || "Unknown"}</span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Calendar className="h-4 w-4" />
                Attestation Time
              </div>
              <span className="font-medium">
                {provenance.attestationTime
                  ? new Date(provenance.attestationTime as any).toLocaleString()
                  : "N/A"}
              </span>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Hash className="h-4 w-4" />
                Algorithm
              </div>
              <Badge variant="outline" className="font-mono">
                {provenance.signatureAlgorithm || "N/A"}
              </Badge>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Supply Chain Information</CardTitle>
          <CardDescription>
            Detailed information about the software supply chain for this service.
          </CardDescription>
        </CardHeader>
        <CardContent>
             <div className="rounded-md bg-muted p-4 font-mono text-sm">
                 <p className="text-muted-foreground mb-2">// Raw Provenance Data</p>
                 <pre className="whitespace-pre-wrap break-all">
                     {JSON.stringify(provenance, null, 2)}
                 </pre>
             </div>
        </CardContent>
      </Card>
    </div>
  );
}
