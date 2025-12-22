"use client"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { Plus, Terminal, Trash2 } from "lucide-react"

export default function ProfilesPage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Profiles</h1>
        <p className="text-muted-foreground">Manage execution profiles for different environments.</p>
      </div>

      <Tabs defaultValue="dev" className="w-full">
        <TabsList>
          <TabsTrigger value="dev">Development</TabsTrigger>
          <TabsTrigger value="staging">Staging</TabsTrigger>
          <TabsTrigger value="prod">Production</TabsTrigger>
          <TabsTrigger value="new"><Plus className="mr-2 h-4 w-4" /> New Profile</TabsTrigger>
        </TabsList>
        <TabsContent value="dev" className="space-y-4">
             <Card>
                <CardHeader>
                    <CardTitle>Development Profile</CardTitle>
                    <CardDescription>
                        Configuration for local development environment.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="flex items-center justify-between">
                         <div className="space-y-0.5">
                            <Label className="text-base">Debug Mode</Label>
                            <p className="text-sm text-muted-foreground">
                                Enable verbose logging and debug features.
                            </p>
                        </div>
                         <Switch checked={true} />
                    </div>
                    <Separator />
                    <div className="space-y-2">
                        <Label>Environment Variables</Label>
                        <div className="grid gap-2">
                             <div className="flex items-center gap-2">
                                <Input value="LOG_LEVEL" readOnly className="w-[200px]" />
                                <Input value="DEBUG" readOnly />
                             </div>
                             <div className="flex items-center gap-2">
                                <Input value="API_ENDPOINT" readOnly className="w-[200px]" />
                                <Input value="http://localhost:8080" readOnly />
                             </div>
                        </div>
                        <Button variant="outline" size="sm" className="mt-2">
                            <Plus className="mr-2 h-4 w-4" /> Add Variable
                        </Button>
                    </div>
                </CardContent>
                <CardFooter className="justify-between">
                    <Button variant="destructive" size="sm">
                        <Trash2 className="mr-2 h-4 w-4" /> Delete Profile
                    </Button>
                    <Button>Save Changes</Button>
                </CardFooter>
             </Card>
        </TabsContent>
         <TabsContent value="staging">
             <Card>
                 <CardHeader>
                    <CardTitle>Staging Profile</CardTitle>
                 </CardHeader>
                 <CardContent>
                     <p className="text-muted-foreground">Staging configuration...</p>
                 </CardContent>
             </Card>
        </TabsContent>
         <TabsContent value="prod">
             <Card>
                 <CardHeader>
                    <CardTitle>Production Profile</CardTitle>
                 </CardHeader>
                 <CardContent>
                     <p className="text-muted-foreground">Production configuration...</p>
                 </CardContent>
             </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
