/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ProxyConfig } from "@proto/config/v1/upstream_service";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface ProxyConfigEditorProps {
    config?: ProxyConfig;
    onChange: (config: ProxyConfig) => void;
}

export function ProxyConfigEditor({ config, onChange }: ProxyConfigEditorProps) {
    const update = (updates: Partial<ProxyConfig>) => {
        const current = config || { url: "", username: "", password: "" };
        onChange({ ...current, ...updates });
    };

    return (
        <div className="space-y-4 border rounded-lg p-4 bg-muted/20">
            <h4 className="text-sm font-medium">Proxy Configuration</h4>
            <div className="space-y-2">
                <Label htmlFor="proxy-url">Proxy URL</Label>
                <Input
                    id="proxy-url"
                    value={config?.url || ""}
                    onChange={(e) => update({ url: e.target.value })}
                    placeholder="http://proxy.example.com:8080"
                />
            </div>
            <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                    <Label htmlFor="proxy-username">Username (Optional)</Label>
                    <Input
                        id="proxy-username"
                        value={config?.username || ""}
                        onChange={(e) => update({ username: e.target.value })}
                        placeholder="user"
                    />
                </div>
                <div className="space-y-2">
                    <Label htmlFor="proxy-password">Password (Optional)</Label>
                    <Input
                        id="proxy-password"
                        type="password"
                        value={config?.password || ""}
                        onChange={(e) => update({ password: e.target.value })}
                        placeholder="pass"
                    />
                </div>
            </div>
        </div>
    );
}
