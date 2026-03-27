/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Download, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

/**
 * Intent: Document DownloadReportButton
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * A button component that fetches dashboard metrics and downloads them as a JSON report.
 * @returns The rendered component.
 */
export function DownloadReportButton() {
    const [isDownloading, setIsDownloading] = useState(false);
    const { toast } = useToast();

    const handleDownload = async () => {
        setIsDownloading(true);
        try {
            // Gather data for the report
            const services = await apiClient.listServices();
            const metrics = await apiClient.getDashboardMetrics();
            const tools = await apiClient.getTopTools();
            const failures = await apiClient.getToolFailures();

            const report = {
                generatedAt: new Date().toISOString(),
                summary: {
                    totalServices: services.length,
                    activeTools: tools.length,
                    toolFailures: failures.length,
                },
                metrics,
                services: services.map((s: any) => ({ name: s.name, version: s.version, type: s.httpService ? "HTTP" : s.grpcService ? "gRPC" : "Other" })),
                topTools: tools,
                recentFailures: failures
            };

            const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(report, null, 2));
            const downloadAnchorNode = document.createElement('a');
            downloadAnchorNode.setAttribute("href", dataStr);
            downloadAnchorNode.setAttribute("download", `mcpany-report-${new Date().toISOString().split('T')[0]}.json`);
            document.body.appendChild(downloadAnchorNode);
            downloadAnchorNode.click();
            downloadAnchorNode.remove();

            toast({
                title: "Report Downloaded",
                description: "Your dashboard report has been successfully downloaded.",
            });
        } catch (error) {
            console.error("Failed to download report", error);
            toast({
                variant: "destructive",
                title: "Download Failed",
                description: "There was an error generating the report.",
            });
        } finally {
            setIsDownloading(false);
        }
    };

    return (
        <Button onClick={handleDownload} disabled={isDownloading}>
            {isDownloading ? (
                <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Generating...
                </>
            ) : (
                <>
                    <Download className="mr-2 h-4 w-4" />
                    Download Report
                </>
            )}
        </Button>
    );
}
