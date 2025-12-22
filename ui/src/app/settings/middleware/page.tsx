"use client"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import { ArrowDown, GripVertical, Settings2, Shield, Zap } from "lucide-react"

export default function MiddlewarePage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Middleware</h1>
        <p className="text-muted-foreground">Manage the request processing pipeline.</p>
      </div>

      <div className="flex flex-col gap-4 max-w-3xl">
         {/* Pipeline Visualization */}
         <Card className="border-dashed bg-muted/20">
             <CardHeader className="pb-4">
                 <CardTitle className="text-sm font-medium uppercase text-muted-foreground tracking-wider">Request Pipeline</CardTitle>
             </CardHeader>
             <CardContent className="flex flex-col items-center gap-0">
                  <div className="w-full p-4 border rounded-md bg-background flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-3">
                             <GripVertical className="h-5 w-5 text-muted-foreground cursor-move" />
                             <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-md">
                                 <Shield className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                             </div>
                             <div>
                                 <h3 className="font-medium">Authentication</h3>
                                 <p className="text-xs text-muted-foreground">Validates API keys and tokens</p>
                             </div>
                        </div>
                        <Switch checked={true} />
                  </div>
                  <ArrowDown className="h-6 w-6 text-muted-foreground/50 my-1" />
                   <div className="w-full p-4 border rounded-md bg-background flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-3">
                             <GripVertical className="h-5 w-5 text-muted-foreground cursor-move" />
                             <div className="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-md">
                                 <Zap className="h-5 w-5 text-orange-600 dark:text-orange-400" />
                             </div>
                             <div>
                                 <h3 className="font-medium">Rate Limiting</h3>
                                 <p className="text-xs text-muted-foreground">Protects against abuse</p>
                             </div>
                        </div>
                        <Switch checked={true} />
                  </div>
                  <ArrowDown className="h-6 w-6 text-muted-foreground/50 my-1" />
                   <div className="w-full p-4 border rounded-md bg-background flex items-center justify-between shadow-sm">
                        <div className="flex items-center gap-3">
                             <GripVertical className="h-5 w-5 text-muted-foreground cursor-move" />
                             <div className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-md">
                                 <Settings2 className="h-5 w-5 text-purple-600 dark:text-purple-400" />
                             </div>
                             <div>
                                 <h3 className="font-medium">Request Transformation</h3>
                                 <p className="text-xs text-muted-foreground">Modifies headers and payload</p>
                             </div>
                        </div>
                        <Switch checked={false} />
                  </div>
                  <ArrowDown className="h-6 w-6 text-muted-foreground/50 my-1" />
                   <div className="w-full p-4 border border-dashed rounded-md bg-muted/40 flex items-center justify-center h-16 text-muted-foreground text-sm">
                        Backend Service
                  </div>
             </CardContent>
         </Card>
      </div>
    </div>
  )
}
