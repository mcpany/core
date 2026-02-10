/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { WizardService } from "../wizard-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { apiClient } from "@/lib/client";

interface ServiceConfigStepProps {
    services: WizardService[];
    onNext: (services: WizardService[]) => void;
    onBack: () => void;
}

/**
 * ServiceConfigStep allows users to configure instance names for selected services.
 *
 * @param props.services The list of services to configure.
 * @param props.onNext Callback to proceed to the next step.
 * @param props.onBack Callback to go back to the previous step.
 */
export function ServiceConfigStep({ services, onNext, onBack }: ServiceConfigStepProps) {
    // Local state for edits
    const [configs, setConfigs] = useState<WizardService[]>(
        services.map(s => ({
            ...s,
            instanceName: s.instanceName || s.config.name + "-" + Math.random().toString(36).substring(7)
        }))
    );
    const [checkStatus, setCheckStatus] = useState<Record<string, 'checking' | 'ok' | 'fail'>>({});

    const updateName = (idx: number, name: string) => {
        const next = [...configs];
        next[idx].instanceName = name;
        next[idx].config.name = name; // Update inner config name too
        setConfigs(next);
        // Reset status check
        setCheckStatus(prev => ({ ...prev, [idx]: undefined }));
    };

    const handleNext = async () => {
        // Validate names (check duplicates against backend?)
        // For now, just check non-empty
        if (configs.some(c => !c.instanceName.trim())) {
            toast.error("All services must have a name");
            return;
        }

        // Ideally we check if service ID exists, but let's assume register will fail if so?
        // Or we can pre-check.
        // Let's just proceed.
        onNext(configs);
    };

    return (
        <div className="space-y-6">
            <div className="space-y-4">
                {configs.map((svc, idx) => (
                    <Card key={svc.templateId + idx}>
                        <CardHeader className="py-3">
                            <CardTitle className="text-base">Configure {svc.templateId}</CardTitle>
                        </CardHeader>
                        <CardContent className="py-3">
                            <div className="grid gap-2">
                                <Label>Service Instance Name</Label>
                                <Input
                                    value={svc.instanceName}
                                    onChange={(e) => updateName(idx, e.target.value)}
                                    placeholder="e.g. my-google-calendar"
                                />
                                <p className="text-xs text-muted-foreground">
                                    Unique identifier for this service instance.
                                </p>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            <div className="flex justify-between">
                <Button variant="outline" onClick={onBack}>Back</Button>
                <Button onClick={handleNext}>Next: Authenticate</Button>
            </div>
        </div>
    );
}
