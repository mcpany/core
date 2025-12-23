
"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, XCircle, AlertTriangle } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";

export function ServiceHealthWidget() {
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/dashboard/health")
      .then(res => res.json())
      .then(data => {
          setServices(data);
          setLoading(false);
      });
  }, []);

  return (
    <Card className="col-span-3 backdrop-blur-xl bg-background/60 border-muted/20 shadow-sm">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>Real-time status of connected upstream services.</CardDescription>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[300px] w-full pr-4">
          {loading ? (
             <div className="space-y-4">
                 <Skeleton className="h-12 w-full" />
                 <Skeleton className="h-12 w-full" />
                 <Skeleton className="h-12 w-full" />
             </div>
          ) : (
            <div className="space-y-4">
                {services.map((service, i) => (
                <div key={i} className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0 border-muted/20">
                    <div className="space-y-1">
                    <p className="text-sm font-medium leading-none">{service.name}</p>
                    <p className="text-xs text-muted-foreground">{service.latency}</p>
                    </div>
                    <div className="flex items-center">
                    {service.status === "healthy" && (
                        <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500/20 hover:bg-green-500/20 transition-colors">
                        <CheckCircle2 className="mr-1 h-3 w-3" /> Healthy
                        </Badge>
                    )}
                    {service.status === "degraded" && (
                        <Badge variant="outline" className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20 hover:bg-yellow-500/20 transition-colors">
                        <AlertTriangle className="mr-1 h-3 w-3" /> Degraded
                        </Badge>
                    )}
                    {service.status === "down" && (
                        <Badge variant="outline" className="bg-red-500/10 text-red-500 border-red-500/20 hover:bg-red-500/20 transition-colors">
                        <XCircle className="mr-1 h-3 w-3" /> Down
                        </Badge>
                    )}
                    </div>
                </div>
                ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
