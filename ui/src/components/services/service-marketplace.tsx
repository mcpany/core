/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import {
    Search,
    Download,
    FolderOpen,
    Database,
    Github,
    DatabaseZap,
    Cloud,
    MessageSquare,
    BrainCircuit,
    ListOrdered,
    Globe,
    Box,
    ExternalLink
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { MARKETPLACE_ITEMS, MarketplaceItem } from "@/lib/marketplace-data";
import { apiClient } from "@/lib/client";

// Icon mapping
const ICON_MAP: Record<string, React.ReactNode> = {
    FolderOpen: <FolderOpen className="h-8 w-8 text-blue-500" />,
    Database: <Database className="h-8 w-8 text-emerald-500" />,
    Github: <Github className="h-8 w-8 text-slate-800 dark:text-slate-100" />,
    DatabaseZap: <DatabaseZap className="h-8 w-8 text-blue-600" />,
    Cloud: <Cloud className="h-8 w-8 text-sky-500" />,
    MessageSquare: <MessageSquare className="h-8 w-8 text-pink-500" />,
    BrainCircuit: <BrainCircuit className="h-8 w-8 text-purple-500" />,
    ListOrdered: <ListOrdered className="h-8 w-8 text-orange-500" />,
    Globe: <Globe className="h-8 w-8 text-cyan-500" />,
    Search: <Search className="h-8 w-8 text-red-500" />
};

interface ServiceMarketplaceProps {
    onInstallComplete?: () => void;
}

export function ServiceMarketplace({ onInstallComplete }: ServiceMarketplaceProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedItem, setSelectedItem] = useState<MarketplaceItem | null>(null);
    const [envValues, setEnvValues] = useState<Record<string, string>>({});
    const [isInstalling, setIsInstalling] = useState(false);
    const [installDialogOpen, setInstallDialogOpen] = useState(false);

    const filteredItems = MARKETPLACE_ITEMS.filter(item =>
        item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.category.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const handleInstallClick = (item: MarketplaceItem) => {
        setSelectedItem(item);
        setEnvValues({});
        // Initialize required envs with empty strings
        const initialEnvs: Record<string, string> = {};
        item.config.envVars.forEach(v => initialEnvs[v.name] = "");
        setEnvValues(initialEnvs);
        setInstallDialogOpen(true);
    };

    const handleConfirmInstall = async () => {
        if (!selectedItem) return;

        // Validation
        const missing = selectedItem.config.envVars.filter(v => v.required && !envValues[v.name]);
        if (missing.length > 0) {
            toast.error(`Please provide values for: ${missing.map(v => v.name).join(", ")}`);
            return;
        }

        setIsInstalling(true);
        try {
            // Construct command with substituted variables
            let finalArgs = [...selectedItem.config.args];
            const finalEnv: Record<string, string> = {};

            // 1. Substitute variables in ARGS (e.g. ${DB_PATH})
            // 2. Put variables that aren't in args into ENV
            selectedItem.config.envVars.forEach(v => {
                const val = envValues[v.name];
                let usedInArgs = false;

                finalArgs = finalArgs.map(arg => {
                    if (arg === `\${${v.name}}`) {
                        usedInArgs = true;
                        return val;
                    }
                    return arg;
                });

                if (!usedInArgs) {
                    finalEnv[v.name] = val;
                }
            });

            // Ensure all defined variables are also present in the environment map
            // This ensures tools that read from env (or both args and env) work correctly.
            selectedItem.config.envVars.forEach(v => {
                 finalEnv[v.name] = envValues[v.name];
            });

            // Construct Service Config
            const serviceConfig = {
                name: selectedItem.id + "-" + Math.random().toString(36).substr(2, 4), // simple unique suffix
                command_line_service: {
                    command: selectedItem.config.command,
                    args: finalArgs,
                    env: finalEnv
                },
                disable: false
            };

            await apiClient.registerService(serviceConfig);
            toast.success(`Successfully installed ${selectedItem.name}`);
            setInstallDialogOpen(false);
            if (onInstallComplete) onInstallComplete();
        } catch (error) {
            console.error("Install failed", error);
            toast.error("Failed to install service. Check console for details.");
        } finally {
            setIsInstalling(false);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
                <div className="relative w-full md:w-96">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search services..."
                        className="pl-9 bg-background/50"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
                <div className="flex gap-2">
                    {/* Category filters could go here */}
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {filteredItems.map((item) => (
                    <Card key={item.id} className="hover:shadow-md transition-shadow group flex flex-col">
                        <CardHeader className="pb-3">
                            <div className="flex justify-between items-start">
                                <div className="p-2 bg-muted/30 rounded-lg group-hover:bg-primary/5 transition-colors">
                                    {ICON_MAP[item.icon] || <Box className="h-8 w-8 text-gray-500" />}
                                </div>
                                <Badge variant="secondary" className="text-[10px] capitalize">
                                    {item.category}
                                </Badge>
                            </div>
                            <CardTitle className="mt-4 text-lg">{item.name}</CardTitle>
                            <CardDescription className="line-clamp-2 h-10 text-xs mt-1">
                                {item.description}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="flex-1">
                            <div className="text-xs text-muted-foreground flex items-center gap-1">
                                <span className="opacity-70">By</span>
                                <span className="font-medium opacity-90">{item.author}</span>
                            </div>
                        </CardContent>
                        <CardFooter className="pt-2 border-t bg-muted/10 p-4">
                            <Button className="w-full gap-2" variant="default" onClick={() => handleInstallClick(item)}>
                                <Download className="h-4 w-4" /> Install
                            </Button>
                        </CardFooter>
                    </Card>
                ))}
                {filteredItems.length === 0 && (
                     <div className="col-span-full py-12 text-center text-muted-foreground">
                        No services found matching "{searchQuery}".
                    </div>
                )}
            </div>

            <Dialog open={installDialogOpen} onOpenChange={setInstallDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                             Installing {selectedItem?.name}
                        </DialogTitle>
                        <DialogDescription>
                            Configure the required settings to deploy this service.
                        </DialogDescription>
                    </DialogHeader>

                    {selectedItem && (
                        <div className="py-4 space-y-4">
                             <div className="p-3 bg-muted/30 rounded-md text-xs font-mono break-all border">
                                <span className="text-muted-foreground select-none">$ </span>
                                {selectedItem.config.command} {selectedItem.config.args.join(" ")}
                             </div>

                             {selectedItem.config.envVars.length > 0 && (
                                <div className="space-y-4">
                                    <h4 className="text-sm font-medium">Configuration</h4>
                                    {selectedItem.config.envVars.map((env) => (
                                        <div key={env.name} className="grid gap-1.5">
                                            <Label htmlFor={env.name} className="text-xs font-semibold flex items-center gap-1">
                                                {env.name} {env.required && <span className="text-red-500">*</span>}
                                            </Label>
                                            <Input
                                                id={env.name}
                                                type={env.type === 'password' ? 'password' : 'text'}
                                                placeholder={env.placeholder || `Enter ${env.name}`}
                                                value={envValues[env.name] || ""}
                                                onChange={(e) => setEnvValues({...envValues, [env.name]: e.target.value})}
                                                className="h-8 text-sm"
                                            />
                                            <p className="text-[10px] text-muted-foreground">{env.description}</p>
                                        </div>
                                    ))}
                                </div>
                             )}

                             {selectedItem.config.envVars.length === 0 && (
                                 <div className="text-sm text-muted-foreground italic text-center py-2">
                                     No additional configuration required.
                                 </div>
                             )}
                        </div>
                    )}

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setInstallDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleConfirmInstall} disabled={isInstalling}>
                            {isInstalling ? "Installing..." : "Install Service"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
