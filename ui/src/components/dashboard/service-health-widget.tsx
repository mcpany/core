
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { GlassCard } from "@/components/layout/glass-card";
import { StatusBadge } from "@/components/layout/status-badge";
import { CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";

export interface ServiceHealthData {
  name: string;
  status: string;
  uptime: string;
  latency: string;
}

export function ServiceHealthWidget({ services }: { services: ServiceHealthData[] }) {
  return (
    <GlassCard className="col-span-3">
      <CardHeader>
        <CardTitle>Service Health</CardTitle>
        <CardDescription>Real-time status of critical upstream services.</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Service Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Uptime</TableHead>
              <TableHead>Latency</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {services.map((service) => (
              <TableRow key={service.name}>
                <TableCell className="font-medium">{service.name}</TableCell>
                <TableCell>
                  <StatusBadge status={service.status as any} />
                </TableCell>
                <TableCell>{service.uptime}</TableCell>
                <TableCell>{service.latency}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </GlassCard>
  );
}
