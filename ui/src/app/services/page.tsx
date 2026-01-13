/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Plus } from "lucide-react";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { useToast } from "@/hooks/use-toast";

import { UpstreamServiceConfig } from "@/lib/client";
import { ServiceList } from "@/components/services/service-list";
import { ServiceEditor } from "@/components/services/editor/service-editor";
import { TemplateSelector } from "@/components/services/template-selector";
import { ServiceTemplate } from "@/data/service-templates";

export default function ServicesPage() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<UpstreamServiceConfig | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isTemplateDialogOpen, setIsTemplateDialogOpen] = useState(false);
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

  const openEdit = useCallback((service: UpstreamServiceConfig) => {
      setSelectedService(service);
      setIsSheetOpen(true);
  }, []);

  const openNew = () => {
      setIsTemplateDialogOpen(true);
  };

  const handleTemplateSelect = (template: ServiceTemplate) => {
      setIsTemplateDialogOpen(false);

      const newService = {
          id: "", // User to fill
          name: "", // User to fill
          version: "1.0.0",
          disable: false,
          priority: 0,
          loadBalancingStrategy: 0,
          ...template.config,
          // Generate a default name/ID if helpful? No, force user to name it.
      } as any;

      setSelectedService(newService);
      setIsSheetOpen(true);
  };

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
             />
        </CardContent>
      </Card>

      {/* Template Selection Dialog */}
      <Dialog open={isTemplateDialogOpen} onOpenChange={setIsTemplateDialogOpen}>
        <DialogContent className="sm:max-w-3xl max-h-[80vh] overflow-y-auto">
             <DialogHeader>
                 <DialogTitle>Choose a Service Template</DialogTitle>
                 <DialogDescription>
                     Select a pre-configured template to get started quickly, or choose 'Custom' to start from scratch.
                 </DialogDescription>
             </DialogHeader>
             <TemplateSelector
                onSelect={handleTemplateSelect}
                onCancel={() => setIsTemplateDialogOpen(false)}
             />
        </DialogContent>
      </Dialog>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent className="sm:max-w-2xl w-full">
            <SheetHeader className="mb-4">
                <SheetTitle>{selectedService?.id ? "Edit Service" : "New Service"}</SheetTitle>
                <SheetDescription>
                    Configure your upstream service details.
                </SheetDescription>
            </SheetHeader>
            {selectedService && (
                <div className="h-[calc(100vh-140px)]">
                    <ServiceEditor
                        service={selectedService}
                        onChange={setSelectedService}
                        onSave={handleSave}
                        onCancel={() => setIsSheetOpen(false)}
                    />
                </div>
            )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
