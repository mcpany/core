
import { MetricCard } from "@/components/dashboard/metric-card";
import { MetricsChart } from "@/components/dashboard/metrics-chart";
import { ServiceStatusList } from "@/components/dashboard/service-status-list";
import { Sidebar } from "@/components/sidebar";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, AlertCircle, Clock, Server } from "lucide-react";
import { mockMetrics } from "@/lib/mock-data";

export default function Home() {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40 md:flex-row">
        <Sidebar />
        <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14 w-full">
            <main className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
                <div className="grid gap-4 md:grid-cols-2 md:gap-8 lg:grid-cols-4">
                    <MetricCard
                        title="Total Requests"
                        value={mockMetrics.totalRequests.toLocaleString()}
                        icon={Activity}
                        description="+20.1% from last month"
                    />
                    <MetricCard
                        title="Active Services"
                        value={mockMetrics.activeServices}
                        icon={Server}
                        description="+2 since last hour"
                    />
                    <MetricCard
                        title="Avg Latency"
                        value={`${mockMetrics.avgLatency}ms`}
                        icon={Clock}
                        description="-5% from last hour"
                    />
                    <MetricCard
                        title="Error Rate"
                        value={`${(mockMetrics.errorRate * 100).toFixed(1)}%`}
                        icon={AlertCircle}
                        description="+0.1% from last hour"
                    />
                </div>
                <div className="grid gap-4 md:gap-8 lg:grid-cols-2 xl:grid-cols-3">
                    <MetricsChart />
                    <Card className="xl:col-span-1">
                        <CardHeader>
                            <CardTitle>Service Status</CardTitle>
                            <CardDescription>
                                Recent activity from your services.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <ServiceStatusList />
                        </CardContent>
                    </Card>
                </div>
            </main>
        </div>
    </div>
  );
}
