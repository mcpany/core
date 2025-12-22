
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const mockResources = [
  { name: "api_spec.yaml", description: "OpenAPI Specification", type: "text/yaml" },
  { name: "logo.png", description: "Company Logo", type: "image/png" },
];

export default function ResourcesPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Type</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mockResources.map((res) => (
              <TableRow key={res.name}>
                <TableCell className="font-medium">{res.name}</TableCell>
                <TableCell>{res.description}</TableCell>
                <TableCell>
                    <Badge variant="outline">{res.type}</Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
