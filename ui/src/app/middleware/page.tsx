
import { Sidebar } from "@/components/sidebar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { ArrowDown, GripVertical, Plus } from "lucide-react"

// Mock Middleware Pipeline
const mockPipeline = [
    { id: 1, name: "Authentication Guard", type: "Security", status: "Active" },
    { id: 2, name: "Rate Limiter", type: "Traffic Control", status: "Active" },
    { id: 3, name: "Logging Interceptor", type: "Observability", status: "Active" },
    { id: 4, name: "Response Caching", type: "Performance", status: "Disabled" },
]

export default function MiddlewarePage() {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40 md:flex-row">
      <Sidebar />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14 w-full">
        <main className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
            <div className="flex items-center">
                <h1 className="text-lg font-semibold md:text-2xl">Middleware Pipeline</h1>
                <div className="ml-auto flex items-center gap-2">
                    <Button size="sm" className="h-8 gap-1">
                        <Plus className="h-3.5 w-3.5" />
                        <span className="sr-only sm:not-sr-only sm:whitespace-nowrap">
                            Add Middleware
                        </span>
                    </Button>
                </div>
            </div>
          <Card className="max-w-2xl mx-auto w-full">
            <CardHeader>
              <CardTitle>Global Request Pipeline</CardTitle>
              <CardDescription>
                Configure the sequence of middleware that processes every request.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="flex flex-col items-center space-y-2">
                     <div className="p-3 bg-muted rounded-full text-muted-foreground text-xs font-mono uppercase">
                        Incoming Request
                    </div>
                     <ArrowDown className="h-4 w-4 text-muted-foreground" />
                </div>

                {mockPipeline.map((mw, index) => (
                    <div key={mw.id} className="flex flex-col items-center w-full space-y-2">
                        <div className="flex items-center w-full p-4 border rounded-lg bg-card shadow-sm hover:shadow-md transition-shadow group cursor-grab active:cursor-grabbing">
                            <GripVertical className="h-5 w-5 text-muted-foreground mr-4 cursor-grab" />
                            <div className="flex-1">
                                <div className="font-medium flex items-center gap-2">
                                    {mw.name}
                                    <Badge variant="outline" className="text-xs font-normal">{mw.type}</Badge>
                                </div>
                            </div>
                            <div className="flex items-center gap-2">
                                <Badge variant={mw.status === "Active" ? "default" : "secondary"}>
                                    {mw.status}
                                </Badge>
                                <Button variant="ghost" size="sm">Edit</Button>
                            </div>
                        </div>
                         {index < mockPipeline.length && (
                            <ArrowDown className="h-4 w-4 text-muted-foreground" />
                        )}
                    </div>
                ))}
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  )
}
