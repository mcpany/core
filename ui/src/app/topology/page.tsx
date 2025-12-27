/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


'use client';

import React, { useState } from 'react';
import { NetworkGraph } from '@/components/topology/network-graph';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Shield, Plus, AlertCircle, CheckCircle2, Activity } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

export default function TopologyPage() {
    const [selectedNode, setSelectedNode] = useState<any>(null);
    const [policyDialogOpen, setPolicyDialogOpen] = useState(false);

    // Policy Wizard State
    const [step, setStep] = useState(1);
    const [policyName, setPolicyName] = useState('');
    const [action, setAction] = useState('allow');

    const handleNodeClick = (e: React.MouseEvent, node: any) => {
        setSelectedNode(node);
    };

    const resetWizard = () => {
        setStep(1);
        setPolicyName('');
        setAction('allow');
    };

    return (
        <div className="flex flex-col h-screen p-6 gap-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Network Topology</h1>
                    <p className="text-muted-foreground">Real-time traffic flow, policy enforcement, and infrastructure health.</p>
                </div>
                <div className="flex gap-3">
                     <Dialog open={policyDialogOpen} onOpenChange={(open) => { setPolicyDialogOpen(open); if(!open) resetWizard(); }}>
                        <DialogTrigger asChild>
                             <Button className="gap-2">
                                <Plus className="w-4 h-4" />
                                Create Policy
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[500px]">
                            <DialogHeader>
                                <DialogTitle>Create Network Policy</DialogTitle>
                                <DialogDescription>
                                    Define traffic rules for your MCP resources step-by-step.
                                </DialogDescription>
                            </DialogHeader>

                            <div className="py-4">
                                {/* Step Indicator */}
                                <div className="flex items-center gap-2 mb-6">
                                    {[1, 2, 3].map((s) => (
                                        <div key={s} className={`
                                            flex items-center justify-center w-8 h-8 rounded-full text-xs font-bold transition-colors
                                            ${step === s ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}
                                            ${step > s ? 'bg-primary/20 text-primary' : ''}
                                        `}>
                                            {step > s ? <CheckCircle2 className="w-4 h-4" /> : s}
                                        </div>
                                    ))}
                                    <div className="h-1 flex-1 bg-muted rounded-full">
                                        <div className="h-full bg-primary rounded-full transition-all duration-300" style={{ width: `${((step-1)/2)*100}%` }} />
                                    </div>
                                </div>

                                {step === 1 && (
                                    <div className="space-y-4 animate-in fade-in slide-in-from-right-4">
                                        <div className="space-y-2">
                                            <Label>Policy Name</Label>
                                            <Input
                                                data-testid="policy-name-input"
                                                placeholder="e.g. Block External Weather Access"
                                                value={policyName}
                                                onChange={(e) => setPolicyName(e.target.value)}
                                            />
                                        </div>
                                        <div className="space-y-2">
                                            <Label>Action</Label>
                                            <RadioGroup value={action} onValueChange={setAction} className="grid grid-cols-2 gap-4">
                                                <div>
                                                    <RadioGroupItem value="allow" id="allow" className="peer sr-only" />
                                                    <Label
                                                        htmlFor="allow"
                                                        className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-emerald-500 peer-data-[state=checked]:text-emerald-500 [&:has([data-state=checked])]:border-primary"
                                                    >
                                                        <CheckCircle2 className="mb-3 h-6 w-6" />
                                                        Allow
                                                    </Label>
                                                </div>
                                                <div>
                                                    <RadioGroupItem value="deny" id="deny" className="peer sr-only" />
                                                    <Label
                                                        htmlFor="deny"
                                                        className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-red-500 peer-data-[state=checked]:text-red-500 [&:has([data-state=checked])]:border-primary"
                                                    >
                                                        <Shield className="mb-3 h-6 w-6" />
                                                        Deny
                                                    </Label>
                                                </div>
                                            </RadioGroup>
                                        </div>
                                    </div>
                                )}

                                {step === 2 && (
                                    <div className="space-y-4 animate-in fade-in slide-in-from-right-4">
                                         <div className="space-y-2">
                                            <Label>Source</Label>
                                            <Select defaultValue="any">
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select source" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="any">Any Source</SelectItem>
                                                    <SelectItem value="client">Client (User)</SelectItem>
                                                    <SelectItem value="service">Specific Service</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                         <div className="space-y-2">
                                            <Label>Destination</Label>
                                            <Select defaultValue="weather">
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select destination" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="weather">Weather Service</SelectItem>
                                                    <SelectItem value="payments">Payment Service</SelectItem>
                                                    <SelectItem value="all">All Services</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    </div>
                                )}

                                {step === 3 && (
                                    <div className="space-y-4 animate-in fade-in slide-in-from-right-4">
                                        <div className="rounded-lg border p-4 bg-muted/50">
                                            <h4 className="font-medium mb-2">Summary</h4>
                                            <div className="grid grid-cols-2 gap-2 text-sm">
                                                <span className="text-muted-foreground">Name:</span>
                                                <span>{policyName || 'Untitled Policy'}</span>
                                                <span className="text-muted-foreground">Action:</span>
                                                <Badge variant={action === 'allow' ? 'default' : 'destructive'}>{action.toUpperCase()}</Badge>
                                                <span className="text-muted-foreground">Traffic:</span>
                                                <span>Any &rarr; Weather Service</span>
                                            </div>
                                        </div>
                                        <AlertCircle className="w-12 h-12 text-muted-foreground mx-auto opacity-20" />
                                        <p className="text-center text-sm text-muted-foreground">
                                            This policy will be applied immediately to the Firewall configuration.
                                        </p>
                                    </div>
                                )}
                            </div>

                            <DialogFooter>
                                {step > 1 && (
                                    <Button variant="outline" onClick={() => setStep(step - 1)}>Back</Button>
                                )}
                                {step < 3 ? (
                                    <Button data-testid="next-button" onClick={() => setStep(step + 1)} disabled={step === 1 && !policyName}>Next</Button>
                                ) : (
                                    <Button onClick={() => setPolicyDialogOpen(false)}>Create Policy</Button>
                                )}
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            <div className="flex-1 flex gap-4 overflow-hidden relative">
                {/* Graph Container */}
                <div className="flex-1 h-full shadow-inner rounded-xl border bg-muted/20 relative">
                    <NetworkGraph onNodeClick={handleNodeClick} />
                    {/* Graph Controls overlay or legend could go here */}
                </div>

                {/* Details Sidebar - Responsive: Overlay on small, Side on large */}
                <div className={`
                    absolute lg:static top-0 right-0 h-full w-80
                    bg-background/95 backdrop-blur-sm lg:bg-transparent
                    border-l lg:border-none shadow-xl lg:shadow-none z-10
                    transition-transform duration-300 ease-in-out
                    ${selectedNode ? 'translate-x-0' : 'translate-x-full lg:translate-x-0 lg:hidden'}
                `}>
                    {selectedNode ? (
                     <Card className="h-full overflow-y-auto flex flex-col border-l">
                        <CardHeader className="flex-shrink-0 p-4 pb-2">
                             <div className="flex justify-between items-start">
                                <div>
                                    <CardTitle className="flex items-center gap-2 text-lg">
                                        {selectedNode.data.label}
                                        {selectedNode.data.status && (
                                            <div className={`w-2 h-2 rounded-full ${selectedNode.data.status === 'active' ? 'bg-emerald-500' : 'bg-amber-500'}`} />
                                        )}
                                    </CardTitle>
                                    <CardDescription className="text-xs mt-1">
                                        {selectedNode.type.toUpperCase()} Node
                                    </CardDescription>
                                </div>
                                <Button variant="ghost" size="icon" className="h-6 w-6 lg:hidden" onClick={() => setSelectedNode(null)}>
                                     <span className="sr-only">Close</span>
                                     <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-x w-4 h-4"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                                </Button>
                             </div>
                        </CardHeader>
                        <CardContent className="space-y-6 p-4 pt-2">
                            {selectedNode.data.metrics && (
                                <div className="space-y-3">
                                    <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                                        <Activity className="w-3 h-3" /> Live Metrics
                                    </h4>
                                    <div className="grid grid-cols-1 gap-2">
                                        {Object.entries(selectedNode.data.metrics).map(([key, value]) => (
                                            <div key={key} className="flex justify-between items-center p-2 bg-muted/50 rounded-md border text-xs">
                                                <span className="capitalize text-muted-foreground">{key.replace(/([A-Z])/g, ' $1').trim()}</span>
                                                <span className="font-mono font-medium text-foreground">{String(value)}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}

                            <div>
                                <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2 flex items-center gap-2">
                                     <Shield className="w-3 h-3" /> Applied Policies
                                </h4>
                                <div className="space-y-2">
                                    <Badge variant="outline" className="w-full justify-start p-2 gap-2 cursor-pointer hover:bg-muted/50 font-normal">
                                        <Shield className="w-3 h-3 text-emerald-500" />
                                        <span>Default Allow</span>
                                    </Badge>
                                    <Badge variant="outline" className="w-full justify-start p-2 gap-2 cursor-pointer hover:bg-muted/50 font-normal">
                                        <Shield className="w-3 h-3 text-blue-500" />
                                        <span>Rate Limit: 500/s</span>
                                    </Badge>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                    ) : (
                        // Empty state for large screens only
                        <div className="hidden lg:flex h-full border-l bg-muted/5 items-center justify-center text-muted-foreground p-6 text-center">
                             <div className="space-y-2">
                                 <Activity className="w-12 h-12 mx-auto opacity-10" />
                                 <p className="text-sm">Select a node to view details</p>
                             </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
