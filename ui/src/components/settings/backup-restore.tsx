/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Download, Upload, Loader2, AlertTriangle, RefreshCw } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

/**
 * BackupRestore component.
 * Provides controls for backing up and restoring the system configuration.
 * @returns The rendered component.
 */
export function BackupRestore() {
    const [creatingBackup, setCreatingBackup] = useState(false);
    const [restoringBackup, setRestoringBackup] = useState(false);
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [isRestoreDialogOpen, setIsRestoreDialogOpen] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const { toast } = useToast();

    const handleCreateBackup = async () => {
        setCreatingBackup(true);
        try {
            const { blob, filename } = await apiClient.createBackup();
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            toast({
                title: "Backup Created",
                description: "System configuration backup has been downloaded.",
            });
        } catch (e) {
            console.error("Backup failed", e);
            toast({
                variant: "destructive",
                title: "Backup Failed",
                description: e instanceof Error ? e.message : "An error occurred while creating backup.",
            });
        } finally {
            setCreatingBackup(false);
        }
    };

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files.length > 0) {
            setSelectedFile(e.target.files[0]);
            setIsRestoreDialogOpen(true);
        }
    };

    const handleRestoreBackup = async () => {
        if (!selectedFile) return;
        setRestoringBackup(true);
        try {
            await apiClient.restoreBackup(selectedFile);
            toast({
                title: "Restore Successful",
                description: "System configuration has been restored. Reloading page...",
            });
            setTimeout(() => {
                window.location.reload();
            }, 2000);
        } catch (e) {
            console.error("Restore failed", e);
            toast({
                variant: "destructive",
                title: "Restore Failed",
                description: e instanceof Error ? e.message : "An error occurred while restoring backup.",
            });
            setRestoringBackup(false);
        }
    };

    return (
        <Card className="backdrop-blur-sm bg-background/50 border-orange-500/20">
            <CardHeader>
                <CardTitle className="flex items-center gap-2">
                    <RefreshCw className="h-5 w-5 text-orange-500" />
                    Backup & Restore
                </CardTitle>
                <CardDescription>
                    Create snapshots of your system configuration or restore from a previous backup.
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
                <Alert variant="default" className="bg-orange-500/10 border-orange-500/20 text-orange-700 dark:text-orange-400">
                    <AlertTriangle className="h-4 w-4 text-orange-500" />
                    <AlertTitle>Important</AlertTitle>
                    <AlertDescription>
                        Backups contain sensitive information (secrets, credentials) in plaintext. Store them securely.
                        Restoring a backup will overwrite existing configurations.
                    </AlertDescription>
                </Alert>

                <div className="flex flex-col sm:flex-row gap-4">
                    <Button
                        onClick={handleCreateBackup}
                        disabled={creatingBackup || restoringBackup}
                        className="flex-1"
                        variant="outline"
                    >
                        {creatingBackup ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Creating Backup...
                            </>
                        ) : (
                            <>
                                <Download className="mr-2 h-4 w-4" /> Download Snapshot
                            </>
                        )}
                    </Button>

                    <div className="flex-1">
                        <input
                            type="file"
                            accept=".json"
                            className="hidden"
                            ref={fileInputRef}
                            onChange={handleFileSelect}
                        />
                        <Button
                            onClick={() => fileInputRef.current?.click()}
                            disabled={creatingBackup || restoringBackup}
                            className="w-full"
                            variant="secondary"
                        >
                            <Upload className="mr-2 h-4 w-4" /> Restore from Snapshot
                        </Button>
                    </div>
                </div>

                <Dialog open={isRestoreDialogOpen} onOpenChange={setIsRestoreDialogOpen}>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Confirm Restore</DialogTitle>
                            <DialogDescription>
                                Are you sure you want to restore from <strong>{selectedFile?.name}</strong>?
                                <br /><br />
                                <span className="text-red-500 font-bold">Warning:</span> This will overwrite your current configuration (Users, Services, Secrets). This action cannot be undone.
                            </DialogDescription>
                        </DialogHeader>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setIsRestoreDialogOpen(false)} disabled={restoringBackup}>
                                Cancel
                            </Button>
                            <Button onClick={handleRestoreBackup} variant="destructive" disabled={restoringBackup}>
                                {restoringBackup ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Restoring...
                                    </>
                                ) : (
                                    "Yes, Restore System"
                                )}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </CardContent>
        </Card>
    );
}
