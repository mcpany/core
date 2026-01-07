
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, Server, Zap, Database } from "lucide-react";
import { GlassCard } from "@/components/layout/glass-card";

export interface MetricsData {
  activeServices: number;
  requestsPerSec: number;
  avgLatency: number;
  activeResources: number;
}

interface MetricCardProps {
  title: string;
  value: string;
  description: string;
  icon: React.ElementType;
}

function MetricCard({ title, value, description, icon: Icon }: MetricCardProps) {
  return (
    <GlassCard>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">{description}</p>
      </CardContent>
    </GlassCard>
  );
}

export function MetricsOverview({ data }: { data: MetricsData }) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <MetricCard
        title="Active Services"
        value={data.activeServices.toString()}
        description="+2 since last hour"
        icon={Server}
      />
      <MetricCard
        title="Requests/sec"
        value={data.requestsPerSec.toLocaleString()}
        description="+15% from average"
        icon={Activity}
      />
      <MetricCard
        title="Avg Latency"
        value={`${data.avgLatency}ms`}
        description="-5ms improvement"
        icon={Zap}
      />
      <MetricCard
        title="Active Resources"
        value={data.activeResources.toLocaleString()}
        description="Across 8 services"
        icon={Database}
      />
    </div>
  );
}
