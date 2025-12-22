
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

const mockTools = [
  { name: "stripe_charge", description: "Create a charge on Stripe", service: "Payment Gateway", type: "function" },
  { name: "get_user", description: "Retrieve user details", service: "User Service", type: "function" },
  { name: "search_docs", description: "Search internal documentation", service: "Search Indexer", type: "read" },
];

export default function ToolsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead>Type</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mockTools.map((tool) => (
              <TableRow key={tool.name}>
                <TableCell className="font-medium">{tool.name}</TableCell>
                <TableCell>{tool.description}</TableCell>
                <TableCell>{tool.service}</TableCell>
                <TableCell>
                    <Badge variant="secondary">{tool.type}</Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
