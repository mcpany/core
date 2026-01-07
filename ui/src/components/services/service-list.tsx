
import { UpstreamServiceConfig } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { StatusBadge } from "@/components/layout/status-badge";
import { Button } from "@/components/ui/button";
import { Edit2, Trash2, Power, PowerOff } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

interface ServiceListProps {
  services: UpstreamServiceConfig[];
  isLoading: boolean;
  onToggle: (name: string, enabled: boolean) => void;
  onEdit: (service: UpstreamServiceConfig) => void;
  onDelete: (name: string) => void;
}

export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete }: ServiceListProps) {
  if (isLoading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    );
  }

  if (services.length === 0) {
      return <div className="p-4 text-center text-muted-foreground">No services configured.</div>;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Priority</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {services.map((service) => {
           let type = "Unknown";
           if (service.httpService) type = "HTTP";
           else if (service.grpcService) type = "gRPC";
           else if (service.commandLineService) type = "Command Line";
           else if (service.mcpService) type = "MCP Proxy";

           const isActive = !service.disable;

           return (
              <TableRow key={service.id || service.name}>
                <TableCell className="font-medium">{service.name}</TableCell>
                <TableCell>{type}</TableCell>
                <TableCell>
                  <StatusBadge status={isActive ? "active" : "inactive"} />
                </TableCell>
                <TableCell>{service.priority}</TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => onToggle(service.name, !isActive)}
                        title={isActive ? "Disable" : "Enable"}
                    >
                        {isActive ? <Power className="h-4 w-4 text-green-600" /> : <PowerOff className="h-4 w-4 text-slate-400" />}
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => onEdit(service)}>
                      <Edit2 className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" onClick={() => onDelete(service.name)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
           );
        })}
      </TableBody>
    </Table>
  );
}
