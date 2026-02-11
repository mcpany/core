/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Download, Upload, AlertTriangle, RefreshCw } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

export function BackupRestore() {
    const { toast } = useToast();
    const [isBackingUp, setIsBackingUp] = useState(false);
    const [isRestoring, setIsRestoring] = useState(false);
    const [restoreData, setRestoreData] = useState<any>(null);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleDownloadBackup = async () => {
        setIsBackingUp(true);
        try {
            const data = await apiClient.createBackup();
            const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
            const url = URL.createObjectURL(blob);
            const link = document.createElement("a");
            link.href = url;
            link.download = `mcpany-backup-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);
            toast({ title: "Backup Created", description: "System snapshot downloaded successfully." });
        } catch (e) {
            console.error(e);
            toast({ variant: "destructive", title: "Backup Failed", description: "Could not create backup." });
        } finally {
            setIsBackingUp(false);
        }
    };

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (ev) => {
            try {
                const json = JSON.parse(ev.target?.result as string);
                setRestoreData(json);
            } catch (err) {
                toast({ variant: "destructive", title: "Invalid File", description: "The selected file is not a valid JSON backup." });
            }
        };
        reader.readAsText(file);
        e.target.value = ''; // Reset
    };

    const handleConfirmRestore = async () => {
        if (!restoreData) return;
        setIsRestoring(true);
        try {
            const res = await apiClient.restoreBackup(restoreData);
            toast({
                title: "Restore Successful",
                description: `Restored ${res.servicesRestored} services, ${res.usersRestored} users, ${res.secretsRestored} secrets.`,
            });
            setRestoreData(null);
            // Optionally reload page to reflect changes
            setTimeout(() => window.location.reload(), 2000);
        } catch (e) {
            console.error(e);
            toast({ variant: "destructive", title: "Restore Failed", description: String(e) });
        } finally {
            setIsRestoring(false);
        }
    };

    return (
        <Card className="backdrop-blur-sm bg-background/50 border-orange-500/20">
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    Maintenance
                </CardTitle>
                <CardDescription>
                    Backup and restore system configuration.
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="flex flex-col md:flex-row gap-6">
                    <div className="flex-1 space-y-4">
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-primary/10 rounded-full">
                                <Download className="h-5 w-5 text-primary" />
                            </div>
                            <div>
                                <h4 className="font-medium">Backup Configuration</h4>
                                <p className="text-sm text-muted-foreground">Download a snapshot of all services, users, and settings.</p>
                            </div>
                        </div>
                        <Button onClick={handleDownloadBackup} disabled={isBackingUp} className="w-full md:w-auto">
                            {isBackingUp ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
                            Download Snapshot
                        </Button>
                    </div>

                    <div className="w-px bg-border hidden md:block" />

                    <div className="flex-1 space-y-4">
                        <div className="flex items-center gap-2">
                            <div className="p-2 bg-destructive/10 rounded-full">
                                <Upload className="h-5 w-5 text-destructive" />
                            </div>
                            <div>
                                <h4 className="font-medium">Restore Configuration</h4>
                                <p className="text-sm text-muted-foreground">Overwrite current state with a backup file.</p>
                            </div>
                        </div>
                        <Button variant="outline" onClick={() => fileInputRef.current?.click()} disabled={isRestoring} className="w-full md:w-auto">
                            <Upload className="mr-2 h-4 w-4" />
                            Upload Backup
                        </Button>
                        <input
                            type="file"
                            ref={fileInputRef}
                            className="hidden"
                            accept=".json"
                            onChange={handleFileSelect}
                        />
                    </div>
                </div>
            </CardContent>

            <AlertDialog open={!!restoreData} onOpenChange={(open) => !open && setRestoreData(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle className="flex items-center gap-2 text-destructive">
                            <AlertTriangle className="h-5 w-5" />
                            Confirm Restore
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                            This action will <strong>overwrite</strong> your current configuration with the data from the backup file.
                            Existing services, users, and secrets may be updated or replaced.
                            <br/><br/>
                            Are you sure you want to proceed?
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel disabled={isRestoring}>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={handleConfirmRestore} className="bg-destructive hover:bg-destructive/90" disabled={isRestoring}>
                            {isRestoring ? "Restoring..." : "Yes, Restore Everything"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </Card>
    );
}
