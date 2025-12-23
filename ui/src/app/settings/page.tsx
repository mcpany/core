
import { Sidebar } from "@/components/sidebar"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"

export default function SettingsPage() {
  return (
    <div className="flex min-h-screen w-full flex-col bg-muted/40 md:flex-row">
      <Sidebar />
      <div className="flex flex-col sm:gap-4 sm:py-4 sm:pl-14 w-full">
        <main className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
             <div className="flex items-center">
                <h1 className="text-lg font-semibold md:text-2xl">Settings</h1>
            </div>
          <Tabs defaultValue="general" className="w-full">
            <TabsList>
              <TabsTrigger value="general">General</TabsTrigger>
              <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
              <TabsTrigger value="security">Security</TabsTrigger>
            </TabsList>
            <TabsContent value="general">
              <Card>
                <CardHeader>
                  <CardTitle>General Configuration</CardTitle>
                  <CardDescription>
                    Manage basic server settings.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-2">
                    <Label htmlFor="server-name">Server Name</Label>
                    <Input id="server-name" defaultValue="MCP Any Main" />
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch id="maintenance-mode" />
                    <Label htmlFor="maintenance-mode">Maintenance Mode</Label>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="webhooks">
              <Card>
                <CardHeader>
                  <CardTitle>Webhook Configuration</CardTitle>
                  <CardDescription>
                    Configure global webhooks for server events.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-2">
                    <Label htmlFor="webhook-url">Global Webhook URL</Label>
                    <Input id="webhook-url" placeholder="https://api.example.com/hooks/mcp" />
                  </div>
                   <div className="grid gap-2">
                    <Label htmlFor="webhook-secret">Webhook Secret</Label>
                    <Input id="webhook-secret" type="password" placeholder="whsec_..." />
                  </div>
                  <div className="grid gap-2">
                    <Label>Events to Trigger</Label>
                    <div className="flex items-center space-x-2">
                        <Switch id="event-service-up" defaultChecked />
                        <Label htmlFor="event-service-up">Service Up</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                        <Switch id="event-service-down" defaultChecked />
                        <Label htmlFor="event-service-down">Service Down</Label>
                    </div>
                     <div className="flex items-center space-x-2">
                        <Switch id="event-error" />
                        <Label htmlFor="event-error">Critical Errors</Label>
                    </div>
                  </div>
                   <Button>Save Webhook Settings</Button>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </main>
      </div>
    </div>
  )
}
