
import { useEffect, useState } from "react";
import { apiClient, DiscoveryProviderStatus } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { CheckCircle2, XCircle, RefreshCw, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { formatDistanceToNow } from "date-fns";

export function DiscoveryStatus() {
    const [statuses, setStatuses] = useState<DiscoveryProviderStatus[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchStatus = async () => {
        setLoading(true);
        try {
            const res = await apiClient.getDiscoveryStatus();
            setStatuses(res.providers || []);
            setError(null);
        } catch (e) {
            console.error("Failed to fetch discovery status", e);
            setError("Failed to load discovery status");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStatus();
    }, []);

    if (statuses.length === 0 && !loading && !error) {
        return null; // Don't show if no providers configured/reported
    }

    return (
        <Card className="backdrop-blur-sm bg-background/50 mb-6 border-dashed">
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-lg font-medium">Auto-Discovery Status</CardTitle>
                        <CardDescription>Status of local tool discovery providers.</CardDescription>
                    </div>
                    <Button variant="ghost" size="icon" onClick={fetchStatus} disabled={loading}>
                        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                    </Button>
                </div>
            </CardHeader>
            <CardContent>
                {error ? (
                     <div className="text-destructive text-sm">{error}</div>
                ) : (
                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {statuses.map((provider) => (
                            <div key={provider.name} className="flex items-start space-x-3 p-3 rounded-md border bg-card/50">
                                {provider.status === "OK" ? (
                                    <CheckCircle2 className="h-5 w-5 text-green-500 mt-0.5" />
                                ) : (
                                    <XCircle className="h-5 w-5 text-destructive mt-0.5" />
                                )}
                                <div className="space-y-1">
                                    <div className="font-medium leading-none flex items-center gap-2">
                                        {provider.name}
                                        <span className="text-xs text-muted-foreground font-normal">
                                            ({provider.discoveredCount} tools)
                                        </span>
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        {provider.lastRunAt && (
                                            <span>Updated {formatDistanceToNow(new Date(provider.lastRunAt), { addSuffix: true })}</span>
                                        )}
                                    </p>
                                    {provider.status !== "OK" && provider.lastError && (
                                        <p className="text-xs text-destructive mt-1 break-all">
                                            {provider.lastError}
                                        </p>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
