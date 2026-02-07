/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus, ChevronLeft } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"
import { useToast } from "@/hooks/use-toast";

import { UpstreamServiceConfig } from "@/lib/client";
import { ServiceList } from "@/components/services/service-list";
import { ServiceEditor } from "@/components/services/editor/service-editor";
import { ServiceTemplateSelector } from "@/components/services/service-template-selector";
import { ServiceTemplate } from "@/lib/templates";
import { BulkServiceImport } from "@/components/services/bulk-service-import";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Download } from "lucide-react";
import { TemplateConfigForm } from "@/components/services/template-config-form";
import { applyTemplateFields } from "@/lib/template-utils";


/**
 * ServicesPage component.
 * @returns The rendered component.
 */
export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [configuringTemplate, setConfiguringTemplate] = useState<ServiceTemplate | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchServices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.listServices();
      if (!res) {
          setServices([]);
          return;
      }
      // Handle both array and object response formats for robustness
      if (Array.isArray(res)) {
          setServices(res);
      } else {
          setServices(res.services || []);
      }
    } catch (e) {
      console.error("Failed to fetch services", e);
      toast({
          variant: "destructive",
          title: "Error",
          description: "Failed to load services."
      });
    } finally {
        setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  const toggleService = useCallback(async (name: string, enabled: boolean) => {
    // Optimistic update
    setServices(prev => prev.map(s => s.name === name ? { ...s, disable: !enabled } : s));

    try {
        await apiClient.setServiceStatus(name, !enabled);
        toast({
            title: enabled ? "Service Enabled" : "Service Disabled",
            description: `Service ${name} has been ${enabled ? "enabled" : "disabled"}.`
        });
    } catch (e) {
        console.error("Failed to toggle service", e);
        fetchServices(); // Revert
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update service status."
        });
    }
  }, [fetchServices, toast]);

  const deleteService = useCallback(async (name: string) => {
    if (!confirm(`Are you sure you want to delete service "${name}"?`)) return;

    try {
        await apiClient.unregisterService(name);
        setServices(prev => prev.filter(s => s.name !== name));
        toast({
            title: "Service Deleted",
            description: `Service ${name} has been removed.`
        });
    } catch (e) {
         console.error("Failed to delete service", e);
         fetchServices();
         toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to delete service."
        });
    }
  }, [fetchServices, toast]);

  const bulkToggleService = useCallback(async (names: string[], enabled: boolean) => {
    // Optimistic update
    setServices(prev => prev.map(s => names.includes(s.name) ? { ...s, disable: !enabled } : s));

    try {
        await Promise.all(names.map(name => apiClient.setServiceStatus(name, !enabled)));
        toast({
            title: enabled ? "Services Enabled" : "Services Disabled",
            description: `${names.length} services have been ${enabled ? "enabled" : "disabled"}.`
        });
    } catch (e) {
        console.error("Failed to bulk toggle services", e);
        fetchServices(); // Revert
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update some services."
        });
    }
  }, [fetchServices, toast]);

  const bulkDeleteService = useCallback(async (names: string[]) => {
    if (!confirm(`Are you sure you want to delete ${names.length} services?`)) return;

    try {
        await Promise.all(names.map(name => apiClient.unregisterService(name)));
        setServices(prev => prev.filter(s => !names.includes(s.name)));
        toast({
            title: "Services Deleted",
            description: `${names.length} services have been removed.`
        });
    } catch (e) {
         console.error("Failed to delete services", e);
         fetchServices();
         toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to delete some services."
        });
    }
  }, [fetchServices, toast]);

  const handleBulkEdit = useCallback(async (names: string[], updates: { tags?: string[], resilience?: { timeout?: string, maxRetries?: number } }) => {
    try {
        const servicesToUpdate = services.filter(s => names.includes(s.name));

        // Execute sequentially to avoid SQLITE_BUSY errors on backend
        for (const service of servicesToUpdate) {
            const updated = { ...service };
            if (updates.tags) {
                updated.tags = [...new Set([...(service.tags || []), ...updates.tags])];
            }
            if (updates.resilience) {
                const currentResilience: any = service.resilience || {};
                const newResilience = { ...currentResilience };

                if (updates.resilience.timeout) {
                    newResilience.timeout = updates.resilience.timeout;
                }
                if (updates.resilience.maxRetries !== undefined) {
                    newResilience.retryPolicy = {
                        ...(currentResilience.retryPolicy || {}),
                        numberOfRetries: updates.resilience.maxRetries
                    };
                }
                updated.resilience = newResilience;
            }
            await apiClient.updateService(updated as any);
        }

        toast({
            title: "Services Updated",
            description: `${names.length} services have been updated.`
        });
        fetchServices();
    } catch (e) {
        console.error("Failed to bulk edit services", e);
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update some services."
        });
    }
  }, [services, fetchServices, toast]);

  const openEdit = useCallback((service: UpstreamServiceConfig) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  }, []);

  const openNew = () => {
      setSelectedService(null);
      setConfiguringTemplate(null);
      setIsSheetOpen(true);
  };

  const initServiceFromConfig = (config: Partial<UpstreamServiceConfig>) => {
      // Deep copy config to avoid mutating template
      const newService = JSON.parse(JSON.stringify(config));
      // Ensure defaults
      newService.version = newService.version || "1.0.0";
      newService.priority = newService.priority || 0;
      newService.disable = false;
      // Ensure ID is empty to mark as new
      newService.id = "";

      return newService;
  }

  const handleTemplateSelect = (template: ServiceTemplate) => {
      if (template.fields && template.fields.length > 0) {
          setConfiguringTemplate(template);
      } else {
          const newService = initServiceFromConfig(template.config);
          setSelectedService(newService);
      }
  };

  const handleTemplateConfigSubmit = (values: Record<string, string>) => {
      if (!configuringTemplate) return;

      const configuredConfig = applyTemplateFields(configuringTemplate, values);
      const newService = initServiceFromConfig(configuredConfig);

      setSelectedService(newService);
      setConfiguringTemplate(null); // Clear the configuration step
  };

  const handleDuplicate = useCallback((service: UpstreamServiceConfig) => {
      // Deep clone
      const newService = JSON.parse(JSON.stringify(service));
      // Reset ID to ensure it creates a new service
      newService.id = "";
      // Append copy to name to avoid collision
      newService.name = `${newService.name}-copy`;
      // Ensure clean state
      delete newService.lastError;
      delete newService.connectionPool; // Runtime status

      setSelectedService(newService);
      setIsSheetOpen(true);
      toast({
        title: "Service Duplicated",
        description: "A copy of the service has been created. Please review and save.",
      });
  }, [toast]);

  const handleExport = useCallback((service: UpstreamServiceConfig) => {
      // Create a clean copy for export (remove runtime fields like lastError)
      const exportData = JSON.parse(JSON.stringify(service));
      delete exportData.lastError;
      // Remove runtime status fields if any
      delete exportData.connectionPool;
      // Keep hooks as they are part of the configuration
      // delete exportData.preCallHooks;
      // delete exportData.postCallHooks;

      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${service.name}-config.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      toast({
        title: "Service Exported",
        description: `Configuration for ${service.name} has been downloaded.`,
      });
  }, [toast]);

  const handleLogin = useCallback(async (service: UpstreamServiceConfig) => {
      try {
          // We use name as ID if ID is missing (common in config files)
          const serviceId = service.id || service.name;
          // Include service_id in redirect URL so we can retrieve it in the callback
          const redirectUrl = `${window.location.origin}/oauth/callback?service_id=${encodeURIComponent(serviceId)}`;

          const res = await apiClient.initiateOAuth(serviceId, redirectUrl);
          if (res.authorization_url && res.state) {
              // Store context for callback verification using unified JSON pattern
              sessionStorage.setItem(`oauth_pending_${res.state}`, JSON.stringify({
                  serviceId: serviceId,
                  credentialId: '',
                  state: res.state,
                  redirectUrl: redirectUrl,
                  returnPath: window.location.pathname + window.location.search
              }));

              window.location.href = res.authorization_url;
          }
      } catch (e) {
          console.error("Failed to initiate OAuth", e);
          toast({
              variant: "destructive",
              title: "Login Failed",
              description: "Could not initiate OAuth flow."
          });
      }
  }, [toast]);

  const handleRestart = useCallback(async (name: string) => {
    try {
        await apiClient.restartService(name);
        toast({
            title: "Service Restarted",
            description: `Service ${name} has been restarted.`
        });
        fetchServices();
    } catch (e) {
        console.error("Failed to restart service", e);
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to restart service."
        });
    }
  }, [fetchServices, toast]);

  const handleSave = async () => {
      if (!selectedService) return;

      try {
          if (selectedService.id) {
               // Update
               await apiClient.updateService(selectedService);
               toast({ title: "Service Updated", description: "Service configuration saved." });
          } else {
              // Create
              await apiClient.registerService(selectedService);
              toast({ title: "Service Created", description: "New service registered successfully." });
          }
          setIsSheetOpen(false);
          fetchServices();
      } catch (err) {
          console.error("Failed to save service", err);
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to save service configuration."
          });
      }
  };

  const renderSheetContent = () => {
      if (selectedService) {
          return (
            <div className="h-[calc(100vh-140px)]">
                <ServiceEditor
                    service={selectedService}
                    onChange={setSelectedService}
                    onSave={handleSave}
                    onCancel={() => {
                        if (!selectedService.id) {
                            setSelectedService(null);
                        } else {
                            setIsSheetOpen(false);
                        }
                    }}
                />
            </div>
          );
      }

      if (configuringTemplate) {
          return (
              <div className="h-[calc(100vh-140px)] overflow-y-auto">
                  <TemplateConfigForm
                      template={configuringTemplate}
                      onCancel={() => setConfiguringTemplate(null)}
                      onSubmit={handleTemplateConfigSubmit}
                  />
              </div>
          );
      }

      return (
        <div className="h-[calc(100vh-140px)] overflow-y-auto">
           <ServiceTemplateSelector onSelect={handleTemplateSelect} />
        </div>
      );
  };

  const getSheetTitle = () => {
      if (selectedService?.id) return "Edit Service";
      if (selectedService) return "New Service"; // After template selection/config
      if (configuringTemplate) return `Configure ${configuringTemplate.name}`;
      return "New Service"; // Template selection
  };

  const getSheetDescription = () => {
      if (selectedService) return "Configure your upstream service details.";
      if (configuringTemplate) return "Enter the required information to set up this service.";
      return "Choose a template to start quickly.";
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Upstream Services</h1>
        <div className="flex items-center gap-2">
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant="outline">
                        <Download className="mr-2 h-4 w-4" /> Bulk Import
                    </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-xl">
                    <DialogHeader>
                        <DialogTitle>Bulk Service Import</DialogTitle>
                        <DialogDescription>
                            Import multiple services at once from a JSON configuration.
                        </DialogDescription>
                    </DialogHeader>
                    <BulkServiceImport
                        onImportSuccess={() => fetchServices()}
                        onCancel={() => {}}
                    />
                </DialogContent>
            </Dialog>
            <Button onClick={openNew}>
                <Plus className="mr-2 h-4 w-4" /> Add Service
            </Button>
        </div>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Upstream Services</CardTitle>
          <CardDescription>Manage your connected upstream services.</CardDescription>
        </CardHeader>
        <CardContent>
             <ServiceList
                services={services}
                isLoading={loading}
                onToggle={toggleService}
                onEdit={openEdit}
                onDelete={deleteService}
                onDuplicate={handleDuplicate}
                onExport={handleExport}
                onBulkToggle={bulkToggleService}
                onBulkDelete={bulkDeleteService}
                onBulkEdit={handleBulkEdit}
                onLogin={handleLogin}
                onRestart={handleRestart}
             />
        </CardContent>
      </Card>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent className="sm:max-w-2xl w-full">
            <SheetHeader className="mb-4">
                <div className="flex items-center gap-2">
                     {(selectedService && !selectedService.id) || configuringTemplate ? (
                         <Button variant="ghost" size="icon" className="-ml-2 h-8 w-8" onClick={() => {
                             if (selectedService) setSelectedService(null);
                             if (configuringTemplate) setConfiguringTemplate(null);
                         }}>
                             <ChevronLeft className="h-4 w-4" />
                         </Button>
                     ) : null}
                     <SheetTitle>{getSheetTitle()}</SheetTitle>
                </div>
                <SheetDescription>
                    {getSheetDescription()}
                </SheetDescription>
            </SheetHeader>
            {renderSheetContent()}
        </SheetContent>
      </Sheet>
    </div>
  );
}
