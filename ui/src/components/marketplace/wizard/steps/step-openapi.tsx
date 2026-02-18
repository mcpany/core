/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { useWizard } from '../wizard-context';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

/**
 * StepOpenAPI component.
 * @returns The rendered component.
 */
export function StepOpenAPI() {
    const { state, updateConfig } = useWizard();
    const { config } = state;

    return (
        <div className="space-y-4">
            <h3 className="text-lg font-medium">OpenAPI Specification</h3>

            <div className="space-y-2">
                <Label htmlFor="spec-url">Specification URL</Label>
                <Input
                    id="spec-url"
                    placeholder="https://example.com/openapi.json"
                    value={config.openapiService?.specUrl || ''}
                    onChange={e => updateConfig({
                        openapiService: {
                            ...config.openapiService!,
                            specUrl: e.target.value
                        }
                    })}
                />
            </div>

            <div className="space-y-2">
                <Label htmlFor="address">Service Address</Label>
                <Input
                    id="address"
                    placeholder="https://api.example.com"
                    value={config.openapiService?.address || ''}
                    onChange={e => updateConfig({
                        openapiService: {
                            ...config.openapiService!,
                            address: e.target.value
                        }
                    })}
                />
            </div>
        </div>
    );
}
