/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect, useState } from "react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ShieldCheck, ShieldAlert, Key, Clock, Fingerprint, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Loader2 } from "lucide-react";

export function AttestationDashboard() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const res = await apiClient.listServices();
        setServices(Array.isArray(res) ? res : res.services || []);
      } catch (err) {
        console.error("Failed to fetch services", err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const filteredServices = services.filter((s) =>
    s.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <span className="ml-2">Verifying provenance...</span>
      </div>
    );
  }

  const verifiedCount = services.filter(s => s.provenance?.verified).length;
  const unverifiedCount = services.length - verifiedCount;

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="bg-gradient-to-br from-green-500/10 to-transparent border-green-500/20">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2 text-green-600">
              <ShieldCheck className="h-4 w-4" /> Verified Services
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{verifiedCount}</div>
            <p className="text-xs text-muted-foreground mt-1">Cryptographically signed & trusted</p>
          </CardContent>
        </Card>

        <Card className="bg-gradient-to-br from-amber-500/10 to-transparent border-amber-500/20">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2 text-amber-600">
              <ShieldAlert className="h-4 w-4" /> Unverified Services
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{unverifiedCount}</div>
            <p className="text-xs text-muted-foreground mt-1">Lacking verifiable provenance</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Fingerprint className="h-4 w-4 text-primary" /> Total Services
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{services.length}</div>
            <p className="text-xs text-muted-foreground mt-1">Active in MCP Any</p>
          </CardContent>
        </Card>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader className="border-b bg-muted/20">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <CardTitle>Attestation Ledger</CardTitle>
              <CardDescription>Review the cryptographic signatures and identity claims of connected servers.</CardDescription>
            </div>
            <div className="relative w-full md:w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search services..."
                className="pl-8 bg-background"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <ScrollArea className="h-[500px] w-full">
            <Table>
              <TableHeader className="bg-muted/50 sticky top-0 z-10 backdrop-blur-md">
                <TableRow>
                  <TableHead className="w-[200px]">Service Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Signer Identity</TableHead>
                  <TableHead>Algorithm</TableHead>
                  <TableHead>Attestation Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredServices.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="h-24 text-center text-muted-foreground">
                      No services found.
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredServices.map((service) => {
                    const prov = service.provenance;
                    const isVerified = prov?.verified;

                    return (
                      <TableRow key={service.id} className="group hover:bg-muted/30 transition-colors">
                        <TableCell className="font-medium truncate max-w-[200px]" title={service.name}>
                          {service.name}
                        </TableCell>
                        <TableCell>
                          {isVerified ? (
                            <Badge variant="outline" className="bg-green-500/10 text-green-600 border-green-500/20 gap-1">
                              <ShieldCheck className="h-3 w-3" /> Verified
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="bg-amber-500/10 text-amber-600 border-amber-500/20 gap-1">
                              <ShieldAlert className="h-3 w-3" /> Unverified
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          {prov?.signerIdentity ? (
                            <div className="flex items-center gap-2 text-sm font-mono text-muted-foreground">
                              <Fingerprint className="h-3 w-3" />
                              <span className="truncate max-w-[150px]" title={prov.signerIdentity}>{prov.signerIdentity}</span>
                            </div>
                          ) : (
                            <span className="text-muted-foreground/50 text-xs italic">Unknown</span>
                          )}
                        </TableCell>
                        <TableCell>
                           {prov?.signatureAlgorithm ? (
                            <div className="flex items-center gap-2 text-sm text-muted-foreground">
                              <Key className="h-3 w-3" />
                              <span>{prov.signatureAlgorithm}</span>
                            </div>
                          ) : (
                            <span className="text-muted-foreground/50 text-xs">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          {prov?.attestationTime ? (
                            <div className="flex items-center gap-2 text-sm text-muted-foreground">
                              <Clock className="h-3 w-3" />
                              <span suppressHydrationWarning>{new Date(prov.attestationTime).toLocaleString()}</span>
                            </div>
                          ) : (
                            <span className="text-muted-foreground/50 text-xs">-</span>
                          )}
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}
