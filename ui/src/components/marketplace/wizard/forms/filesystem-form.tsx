/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { FolderOpen, Plus, Trash2 } from 'lucide-react';

interface FilesystemFormProps {
    paths: string[];
    onChange: (paths: string[]) => void;
}

/**
 * FilesystemForm component.
 * A smart form for configuring the Filesystem MCP server.
 *
 * @param props - The component props.
 * @returns The rendered component.
 */
export function FilesystemForm({ paths, onChange }: FilesystemFormProps) {
    const [newPath, setNewPath] = useState("");

    const addPath = () => {
        if (!newPath) return;
        onChange([...paths, newPath]);
        setNewPath("");
    };

    const removePath = (index: number) => {
        const newPaths = [...paths];
        newPaths.splice(index, 1);
        onChange(newPaths);
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center gap-2 mb-4">
                <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-full">
                    <FolderOpen className="h-5 w-5 text-amber-600 dark:text-amber-300" />
                </div>
                <div>
                    <h3 className="text-lg font-medium">Filesystem Configuration</h3>
                    <p className="text-sm text-muted-foreground">Specify directories allowed for access.</p>
                </div>
            </div>

            <Card>
                <CardContent className="pt-6 space-y-4">
                    <div className="space-y-2">
                        <Label>Allowed Directories</Label>
                        {paths.map((path, idx) => (
                            <div key={idx} className="flex gap-2">
                                <Input value={path} readOnly className="font-mono text-sm bg-muted" />
                                <Button variant="outline" size="icon" onClick={() => removePath(idx)}>
                                    <Trash2 className="h-4 w-4 text-destructive" />
                                </Button>
                            </div>
                        ))}
                        {paths.length === 0 && (
                            <div className="text-sm text-muted-foreground italic text-center py-4 border border-dashed rounded-md">
                                No directories added.
                            </div>
                        )}
                    </div>

                    <div className="flex gap-2">
                        <Input
                            placeholder="/path/to/directory"
                            value={newPath}
                            onChange={(e) => setNewPath(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && addPath()}
                        />
                        <Button onClick={addPath}>
                            <Plus className="h-4 w-4 mr-2" /> Add
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
