
import {
  Users,
  Activity,
  Server,
  Zap,
  ArrowUpRight,
  ArrowDownRight
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface Metric {
  label: string;
  value: string;
  change?: string;
  trend?: "up" | "down" | "neutral";
  icon: any;
}

const metrics: Metric[] = [
  {
    label: "Total Requests",
    value: "1.2M",
    change: "+12.5%",
    trend: "up",
    icon: Activity,
  },
  {
    label: "Active Services",
    value: "14",
    change: "+2",
    trend: "up",
    icon: Server,
  },
  {
    label: "Avg Latency",
    value: "45ms",
    change: "-5ms",
    trend: "down", // down is good for latency, but usually green. handled by component logic?
    icon: Zap,
  },
  {
    label: "Active Users",
    value: "573",
    change: "+201",
    trend: "up",
    icon: Users,
  },
];

export function MetricsOverview() {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {metrics.map((metric) => (
        <Card key={metric.label} className="backdrop-blur-sm bg-background/50 border shadow-sm hover:shadow-md transition-all">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {metric.label}
            </CardTitle>
            <metric.icon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metric.value}</div>
            <p className="text-xs text-muted-foreground flex items-center mt-1">
              {metric.trend === "up" ? (
                 <ArrowUpRight className="h-3 w-3 text-green-500 mr-1" />
              ) : (
                 <ArrowDownRight className="h-3 w-3 text-green-500 mr-1" />
              )}
              <span className={metric.trend === "up" ? "text-green-500" : "text-green-500"}>
                {metric.change}
              </span>
              <span className="ml-1">from last month</span>
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
