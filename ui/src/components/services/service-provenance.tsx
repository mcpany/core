"use client";

import { UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Fingerprint, Calendar, FileKey, ExternalLink } from "lucide-react";
import { Separator } from "@/components/ui/separator";

interface ServiceProvenanceProps {
  service: UpstreamServiceConfig;
}

export function ServiceProvenance({ service }: ServiceProvenanceProps) {
  const verified = service.provenance?.verified;
  const signer = service.provenance?.signerIdentity;
  const time = service.provenance?.attestationTime;
  const algorithm = service.provenance?.signatureAlgorithm;

  let formattedTime = "Unknown";
  if (time) {
      if (typeof time === 'string') {
          formattedTime = new Date(time).toLocaleString();
      } else if (typeof time === 'object' && 'seconds' in (time as any)) {
          // Protobuf Timestamp
          const ts = time as { seconds: number, nanos?: number };
          formattedTime = new Date(ts.seconds * 1000).toLocaleString();
      } else if (time instanceof Date) {
            formattedTime = time.toLocaleString();
      }
  }

  return (
    <div className="space-y-6">
      <Card className={verified ? "border-green-200 bg-green-50/50 dark:bg-green-900/10" : "border-amber-200 bg-amber-50/50 dark:bg-amber-900/10"}>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-4">
            <div className={verified ? "bg-green-100 p-3 rounded-full" : "bg-amber-100 p-3 rounded-full"}>
                {verified ? <ShieldCheck className="h-8 w-8 text-green-600" /> : <ShieldAlert className="h-8 w-8 text-amber-600" />}
            </div>
            <div>
              <CardTitle className="text-xl flex items-center gap-2">
                {verified ? "Verified Source" : "Unverified Source"}
                {verified && <Badge variant="default" className="bg-green-600 hover:bg-green-700">Trusted</Badge>}
              </CardTitle>
              <CardDescription>
                {verified
                  ? `This service has been cryptographically verified by ${signer || "Unknown Signer"}.`
                  : "The origin of this service cannot be verified. Proceed with caution."}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>

      {verified && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
                <FileKey className="h-4 w-4" /> Attestation Details
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-6 md:grid-cols-2">
            <div className="space-y-1">
                <span className="text-xs font-medium text-muted-foreground uppercase flex items-center gap-1">
                    <Fingerprint className="h-3 w-3" /> Signer Identity
                </span>
                <p className="font-mono text-sm">{signer}</p>
            </div>
            <div className="space-y-1">
                <span className="text-xs font-medium text-muted-foreground uppercase flex items-center gap-1">
                    <Calendar className="h-3 w-3" /> Attestation Time
                </span>
                <p className="font-mono text-sm">{formattedTime}</p>
            </div>
            <div className="space-y-1">
                <span className="text-xs font-medium text-muted-foreground uppercase flex items-center gap-1">
                    <FileKey className="h-3 w-3" /> Algorithm
                </span>
                <p className="font-mono text-sm">{algorithm || "Unknown"}</p>
            </div>
             <div className="space-y-1">
                <span className="text-xs font-medium text-muted-foreground uppercase flex items-center gap-1">
                    <ShieldCheck className="h-3 w-3" /> Status
                </span>
                <p className="font-medium text-sm text-green-600">Valid Signature</p>
            </div>
          </CardContent>
        </Card>
      )}

       {!verified && (
        <Card>
            <CardHeader>
                <CardTitle className="text-sm font-medium text-amber-600 flex items-center gap-2">
                    <ShieldAlert className="h-4 w-4" /> Risk Assessment
                </CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground space-y-2">
                <p>
                    This service does not have a valid cryptographic signature. This means:
                </p>
                <ul className="list-disc pl-5 space-y-1">
                    <li>The author cannot be verified.</li>
                    <li>The code might have been tampered with.</li>
                    <li>It may not be safe to expose sensitive data or credentials to this service.</li>
                </ul>
                <Separator className="my-4" />
                <p className="text-xs">
                    If this is a local development service or internal tool, this warning can be ignored.
                    For production or public services, ensure they are signed by a trusted authority.
                </p>
            </CardContent>
        </Card>
      )}
    </div>
  );
}
