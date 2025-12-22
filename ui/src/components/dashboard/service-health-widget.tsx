
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, XCircle, AlertCircle } from "lucide-react";

// Mock data based on proto/config/v1/upstream_service.proto
interface ServiceStatus {
  id: string;
  name: string;
  status: "healthy" | "unhealthy" | "degraded";
  uptime: string;
  version: string;
}

const services: ServiceStatus[] = [
  { id: "1", name: "Payment Gateway", status: "healthy", uptime: "99.99%", version: "v1.2.0" },
  { id: "2", name: "User Service", status: "healthy", uptime: "99.95%", version: "v2.1.0" },
  { id: "3", name: "Notification Service", status: "degraded", uptime: "98.50%", version: "v1.0.1" },
  { id: "4", name: "Search Indexer", status: "unhealthy", uptime: "85.00%", version: "v0.9.0" },
  { id: "5", name: "Analytics Engine", status: "healthy", uptime: "99.90%", version: "v3.0.0" },
];

export function ServiceHealthWidget() {
  return (
    <Card className="col-span-3 backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>Real-time status of connected upstream services.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {services.map((service) => (
            <div key={service.id} className="flex items-center justify-between p-2 rounded-lg hover:bg-muted/50 transition-colors">
              <div className="flex items-center space-x-4">
                {service.status === "healthy" && <CheckCircle2 className="text-green-500 h-5 w-5" />}
                {service.status === "degraded" && <AlertCircle className="text-yellow-500 h-5 w-5" />}
                {service.status === "unhealthy" && <XCircle className="text-red-500 h-5 w-5" />}
                <div>
                  <p className="text-sm font-medium leading-none">{service.name}</p>
                  <p className="text-xs text-muted-foreground">{service.version}</p>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <div className="text-right">
                    <p className="text-sm font-medium">{service.uptime}</p>
                    <p className="text-xs text-muted-foreground">Uptime</p>
                </div>
                <Badge variant={service.status === "healthy" ? "default" : service.status === "degraded" ? "secondary" : "destructive"}>
                    {service.status}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
