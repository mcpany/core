"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"

export default function SettingsPage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">General platform configuration.</p>
      </div>

      <div className="grid gap-6">
          <Card>
              <CardHeader>
                  <CardTitle>Global Configuration</CardTitle>
                  <CardDescription>Server-wide settings applied to all services.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                  <div className="grid gap-2">
                      <Label htmlFor="listen-addr">Listen Address</Label>
                      <Input id="listen-addr" defaultValue=":8080" />
                  </div>
                   <div className="grid gap-2">
                      <Label htmlFor="log-level">Log Level</Label>
                      <Input id="log-level" defaultValue="INFO" />
                  </div>
                  <Button>Save Changes</Button>
              </CardContent>
          </Card>
      </div>
    </div>
  )
}
