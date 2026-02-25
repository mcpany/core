/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, FileKey, CheckCircle, AlertTriangle } from "lucide-react";
import { format } from "date-fns";

interface ServiceProvenanceProps {
  service: UpstreamServiceConfig;
}

export function ServiceProvenance({ service }: ServiceProvenanceProps) {
  const { provenance } = service;

  if (!provenance || !provenance.verified) {
    return (
      <div className="h-full p-6">
        <Card className="border-destructive/50 bg-destructive/5">
          <CardHeader>
            <div className="flex items-center gap-2 text-destructive">
              <ShieldAlert className="h-6 w-6" />
              <CardTitle>Unverified Service</CardTitle>
            </div>
            <CardDescription className="text-destructive/80">
              This service has no supply chain attestation or verification could not be confirmed.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 p-4 rounded-md border border-destructive/20 bg-background/50 text-sm">
              <AlertTriangle className="h-4 w-4 text-destructive" />
              <p>Proceed with caution. The origin and integrity of this service cannot be guaranteed.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Format date safely
  let attestationDate = "Unknown";
  try {
      if (provenance.attestationTime) {
          // protobuf timestamp might be string or object depending on serialization
          const date = new Date(provenance.attestationTime as any);
          attestationDate = format(date, "PPP pp");
      }
  } catch (e) {
      console.warn("Failed to format date", e);
  }

  return (
    <div className="h-full p-6 space-y-6">
      <Card className="border-green-500/50 bg-green-500/5">
        <CardHeader>
          <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
            <ShieldCheck className="h-6 w-6" />
            <CardTitle>Verified Service</CardTitle>
          </div>
          <CardDescription className="text-green-700/80 dark:text-green-300/80">
            This service has a valid supply chain attestation and signature.
          </CardDescription>
        </CardHeader>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Fingerprint className="h-4 w-4 text-primary" />
              Identity
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground uppercase font-semibold">Signer</p>
              <p className="font-mono text-lg font-medium">{provenance.signerIdentity || "Unknown"}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Calendar className="h-4 w-4 text-primary" />
              Attestation Time
            </CardTitle>
          </CardHeader>
          <CardContent>
             <div className="space-y-1">
              <p className="text-xs text-muted-foreground uppercase font-semibold">Timestamp</p>
              <p className="font-mono text-lg font-medium">{attestationDate}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <FileKey className="h-4 w-4 text-primary" />
              Signature Details
            </CardTitle>
          </CardHeader>
          <CardContent>
             <div className="space-y-1">
              <p className="text-xs text-muted-foreground uppercase font-semibold">Algorithm</p>
              <Badge variant="outline" className="font-mono">
                  {provenance.signatureAlgorithm || "Unknown"}
              </Badge>
            </div>
          </CardContent>
        </Card>

         <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-primary" />
              Status
            </CardTitle>
          </CardHeader>
          <CardContent>
             <div className="space-y-1">
              <p className="text-xs text-muted-foreground uppercase font-semibold">Integrity</p>
              <div className="flex items-center gap-2 text-green-600 font-medium">
                  <CheckCircle className="h-4 w-4" />
                  Valid Signature
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
