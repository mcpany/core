/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from 'react';
import { useWizard } from '../wizard-context';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { apiClient } from '@/lib/client';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { RotateCw } from 'lucide-react';
import { Credential } from '@proto/config/v1/auth';

/**
 * StepAuth component.
 * @returns The rendered component.
 */
export function StepAuth() {
    const { state, updateConfig } = useWizard();
    const { config } = state;
    const [credentials, setCredentials] = useState<Credential[]>([]);
    const [loading, setLoading] = useState(false);
    const [selectedCredId, setSelectedCredId] = useState<string>('none');

    const loadCredentials = async () => {
        setLoading(true);
        try {
            const list = await apiClient.listCredentials();
            setCredentials(list);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadCredentials();
    }, []);

    const handleSelect = (val: string) => {
        setSelectedCredId(val);
        if (val === 'none') {
            updateConfig({ upstreamAuth: undefined });
        } else {
            const cred = credentials.find(c => c.id === val);
            if (cred && cred.authentication) {
                updateConfig({ upstreamAuth: cred.authentication });
            }
        }
    };

    return (
        <div className="space-y-4">
            <h3 className="text-lg font-medium">Authentication</h3>

            <div className="space-y-2">
                <Label htmlFor="auth-select">Select Credential (Optional)</Label>
                <div className="flex gap-2">
                    <Select value={selectedCredId} onValueChange={handleSelect}>
                        <SelectTrigger id="auth-select" className="w-full">
                            <SelectValue placeholder="No Authentication" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="none">None</SelectItem>
                            {credentials.map(c => (
                                <SelectItem key={c.id || c.name} value={c.id || c.name}>{c.name}</SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <Button size="icon" variant="ghost" onClick={loadCredentials} disabled={loading}>
                        <RotateCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    </Button>
                </div>
                <p className="text-sm text-muted-foreground">
                    Bind a saved credential to this service. The authentication configuration (e.g. API keys, OAuth tokens) will be applied.
                </p>
            </div>

            {config.upstreamAuth && (
                <div className="p-3 bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-300 rounded-md text-sm border border-green-200 dark:border-green-900">
                    Authentication configured from credential.
                </div>
            )}
        </div>
    );
}
