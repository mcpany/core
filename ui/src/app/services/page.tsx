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

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchServices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.listServices();
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

  const openEdit = useCallback((service: UpstreamServiceConfig) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  }, []);

  const openNew = () => {
      setSelectedService(null);
      setIsSheetOpen(true);
  };

  const handleTemplateSelect = (template: ServiceTemplate) => {
      // Deep copy config to avoid mutating template
      const newService = JSON.parse(JSON.stringify(template.config));
      // Ensure defaults
      newService.version = newService.version || "1.0.0";
      newService.priority = newService.priority || 0;
      newService.disable = false;
      // Ensure ID is empty to mark as new
      newService.id = "";

      setSelectedService(newService);
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
              // Store context for callback verification using unified keys
              sessionStorage.setItem('oauth_service_id', serviceId);
              sessionStorage.setItem('oauth_state', res.state);
              sessionStorage.setItem('oauth_redirect_url', redirectUrl);
              sessionStorage.setItem('oauth_return_path', window.location.pathname + window.location.search);

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

  const handleSave = async () => {
      if (!selectedService) return;

      try {
          if (selectedService.id) {
               // Update
               await apiClient.updateService(selectedService as any);
               toast({ title: "Service Updated", description: "Service configuration saved." });
          } else {
              // Create
              await apiClient.registerService(selectedService as any);
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

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Services</h2>
        <Button onClick={openNew}>
            <Plus className="mr-2 h-4 w-4" /> Add Service
        </Button>
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
                onLogin={handleLogin}
             />
        </CardContent>
      </Card>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent className="sm:max-w-2xl w-full">
            <SheetHeader className="mb-4">
                <div className="flex items-center gap-2">
                     {selectedService && !selectedService.id && (
                         <Button variant="ghost" size="icon" className="-ml-2 h-8 w-8" onClick={() => setSelectedService(null)}>
                             <ChevronLeft className="h-4 w-4" />
                         </Button>
                     )}
                     <SheetTitle>{selectedService?.id ? "Edit Service" : "New Service"}</SheetTitle>
                </div>
                <SheetDescription>
                    {selectedService ? "Configure your upstream service details." : "Choose a template to start quickly."}
                </SheetDescription>
            </SheetHeader>
            {!selectedService ? (
                 <div className="h-[calc(100vh-140px)] overflow-y-auto">
                    <ServiceTemplateSelector onSelect={handleTemplateSelect} />
                 </div>
            ) : (
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
            )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
