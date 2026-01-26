/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Database } from 'lucide-react';

interface PostgresFormProps {
    connectionString: string;
    onChange: (val: string) => void;
}

/**
 * PostgresForm component.
 * A smart form for configuring the PostgreSQL MCP server.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function PostgresForm({ connectionString, onChange }: PostgresFormProps) {
    return (
        <div className="space-y-4">
            <div className="flex items-center gap-2 mb-4">
                <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-full">
                    <Database className="h-5 w-5 text-blue-600 dark:text-blue-300" />
                </div>
                <div>
                    <h3 className="text-lg font-medium">PostgreSQL Configuration</h3>
                    <p className="text-sm text-muted-foreground">Configure the connection to your database.</p>
                </div>
            </div>

            <Card>
                <CardContent className="pt-6">
                    <div className="grid gap-4">
                        <div className="grid gap-2">
                            <Label htmlFor="postgres-url">Connection URL</Label>
                            <Input
                                id="postgres-url"
                                placeholder="postgresql://user:password@localhost:5432/dbname"
                                value={connectionString}
                                onChange={(e) => onChange(e.target.value)}
                            />
                            <p className="text-xs text-muted-foreground">
                                Format: <code>postgresql://[user[:password]@][netloc][:port][/dbname]</code>
                            </p>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
