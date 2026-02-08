"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { Loader2, ArrowLeft, Trash2, Activity, Wrench, FileText, Terminal, Settings, Eye } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ServiceOverview } from "@/components/services/service-overview";
import { ServiceTools } from "@/components/services/service-tools";
import { ServiceResources } from "@/components/services/service-resources";
import { ServiceEditor } from "@/components/services/editor/service-editor";
import { LogStream } from "@/components/logs/log-stream";
import { ServiceInspector } from "@/components/services/editor/service-inspector";

export default function UpstreamServiceDetailPage() {
    const params = useParams();
    const router = useRouter();
    const { toast } = useToast();
    const serviceId = params.serviceId as string;

    const [service, setService] = useState<UpstreamServiceConfig | null>(null);
    const [status, setStatus] = useState<any>(null);
    const [trafficData, setTrafficData] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState("overview");

    const fetchServiceData = async () => {
        if (!serviceId) return;
        try {
            // Fetch config
            const data = await apiClient.getService(serviceId);
            const loadedService = data.service || data;
            setService(loadedService);

            // Fetch status (runtime info)
            try {
                const statusData = await apiClient.getServiceStatus(loadedService.name || serviceId);
                setStatus(statusData);
            } catch (err) {
                 console.warn("Failed to fetch status (service might be down/disabled):", err);
                 setStatus(null);
            }

            // Fetch traffic history
            try {
                const traffic = await apiClient.getDashboardTraffic(loadedService.name || serviceId, "1h");
                setTrafficData(Array.isArray(traffic) ? traffic : []);
            } catch (err) {
                console.warn("Failed to fetch traffic data:", err);
                setTrafficData([]);
            }

        } catch (e) {
            console.error(e);
            toast({ title: "Failed to load service", description: "Service not found or error occurred.", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchServiceData();
    }, [serviceId, toast]);

    const handleSave = async () => {
        if (!service) return;
        try {
            await apiClient.updateService(service);
            toast({ title: "Service Updated", description: "Configuration saved successfully." });
            fetchServiceData(); // Refresh data
        } catch (e) {
            toast({ title: "Update Failed", description: String(e), variant: "destructive" });
        }
    };

    const handleUnregister = async () => {
        if (!confirm("Are you sure you want to unregister this service? This action cannot be undone.")) return;
        try {
            await apiClient.unregisterService(serviceId);
            toast({ title: "Service Unregistered" });
            router.push("/upstream-services");
        } catch (e) {
            toast({ title: "Unregister Failed", description: String(e), variant: "destructive" });
        }
    };

    if (loading) {
        return <div className="flex h-screen items-center justify-center"><Loader2 className="h-8 w-8 animate-spin" /></div>;
    }

    if (!service) {
        return (
            <div className="p-8 text-center">
                <h1 className="text-2xl font-bold">Service Not Found</h1>
                <Button variant="link" onClick={() => router.push("/upstream-services")}>Back to Services</Button>
            </div>
        );
    }

    // Extract resources from config
    const resources = service.httpService?.resources ||
                      service.grpcService?.resources ||
                      service.mcpService?.resources ||
                      service.openapiService?.resources ||
                      service.commandLineService?.resources ||
                      [];

    // Tools from status (dynamic) or config (static)
    const tools = status?.tools ||
                  service.httpService?.tools ||
                  service.grpcService?.tools ||
                  service.mcpService?.tools ||
                  service.openapiService?.tools ||
                  service.commandLineService?.tools ||
                  [];

    return (
        <div className="flex flex-col h-screen overflow-hidden bg-background">
             {/* Header */}
            <div className="flex-none border-b p-4 flex items-center justify-between bg-muted/20">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" onClick={() => router.push("/upstream-services")}>
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                    <div>
                        <h1 className="text-2xl font-bold tracking-tight flex items-center gap-3">
                            {service.name}
                            <Badge variant={service.disable ? "secondary" : "default"} className={service.disable ? "bg-muted text-muted-foreground" : "bg-green-500 hover:bg-green-600"}>
                                {service.disable ? "Disabled" : "Active"}
                            </Badge>
                        </h1>
                         <p className="text-muted-foreground text-xs font-mono">{service.id || "ID not assigned"}</p>
                    </div>
                </div>
                <div className="flex gap-2">
                     <Button variant="destructive" size="sm" onClick={handleUnregister}>
                        <Trash2 className="mr-2 h-4 w-4" />
                        Unregister
                    </Button>
                </div>
            </div>

            {/* Dashboard Content */}
            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
                <div className="border-b px-4 flex-none bg-background/50 backdrop-blur-sm z-10">
                    <TabsList className="bg-transparent h-12">
                        <TabsTrigger value="overview" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <Activity className="mr-2 h-4 w-4" /> Overview
                        </TabsTrigger>
                        <TabsTrigger value="tools" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <Wrench className="mr-2 h-4 w-4" /> Tools <Badge variant="secondary" className="ml-2 text-[10px] h-4 px-1">{tools.length}</Badge>
                        </TabsTrigger>
                        <TabsTrigger value="resources" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <FileText className="mr-2 h-4 w-4" /> Resources <Badge variant="secondary" className="ml-2 text-[10px] h-4 px-1">{resources.length}</Badge>
                        </TabsTrigger>
                        <TabsTrigger value="logs" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <Terminal className="mr-2 h-4 w-4" /> Logs
                        </TabsTrigger>
                        <TabsTrigger value="inspector" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <Eye className="mr-2 h-4 w-4" /> Inspector
                        </TabsTrigger>
                         <TabsTrigger value="settings" className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary">
                            <Settings className="mr-2 h-4 w-4" /> Settings
                        </TabsTrigger>
                    </TabsList>
                </div>

                <div className="flex-1 overflow-hidden bg-muted/5">
                    <TabsContent value="overview" className="h-full p-6 overflow-y-auto m-0">
                        <ServiceOverview service={service} status={status} trafficData={trafficData} />
                    </TabsContent>

                    <TabsContent value="tools" className="h-full p-6 overflow-y-auto m-0">
                        <ServiceTools tools={tools} />
                    </TabsContent>

                    <TabsContent value="resources" className="h-full p-6 overflow-y-auto m-0">
                        <ServiceResources resources={resources} />
                    </TabsContent>

                    <TabsContent value="logs" className="h-full p-4 m-0">
                        <div className="h-full rounded-lg border bg-background overflow-hidden">
                             <LogStream source={service.name} />
                        </div>
                    </TabsContent>

                    <TabsContent value="inspector" className="h-full p-6 overflow-y-auto m-0">
                        <ServiceInspector service={service} />
                    </TabsContent>

                    <TabsContent value="settings" className="h-full overflow-hidden m-0">
                         {/* ServiceEditor handles its own scrolling */}
                         <ServiceEditor
                            service={service}
                            onChange={setService}
                            onSave={handleSave}
                            onCancel={() => router.push("/upstream-services")}
                        />
                    </TabsContent>
                </div>
            </Tabs>
        </div>
    );
}
