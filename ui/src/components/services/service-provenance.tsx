/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, User, Lock, Award } from "lucide-react";

interface ServiceProvenanceProps {
  service: UpstreamServiceConfig;
}

/**
 * ServiceProvenance component displays the supply chain attestation details for a service.
 * @param props - The component props.
 * @param props.service - The upstream service configuration.
 * @returns The rendered component.
 */
export function ServiceProvenance({ service }: ServiceProvenanceProps) {
  const provenance = service.provenance;
  const isVerified = provenance?.verified;

  return (
    <div className="space-y-6">
      <Card className={isVerified ? "border-green-200 bg-green-50/20" : "border-amber-200 bg-amber-50/20"}>
        <CardHeader>
          <div className="flex items-center gap-2">
            {isVerified ? (
              <ShieldCheck className="h-8 w-8 text-green-600" />
            ) : (
              <ShieldAlert className="h-8 w-8 text-amber-600" />
            )}
            <div>
              <CardTitle className={isVerified ? "text-green-700" : "text-amber-700"}>
                {isVerified ? "Verified Source" : "Unverified Source"}
              </CardTitle>
              <CardDescription>
                {isVerified
                  ? "This service has been cryptographically verified and attested by a trusted signer."
                  : "This service lacks provenance information. Proceed with caution."}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>

      {isVerified && provenance && (
        <div className="grid gap-6 md:grid-cols-2">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <User className="h-4 w-4 text-muted-foreground" />
                Signer Identity
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{provenance.signerIdentity}</div>
              <p className="text-xs text-muted-foreground mt-1">
                The entity that signed and verified this service configuration.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                Attestation Time
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {provenance.attestationTime
                  ? new Date(provenance.attestationTime).toISOString().split('T')[0]
                  : "Unknown"}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {provenance.attestationTime
                  ? new Date(provenance.attestationTime).toISOString().split('T')[1].replace('Z', ' UTC')
                  : ""}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Lock className="h-4 w-4 text-muted-foreground" />
                Signature Algorithm
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Badge variant="outline" className="font-mono">
                {provenance.signatureAlgorithm || "Unknown"}
              </Badge>
              <p className="text-xs text-muted-foreground mt-1">
                Cryptographic algorithm used for the signature.
              </p>
            </CardContent>
          </Card>

           <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium flex items-center gap-2">
                <Award className="h-4 w-4 text-muted-foreground" />
                Certificate Status
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2">
                  <Badge className="bg-green-600 hover:bg-green-700">Valid</Badge>
                  <span className="text-sm text-muted-foreground">Chain of Trust intact</span>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {!isVerified && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Fingerprint className="h-4 w-4 text-muted-foreground" />
              Service Identity
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-col gap-1">
                <span className="text-xs font-semibold uppercase text-muted-foreground">Service ID</span>
                <code className="bg-muted p-2 rounded text-sm font-mono break-all">{service.id || "N/A"}</code>
            </div>
             <div className="flex flex-col gap-1">
                <span className="text-xs font-semibold uppercase text-muted-foreground">Service Name</span>
                <span className="font-medium">{service.name}</span>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
