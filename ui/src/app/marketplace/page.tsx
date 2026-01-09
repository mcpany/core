/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useEffect, useState } from "react";
import { marketplaceService, ServiceCollection, ExternalMarketplace } from "@/lib/marketplace-service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { Download, Package, Globe, ExternalLink } from "lucide-react";
import Link from "next/link";
import { ShareCollectionDialog } from "@/components/share-collection-dialog";

export default function MarketplacePage() {
  const { toast } = useToast();
  const [collections, setCollections] = useState<ServiceCollection[]>([]);
  const [publicMarkets, setPublicMarkets] = useState<ExternalMarketplace[]>([]);
  const [loading, setLoading] = useState(true);
  const [importUrl, setImportUrl] = useState("");
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);

  useEffect(() => {
    async function loadData() {
        try {
            const [cols, markets] = await Promise.all([
                marketplaceService.fetchOfficialCollections(),
                marketplaceService.fetchPublicMarketplaces()
            ]);
            setCollections(cols);
            setPublicMarkets(markets);
        } catch (e) {
            console.error(e);
            toast({ title: "Failed to load marketplace data", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    }
    loadData();
  }, [toast]);

  const handleImportCollection = async () => {
    if (!importUrl) return;
    try {
        const col = await marketplaceService.importCollection(importUrl);
        // In real app, we would install it here or show a confirmation dialog with contents
        toast({ title: "Collection Imported", description: `Would install: ${col.name} (${col.services.length} services)` });
        setIsImportDialogOpen(false);
        setImportUrl("");
    } catch (e) {
        toast({ title: "Import Failed", variant: "destructive", description: String(e) });
    }
  };

  return (
    <div className="flex flex-col gap-8 p-8 h-[calc(100vh-4rem)] overflow-y-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Marketplace</h1>
          <p className="text-muted-foreground mt-2">
            Import Service Collections or browse Public Marketplaces.
          </p>
        </div>
        <div className="flex gap-2">
            <ShareCollectionDialog />
            <Button onClick={() => setIsImportDialogOpen(true)}>
                <Download className="mr-2 h-4 w-4" />
                Import from URL
            </Button>
        </div>
      </div>

      {/* Official Collections */}
      <section>
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <Package className="h-5 w-5" />
              Official Service Collections
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {collections.map((col, idx) => (
                  <Card key={idx} className="flex flex-col">
                      <CardHeader>
                          <CardTitle>{col.name}</CardTitle>
                          <CardDescription>{col.description}</CardDescription>
                      </CardHeader>
                      <CardContent className="flex-1">
                          <div className="text-sm text-muted-foreground">
                              {col.services.length} Services • by {col.author}
                          </div>
                      </CardContent>
                      <CardFooter>
                          <Button className="w-full" variant="outline">
                              View Details
                          </Button>
                      </CardFooter>
                  </Card>
              ))}
              {collections.length === 0 && !loading && (
                  <div className="col-span-full text-center p-8 text-muted-foreground border rounded-lg border-dashed">
                      No official collections found.
                  </div>
              )}
          </div>
      </section>

      {/* Public Marketplaces */}
      <section>
          <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Public MCP Server Marketplaces
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {publicMarkets.map((market) => (
                  <Card key={market.id} className="cursor-pointer hover:border-primary transition-colors">
                      <Link href={`/marketplace/external/${market.id}`}>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                {market.name}
                                <ExternalLink className="h-4 w-4 opacity-50" />
                            </CardTitle>
                            <CardDescription>{market.description}</CardDescription>
                        </CardHeader>
                        <CardContent>
                             <div className="text-sm text-muted-foreground truncate">
                                 {market.url}
                             </div>
                        </CardContent>
                      </Link>
                  </Card>
              ))}
          </div>
      </section>

      <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
          <DialogContent>
              <DialogHeader>
                  <DialogTitle>Import Service Collection</DialogTitle>
                  <DialogDescription>
                      Enter the URL of a Service Collection (JSON) to import.
                  </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                  <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="url" className="text-right">
                          URL
                      </Label>
                      <Input
                          id="url"
                          value={importUrl}
                          onChange={(e) => setImportUrl(e.target.value)}
                          className="col-span-3"
                          placeholder="https://..."
                      />
                  </div>
              </div>
              <DialogFooter>
                  <Button onClick={handleImportCollection}>Import</Button>
              </DialogFooter>
          </DialogContent>
      </Dialog>
    </div>
  );
}
