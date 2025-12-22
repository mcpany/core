
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const mockPrompts = [
  { name: "summarize_text", description: "Summarizes the given text", arguments: ["text", "length"] },
  { name: "code_review", description: "Reviews code for bugs", arguments: ["code", "language"] },
];

export default function PromptsPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Arguments</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mockPrompts.map((prompt) => (
              <TableRow key={prompt.name}>
                <TableCell className="font-medium">{prompt.name}</TableCell>
                <TableCell>{prompt.description}</TableCell>
                <TableCell>{prompt.arguments.join(", ")}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
