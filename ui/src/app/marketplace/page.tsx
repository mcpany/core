/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


"use client";

import { useEffect, useState } from "react";
import { marketplaceService, ServiceCollection, ExternalMarketplace, CommunityServer } from "@/lib/marketplace-service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { Download, Package, Globe, ExternalLink, Users, Search } from "lucide-react";
import Link from "next/link";

import { ShareCollectionDialog } from "@/components/share-collection-dialog";
import { CreateConfigWizard } from "@/components/marketplace/wizard/create-config-wizard";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Plus, Trash2 } from "lucide-react";
import { InstantiateDialog } from "@/components/marketplace/instantiate-dialog";
import { CollectionDetailsDialog } from "@/components/marketplace/collection-details-dialog";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { parseGithubUrl } from "@/lib/repo-analyzer";

/**
 * MarketplacePage component.
 * @returns The rendered component.
 */
export default function MarketplacePage() {
  const { toast } = useToast();
  const [collections, setCollections] = useState<ServiceCollection[]>([]);
  const [backendTemplates, setBackendTemplates] = useState<ServiceCollection[]>([]);
  const [publicMarkets, setPublicMarkets] = useState<ExternalMarketplace[]>([]);
  const [communityServers, setCommunityServers] = useState<CommunityServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [importUrl, setImportUrl] = useState("");
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [isWizardOpen, setIsWizardOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  // Instantiate State
  const [isInstantiateOpen, setIsInstantiateOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<UpstreamServiceConfig | undefined>(undefined);
  const [selectedRepoUrl, setSelectedRepoUrl] = useState<string | undefined>(undefined);

  // Collection Details State
  const [selectedCollection, setSelectedCollection] = useState<ServiceCollection | undefined>(undefined);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  // Load Data
  const loadData = async () => {
    try {
        const [cols, markets, templates, community] = await Promise.all([
            marketplaceService.fetchOfficialCollections(),
            marketplaceService.fetchPublicMarketplaces(),
            apiClient.listTemplates().catch(e => {
                console.warn("Failed to list templates", e);
                return [];
            }),
            marketplaceService.fetchCommunityServers().catch(e => {
                console.warn("Failed to fetch community servers", e);
                return [];
            })
        ]);
        setCollections(cols);
        setPublicMarkets(markets);
        setCommunityServers(community);

        // Map backend templates to ServiceCollection format for consistent display
        const mappedTemplates = templates.map((t: UpstreamServiceConfig) => ({
            name: t.name,
            description: "Backend Template",
            author: "User",
            version: t.version || "0.0.1",
            services: [t]
        }));
        setBackendTemplates(mappedTemplates);
    } catch (e) {
        console.error(e);
        toast({ title: "Failed to load marketplace data", variant: "destructive" });
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [toast]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('wizard') === 'open') {
        setIsWizardOpen(true);
    }
  }, []);

  const handleImportSubmit = async () => {
    if (!importUrl) return;

    // Smart Import Logic: Check if it's a valid GitHub URL (not JSON)
    if (parseGithubUrl(importUrl) && !importUrl.endsWith(".json")) {
        setSelectedRepoUrl(importUrl);
        setSelectedTemplate(undefined);
        setIsInstantiateOpen(true);
        setIsImportDialogOpen(false);
        setImportUrl("");
        return;
    }

    // Fallback to Service Collection Import
    try {
        const col = await marketplaceService.importCollection(importUrl);
        marketplaceService.saveLocalCollection(col);
        toast({ title: "Collection Imported", description: `Saved ${col.name} to Local Marketplace` });
        setIsImportDialogOpen(false);
        setImportUrl("");
        loadData();
    } catch (e) {
        toast({ title: "Import Failed", variant: "destructive", description: String(e) });
    }
  };

  const handleWizardComplete = async (config: UpstreamServiceConfig) => {
      try {
          await apiClient.saveTemplate(config);
          toast({ title: "Config Saved", description: `${config.name} saved to Backend Templates.` });
          setIsWizardOpen(false);
          loadData();
      } catch (e) {
          console.error(e);
          toast({ title: "Failed to save template", variant: "destructive", description: String(e) });
      }
  };

  const deleteTemplate = async (templateSvc: UpstreamServiceConfig) => {
      if (!templateSvc.id && !templateSvc.name) return;
      try {
        await apiClient.deleteTemplate(templateSvc.id || templateSvc.name);
        loadData();
        toast({ title: "Deleted", description: "Backend template deleted." });
      } catch (e) {
          toast({ title: "Failed to delete", variant: "destructive", description: String(e) });
      }
  };

  const openInstantiate = (service: UpstreamServiceConfig) => {
      setSelectedTemplate(service);
      setSelectedRepoUrl(undefined);
      setIsInstantiateOpen(true);
  };

  const openInstantiateCommunity = (server: CommunityServer) => {
      // ⚡ Smart Import: Pass the URL to the analyzer instead of hardcoding
      setSelectedRepoUrl(server.url);
      setSelectedTemplate(undefined);
      setIsInstantiateOpen(true);
  };

  const openCollectionDetails = (col: ServiceCollection) => {
      setSelectedCollection(col);
      setIsDetailsOpen(true);
  };

  // Filter Community Servers
  const filteredCommunityServers = communityServers.filter(s =>
      s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.tags.some(t => t.toLowerCase().includes(searchQuery.toLowerCase())) ||
      s.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
            <Button onClick={() => setIsImportDialogOpen(true)} variant="outline">
                <Download className="mr-2 h-4 w-4" />
                Import from URL
            </Button>
            <Button onClick={() => setIsWizardOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Config
            </Button>
        </div>
      </div>

      <Tabs defaultValue="community" className="w-full">
          <TabsList>
              <TabsTrigger value="community">Community</TabsTrigger>
              <TabsTrigger value="official">Official</TabsTrigger>
              <TabsTrigger value="public">Public</TabsTrigger>
              <TabsTrigger value="local">Local</TabsTrigger>
          </TabsList>

          <TabsContent value="official" className="mt-6">
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
                                <Button className="w-full" variant="outline" onClick={() => openCollectionDetails(col)}>
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
          </TabsContent>

          <TabsContent value="public" className="mt-6">
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
          </TabsContent>

          <TabsContent value="local" className="mt-6">
            <section>
                <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                    <Package className="h-5 w-5" />
                    Local Templates
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                     {backendTemplates.map((col, idx) => (
                        <Card key={idx} className="flex flex-col border-dashed">
                            <CardHeader>
                                <CardTitle>{col.name}</CardTitle>
                                <CardDescription>{col.description}</CardDescription>
                            </CardHeader>
                            <CardContent className="flex-1">
                                <div className="text-sm text-muted-foreground">
                                    {col.services.length} Services • {col.version}
                                </div>
                            </CardContent>
                            <CardFooter className="gap-2">
                                <Button className="flex-1" onClick={() => col.services[0] && openInstantiate(col.services[0])}>
                                    Instantiate
                                </Button>
                                <Button variant="destructive" size="icon" onClick={() => col.services[0] && deleteTemplate(col.services[0])}>
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </CardFooter>
                        </Card>
                    ))}
                     {backendTemplates.length === 0 && (
                        <div className="col-span-full text-center p-12 text-muted-foreground border rounded-lg border-dashed bg-muted/20">
                            No local templates. Create one to get started.
                        </div>
                    )}
                </div>
            </section>
          </TabsContent>

          <TabsContent value="community" className="mt-6">
              <section>
                  <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 gap-4">
                      <div>
                        <h2 className="text-xl font-semibold flex items-center gap-2">
                            <Users className="h-5 w-5" />
                            Community Servers (Awesome List)
                        </h2>
                        <p className="text-sm text-muted-foreground mt-1">
                            Discover {communityServers.length} servers from the community.
                        </p>
                      </div>
                      <div className="relative w-full md:w-72">
                          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                          <Input
                              placeholder="Search community servers..."
                              className="pl-8"
                              value={searchQuery}
                              onChange={(e) => setSearchQuery(e.target.value)}
                          />
                      </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                      {filteredCommunityServers.slice(0, 100).map((server, idx) => (
                          <Card key={idx} className="flex flex-col">
                              <CardHeader>
                                  <div className="flex justify-between items-start gap-2">
                                      <CardTitle className="text-lg truncate" title={server.name}>
                                          {server.name}
                                      </CardTitle>
                                      <a href={server.url} target="_blank" rel="noopener noreferrer" className="text-muted-foreground hover:text-primary">
                                          <ExternalLink className="h-4 w-4" />
                                      </a>
                                  </div>
                                  <CardDescription className="line-clamp-2 min-h-[40px] text-xs">
                                      {server.description}
                                  </CardDescription>
                              </CardHeader>
                              <CardContent className="flex-1">
                                  <div className="flex flex-wrap gap-1 mb-2">
                                      <Badge variant="outline" className="text-[10px] h-5">{server.category}</Badge>
                                      {server.tags.map((tag, i) => (
                                          <span key={i} className="text-sm" title="Tag">{tag}</span>
                                      ))}
                                  </div>
                              </CardContent>
                              <CardFooter>
                                  <Button className="w-full" variant="secondary" onClick={() => openInstantiateCommunity(server)}>
                                      Install
                                  </Button>
                              </CardFooter>
                          </Card>
                      ))}
                      {filteredCommunityServers.length === 0 && !loading && (
                          <div className="col-span-full text-center p-12 text-muted-foreground border rounded-lg border-dashed">
                              No servers found matching "{searchQuery}".
                          </div>
                      )}
                  </div>
                  {filteredCommunityServers.length > 100 && (
                      <div className="text-center mt-8 text-muted-foreground text-sm">
                          Showing first 100 of {filteredCommunityServers.length} results. Use search to find more.
                      </div>
                  )}
              </section>
          </TabsContent>
      </Tabs>

      <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
          <DialogContent>
              <DialogHeader>
                  <DialogTitle>Import Service Collection or Repository</DialogTitle>
                  <DialogDescription>
                      Enter the URL of a Service Collection (JSON) or a GitHub Repository to import.
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
                          placeholder="https://github.com/..."
                      />
                  </div>
              </div>
              <DialogFooter>
                  <Button onClick={handleImportSubmit}>Import</Button>
              </DialogFooter>
          </DialogContent>
      </Dialog>

      <CreateConfigWizard
        open={isWizardOpen}
        onOpenChange={setIsWizardOpen}
        onComplete={handleWizardComplete}
      />

      <InstantiateDialog
        open={isInstantiateOpen}
        onOpenChange={setIsInstantiateOpen}
        templateConfig={selectedTemplate}
        onComplete={() => {}}
        repoUrl={selectedRepoUrl}
      />

      <CollectionDetailsDialog
        open={isDetailsOpen}
        onOpenChange={setIsDetailsOpen}
        collection={selectedCollection}
        onInstantiateService={openInstantiate}
      />
    </div>
  );
}
