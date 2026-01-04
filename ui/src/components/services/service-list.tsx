
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { MoreHorizontal, Edit, Trash2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { UpstreamServiceConfig } from "@/lib/client";

interface ServiceListProps {
    services: UpstreamServiceConfig[];
    isLoading: boolean;
    onToggle: (name: string, enabled: boolean) => void;
    onEdit: (service: UpstreamServiceConfig) => void;
    onDelete: (name: string) => void;
}

export function ServiceList({ services, isLoading, onToggle, onEdit, onDelete }: ServiceListProps) {
    if (isLoading) {
        return <div className="p-4 text-center text-muted-foreground">Loading services...</div>;
    }

    if (services.length === 0) {
        return <div className="p-4 text-center text-muted-foreground">No services found.</div>;
    }

    const getType = (service: UpstreamServiceConfig) => {
        if (service.http_service) return "HTTP";
        if (service.grpc_service) return "gRPC";
        if (service.command_line_service) return "CMD";
        if (service.mcp_service) return "MCP";
        if (service.openapi_service) return "OpenAPI";
        return "Unknown";
    };

    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead className="w-[200px]">Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Version</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {services.map((service) => (
                    <TableRow key={service.name}>
                        <TableCell className="font-medium">{service.name}</TableCell>
                        <TableCell>
                            <Badge variant="secondary">{getType(service)}</Badge>
                        </TableCell>
                        <TableCell>{service.version || "N/A"}</TableCell>
                        <TableCell>
                            <div className="flex items-center space-x-2">
                                <Switch
                                    checked={!service.disable}
                                    onCheckedChange={(checked) => onToggle(service.name, checked)}
                                />
                                <span className="text-sm text-muted-foreground">
                                    {service.disable ? "Disabled" : "Enabled"}
                                </span>
                            </div>
                        </TableCell>
                        <TableCell className="text-right">
                             <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button variant="ghost" size="icon">
                                        <MoreHorizontal className="h-4 w-4" />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                    <DropdownMenuItem onClick={() => onEdit(service)}>
                                        <Edit className="mr-2 h-4 w-4" /> Edit
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem className="text-red-600" onClick={() => onDelete(service.name)}>
                                        <Trash2 className="mr-2 h-4 w-4" /> Delete
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </TableCell>
                    </TableRow>
                ))}
            </TableBody>
        </Table>
    );
}
