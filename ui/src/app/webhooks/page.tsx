
"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Plus, Play, Trash2, Check, X } from "lucide-react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// Mock webhooks
const initialWebhooks = [
    { id: "wh-1", url: "https://webhook.site/uuid", events: ["service.up", "service.down"], status: "active", lastStatus: 200 },
    { id: "wh-2", url: "https://api.slack.com/webhook", events: ["alert.critical"], status: "active", lastStatus: 200 },
];

export default function WebhooksPage() {
    const [webhooks, setWebhooks] = useState(initialWebhooks);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [newWebhook, setNewWebhook] = useState({ url: "", events: "all" });

    const addWebhook = () => {
        setWebhooks([...webhooks, {
            id: `wh-${Date.now()}`,
            url: newWebhook.url,
            events: [newWebhook.events],
            status: "active",
            lastStatus: 0
        }]);
        setIsDialogOpen(false);
        setNewWebhook({ url: "", events: "all" });
    };

    const deleteWebhook = (id: string) => {
        setWebhooks(webhooks.filter(w => w.id !== id));
    };

    const testWebhook = (id: string) => {
        alert(`Sending test payload to webhook ${id}...`);
    };

    return (
        <div className="flex-1 space-y-4 p-8 pt-6">
            <div className="flex items-center justify-between">
                <h2 className="text-3xl font-bold tracking-tight">Webhooks</h2>
                <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                    <DialogTrigger asChild>
                        <Button>
                            <Plus className="mr-2 h-4 w-4" /> Add Webhook
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Add Webhook</DialogTitle>
                            <DialogDescription>
                                Receive notifications when specific events occur.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid grid-cols-4 items-center gap-4">
                                <Label htmlFor="url" className="text-right">Payload URL</Label>
                                <Input
                                    id="url"
                                    value={newWebhook.url}
                                    onChange={(e) => setNewWebhook({ ...newWebhook, url: e.target.value })}
                                    className="col-span-3"
                                    placeholder="https://..."
                                />
                            </div>
                            <div className="grid grid-cols-4 items-center gap-4">
                                <Label htmlFor="events" className="text-right">Events</Label>
                                <Select
                                    value={newWebhook.events}
                                    onValueChange={(val) => setNewWebhook({ ...newWebhook, events: val })}
                                >
                                    <SelectTrigger className="col-span-3">
                                        <SelectValue placeholder="Select events" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all">All Events</SelectItem>
                                        <SelectItem value="service.up">Service Up</SelectItem>
                                        <SelectItem value="service.down">Service Down</SelectItem>
                                        <SelectItem value="alert">Alerts</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                        <DialogFooter>
                            <Button onClick={addWebhook}>Add Webhook</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <Card className="backdrop-blur-sm bg-background/50">
                <CardHeader>
                    <CardTitle>Configured Webhooks</CardTitle>
                    <CardDescription>Manage your external notification endpoints.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>URL</TableHead>
                                <TableHead>Events</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Last Delivery</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {webhooks.map((webhook) => (
                                <TableRow key={webhook.id}>
                                    <TableCell className="font-mono text-xs">{webhook.url}</TableCell>
                                    <TableCell>
                                        <div className="flex gap-1">
                                            {webhook.events.map(e => <Badge key={e} variant="outline">{e}</Badge>)}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                         <Badge variant={webhook.status === 'active' ? 'default' : 'secondary'}>{webhook.status}</Badge>
                                    </TableCell>
                                    <TableCell>
                                        {webhook.lastStatus === 200 ? (
                                            <span className="flex items-center text-green-500 text-xs"><Check className="h-3 w-3 mr-1"/> Success</span>
                                        ) : webhook.lastStatus === 0 ? (
                                            <span className="text-muted-foreground text-xs">Never</span>
                                        ) : (
                                            <span className="flex items-center text-red-500 text-xs"><X className="h-3 w-3 mr-1"/> Failed</span>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <div className="flex justify-end gap-2">
                                            <Button variant="outline" size="icon" onClick={() => testWebhook(webhook.id)}>
                                                <Play className="h-4 w-4" />
                                            </Button>
                                            <Button variant="ghost" size="icon" className="text-red-600 hover:text-red-700 hover:bg-red-50" onClick={() => deleteWebhook(webhook.id)}>
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
