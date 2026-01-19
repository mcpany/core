/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, use } from "react";
import { useRouter } from "next/navigation";
import { marketplaceService, ExternalServer, ExternalMarketplace } from "@/lib/marketplace-service";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { ArrowLeft, Download, Terminal } from "lucide-react";
import Link from "next/link";

/**
 * ExternalMarketplacePage component.
 * @param props - The component props.
 * @param props.params - The params property.
 * @returns The rendered component.
 */
export default function ExternalMarketplacePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const { toast } = useToast();

  const [market, setMarket] = useState<ExternalMarketplace | null>(null);
  const [servers, setServers] = useState<ExternalServer[]>([]);
  const [loading, setLoading] = useState(true);

  // Install State
  const [selectedServer, setSelectedServer] = useState<ExternalServer | null>(null);
  const [isInstalling, setIsInstalling] = useState(false);
  const [envValues, setEnvValues] = useState<Record<string, string>>({});

  useEffect(() => {
    async function load() {
        try {
            const markets = await marketplaceService.fetchPublicMarketplaces();
            const m = markets.find(m => m.id === id);
            setMarket(m || null);

            if (m) {
                const s = await marketplaceService.fetchExternalServers(id);
                setServers(s);
            }
        } catch (e) {
            console.error(e);
            toast({ title: "Failed to load marketplace", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    }
    load();
  }, [id, toast]);

  const handleInstallClick = (server: ExternalServer) => {
      setSelectedServer(server);
      // Pre-populate env vars from config if any
      const initialEnv: Record<string, string> = {};
      if (server.config.commandLineService?.env) {
          Object.keys(server.config.commandLineService.env).forEach(k => {
              initialEnv[k] = "";
          });
      }
      setEnvValues(initialEnv);
  };

  const handleConfirmInstall = async () => {
      if (!selectedServer) return;
      setIsInstalling(true);
      try {
          // Clone config
          const config = { ...selectedServer.config };
          // Update Env
          if (config.commandLineService && config.commandLineService.env) {
              Object.keys(config.commandLineService.env).forEach(key => {
                  if (config.commandLineService?.env?.[key]) {
                      config.commandLineService.env[key].plainText = envValues[key] || "";
                  }
              });
          }

          await apiClient.registerService(config);
          toast({ title: "Service Installed", description: `${selectedServer.name} has been installed.` });
          setSelectedServer(null);
      } catch (e) {
          toast({ title: "Installation Failed", variant: "destructive", description: String(e) });
      } finally {
          setIsInstalling(false);
      }
  };

  if (loading) return <div className="p-8">Loading...</div>;
  if (!market) return <div className="p-8">Marketplace not found</div>;

  return (
    <div className="flex flex-col gap-8 p-8 h-[calc(100vh-4rem)] overflow-y-auto">
        <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => router.back()}>
                <ArrowLeft className="h-4 w-4" />
            </Button>
            <div>
                <h1 className="text-3xl font-bold tracking-tight">{market.name}</h1>
                <p className="text-muted-foreground">{market.description}</p>
            </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {servers.map(server => (
                <Card key={server.id} className="flex flex-col">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                             <Terminal className="h-5 w-5" />
                             {server.name}
                        </CardTitle>
                        <CardDescription>{server.description}</CardDescription>
                    </CardHeader>
                    <CardContent className="flex-1">
                        <div className="text-sm text-muted-foreground">
                            by {server.author || 'Unknown'}
                        </div>
                    </CardContent>
                    <CardFooter>
                        <Button className="w-full" onClick={() => handleInstallClick(server)}>
                            <Download className="mr-2 h-4 w-4" />
                            Import & Install
                        </Button>
                    </CardFooter>
                </Card>
            ))}
             {servers.length === 0 && (
                <div className="col-span-full text-center p-8 text-muted-foreground border rounded-lg border-dashed">
                    No servers found in this marketplace.
                </div>
            )}
        </div>

        <Dialog open={!!selectedServer} onOpenChange={(open) => !open && setSelectedServer(null)}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Install {selectedServer?.name}</DialogTitle>
                    <DialogDescription>
                        Configure environment variables.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    {Object.keys(envValues).map(key => (
                        <div key={key} className="grid grid-cols-4 items-center gap-4">
                            <Label htmlFor={key} className="text-right">{key}</Label>
                            <Input
                                id={key}
                                value={envValues[key]}
                                onChange={e => setEnvValues({...envValues, [key]: e.target.value})}
                                className="col-span-3"
                            />
                        </div>
                    ))}
                     {Object.keys(envValues).length === 0 && (
                        <div className="text-center text-muted-foreground">
                            No configuration needed.
                        </div>
                    )}
                </div>
                <DialogFooter>
                    <Button onClick={handleConfirmInstall} disabled={isInstalling}>
                        {isInstalling ? 'Installing...' : 'Install'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
  );
}
