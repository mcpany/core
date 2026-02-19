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
import { Download, Package, Globe, ExternalLink, Users, Search, ShieldCheck } from "lucide-react";
import Link from "next/link";

import { ShareCollectionDialog } from "@/components/share-collection-dialog";
import { CreateConfigWizard } from "@/components/marketplace/wizard/create-config-wizard";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Plus, Trash2 } from "lucide-react";
import { InstantiateDialog } from "@/components/marketplace/instantiate-dialog";
import { CollectionDetailsDialog } from "@/components/marketplace/collection-details-dialog";
import { apiClient, UpstreamServiceConfig, ServiceTemplate } from "@/lib/client";
import { Badge } from "@/components/ui/badge";
import { SERVICE_REGISTRY } from "@/lib/service-registry";

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
    const [popularServices, setPopularServices] = useState<UpstreamServiceConfig[]>([]);

  // Instantiate State
  const [isInstantiateOpen, setIsInstantiateOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<UpstreamServiceConfig | undefined>(undefined);

  // Collection Details State
  const [selectedCollection, setSelectedCollection] = useState<ServiceCollection | undefined>(undefined);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  // Load Data
  const loadData = async () => {
    try {
        const [cols, markets, templates, community, catalogServices] = await Promise.all([
            marketplaceService.fetchOfficialCollections(),
            marketplaceService.fetchPublicMarketplaces(),
            apiClient.listTemplates().catch(e => {
                console.warn("Failed to list templates", e);
                return [];
            }),
            marketplaceService.fetchCommunityServers().catch(e => {
                console.warn("Failed to fetch community servers", e);
                return [];
            }),
            apiClient.listCatalog().catch(e => {
                console.warn("Failed to list catalog", e);
                return [];
            })
        ]);
        setCollections(cols);
        setPublicMarkets(markets);
        setCommunityServers(community);
        // Use catalog services for "Popular" tab
        // We need to map them to the UI format if needed, but UpstreamServiceConfig is compatible?
        // POPULAR_SERVICES was UpstreamServiceConfig[].
        setPopularServices(catalogServices);

        // Map backend templates to ServiceCollection format for consistent display
        const mappedTemplates = templates.map((t: ServiceTemplate) => ({
            name: t.name,
            description: t.description || "Backend Template",
            author: "User",
            version: t.serviceConfig?.version || "0.0.1",
            services: [
                {
                    ...t.serviceConfig,
                    templateId: t.id // Inject template ID for deletion
                }
            ]
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

  const handleImportCollection = async () => {
    if (!importUrl) return;
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
          // Context: Wizard returns UpstreamServiceConfig
          // We need to wrap it in ServiceTemplate to save
          const template: ServiceTemplate = {
              id: "", // Server generates ID
              name: config.name,
              description: config.description || "Created via Wizard",
              icon: "package",
              tags: config.tags || [],
              serviceConfig: config
          };
          await apiClient.saveTemplate(template);
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
          // Use templateId if available (injected during mapping), otherwise fallback to name?
          // Actually for locally created ones via wizard, we might not have templateId if we didn't reload?
          // But loadData is called after save.
          const idToDelete = templateSvc.templateId || templateSvc.id || templateSvc.name;
          await apiClient.deleteTemplate(idToDelete);
        loadData();
        toast({ title: "Deleted", description: "Backend template deleted." });
      } catch (e) {
          toast({ title: "Failed to delete", variant: "destructive", description: String(e) });
      }
  };

  const openInstantiate = (service: UpstreamServiceConfig) => {
      setSelectedTemplate(service);
      setIsInstantiateOpen(true);
  };

  const openInstantiateCommunity = (server: CommunityServer) => {
      // Check Registry First
      const registryMatch = SERVICE_REGISTRY.find(item => {
          // Check by name exact match
          if (item.name.toLowerCase() === server.name.toLowerCase()) return true;
          // Check by repo URL substring match
          if (server.url.includes(item.repo)) return true;
          return false;
      });

      let command = "";
      let configurationSchema = "";

      if (registryMatch) {
          command = registryMatch.command;
          configurationSchema = JSON.stringify(registryMatch.configurationSchema);
      } else {
          // Fallback to best-effort heuristic
          const isPython = server.tags.some(t => t.includes('🐍'));

          // Try to extract repo name for npx
          const repoMatch = server.url.match(/github\.com\/([^/]+)\/([^/]+)/);

          if (repoMatch) {
              const owner = repoMatch[1];
              const repo = repoMatch[2];
              if (isPython) {
                 command = `uvx ${repo}`;
              } else {
                 if (owner === 'modelcontextprotocol' && repo.startsWith('server-')) {
                     command = `npx -y @modelcontextprotocol/${repo}`;
                 } else {
                     command = `npx -y ${repo}`;
                 }
              }
          } else {
              command = isPython ? "uvx package-name" : "npx -y package-name";
          }
      }

      const config: UpstreamServiceConfig = {
          id: server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
          name: server.name,
          configurationSchema: configurationSchema,
          version: "1.0.0",
          commandLineService: {
              command: command,
              env: {},
              workingDirectory: "",
              tools: [],
              resources: [],
              prompts: [],
              calls: {},
              communicationProtocol: 0,
              local: false
          },
          disable: false,
          sanitizedName: server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-'),
          priority: 0,
          loadBalancingStrategy: 0,
          callPolicies: [],
          preCallHooks: [],
          postCallHooks: [],
          prompts: [],
          autoDiscoverTool: true,
          configError: "",
          tags: server.tags,
          readOnly: false
      };

      setSelectedTemplate(config);
      setIsInstantiateOpen(true);
  };

  const openCollectionDetails = (col: ServiceCollection) => {
      setSelectedCollection(col);
      setIsDetailsOpen(true);
  };

  const isVerified = (server: CommunityServer) => {
       return SERVICE_REGISTRY.some(item => {
          if (item.name.toLowerCase() === server.name.toLowerCase()) return true;
          if (server.url.includes(item.repo)) return true;
          return false;
      });
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
                  <TabsTrigger value="popular">Popular</TabsTrigger>
              <TabsTrigger value="community">Community</TabsTrigger>
              <TabsTrigger value="official">Official</TabsTrigger>
              <TabsTrigger value="public">Public</TabsTrigger>
              <TabsTrigger value="local">Local</TabsTrigger>
          </TabsList>

              <TabsContent value="popular" className="mt-6">
                  <section>
                      <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                          <Package className="h-5 w-5" />
                          Popular Services
                      </h2>
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                          {popularServices.map((service) => (
                              <Card key={service.id || service.name} className="flex flex-col">
                                  <CardHeader>
                                      <CardTitle>{service.name}</CardTitle>
                                      <CardDescription>{service.description || "No description"}</CardDescription>
                                  </CardHeader>
                                  <CardContent className="flex-1">
                                      <code className="text-xs bg-muted p-1 rounded break-all">
                                          {service.commandLineService?.command || "Remote Service"}
                                      </code>
                                  </CardContent>
                                  <CardFooter>
                                      <Button className="w-full" onClick={() => openInstantiate(service)}>
                                          Configure
                                      </Button>
                                  </CardFooter>
                              </Card>
                          ))}
                      </div>
                  </section>
              </TabsContent>

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
                                      <CardTitle className="text-lg truncate flex items-center gap-2" title={server.name}>
                                          {server.name}
                                          {isVerified(server) && (
                                              <ShieldCheck className="h-4 w-4 text-blue-500" aria-label="Verified Config" />
                                          )}
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
