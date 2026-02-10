/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { WizardService } from "../wizard-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

interface ProfileStepProps {
  services: WizardService[];
  onBack: () => void;
  onComplete: (profileName: string) => void;
}

/**
 * Final step for creating the profile.
 *
 * @param props - The component props.
 * @param props.services - The list of configured and authenticated services.
 * @param props.onBack - Callback to go back to the previous step.
 * @param props.onComplete - Callback invoked when the profile is created.
 * @returns The rendered ProfileStep component.
 */
export function ProfileStep({ services, onBack, onComplete }: ProfileStepProps) {
  const [profileName, setProfileName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleCreate = async () => {
    if (!profileName.trim()) {
      toast.error("Please enter a profile name");
      return;
    }

    setSubmitting(true);
    try {
      // 1. Ensure all services are updated/registered with latest config (names, auth tokens)
      for (const svc of services) {
        try {
          await apiClient.getService(svc.instanceName);
          await apiClient.updateService({ ...svc.config, id: svc.instanceName, name: svc.instanceName });
        } catch {
          await apiClient.registerService({ ...svc.config, id: svc.instanceName, name: svc.instanceName });
        }
      }

      // 2. Create Profile
      const serviceConfig: Record<string, any> = {};
      services.forEach(s => {
        serviceConfig[s.instanceName] = { enabled: true };
      });

      const profileData = {
        name: profileName,
        selector: {
          tags: ["dev"], // Default to dev for wizard
          toolProperties: {}
        },
        serviceConfig,
        secrets: {},
        requiredRoles: [],
        parentProfileIds: []
      };

      await apiClient.createProfile(profileData);
      onComplete(profileName);

    } catch (error: any) {
      console.error(error);
      toast.error(`Failed to create profile: ${error.message}`);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-6 max-w-md mx-auto py-8">
      <div className="space-y-4">
        <div className="space-y-2">
          <h3 className="text-lg font-medium">Finalize Profile</h3>
          <p className="text-sm text-muted-foreground">
            Give your new profile a name. It will be created with {services.length} connected services.
          </p>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="pName">Profile Name</Label>
          <Input
            id="pName"
            value={profileName}
            onChange={(e) => setProfileName(e.target.value)}
            placeholder="e.g. personal-dev"
          />
        </div>

        <div className="bg-muted p-4 rounded-md text-sm">
          <strong>Included Services:</strong>
          <ul className="list-disc list-inside mt-2">
            {services.map((s, i) => (
              <li key={i}>{s.instanceName} <span className="text-muted-foreground">({s.templateId})</span></li>
            ))}
          </ul>
        </div>
      </div>

      <div className="flex justify-between pt-4">
        <Button variant="outline" onClick={onBack} disabled={submitting}>Back</Button>
        <Button onClick={handleCreate} disabled={submitting}>
          {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create Profile
        </Button>
      </div>
    </div>
  );
}
