
import { Sidebar } from "@/components/sidebar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { MoreHorizontal } from "lucide-react"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

// Mock Prompts
const mockPrompts = [
    { name: "summarize_text", description: "Summarizes the given text into a concise paragraph.", arguments: ["text"] },
    { name: "generate_code", description: "Generates code based on a natural language description.", arguments: ["description", "language"] }
]

export default function PromptsPage() {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40 md:flex-row">
      <Sidebar />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14 w-full">
        <main className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
            <div className="flex items-center">
                <h1 className="text-lg font-semibold md:text-2xl">Prompts</h1>
            </div>
          <Card x-chunk="dashboard-06-chunk-0">
            <CardHeader>
              <CardTitle>Prompts</CardTitle>
              <CardDescription>
                Manage pre-defined prompts available for use.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Arguments</TableHead>
                     <TableHead>
                      <span className="sr-only">Actions</span>
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                    {mockPrompts.map((prompt, index) => (
                        <TableRow key={index}>
                            <TableCell className="font-medium">
                                {prompt.name}
                            </TableCell>
                            <TableCell>
                                {prompt.description}
                            </TableCell>
                            <TableCell>
                                {prompt.arguments.map(arg => (
                                    <Badge key={arg} variant="outline" className="mr-1">{arg}</Badge>
                                ))}
                            </TableCell>
                            <TableCell>
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button
                                            aria-haspopup="true"
                                            size="icon"
                                            variant="ghost"
                                        >
                                            <MoreHorizontal className="h-4 w-4" />
                                            <span className="sr-only">Toggle menu</span>
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                        <DropdownMenuItem>Execute</DropdownMenuItem>
                                        <DropdownMenuItem>Edit</DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  )
}
