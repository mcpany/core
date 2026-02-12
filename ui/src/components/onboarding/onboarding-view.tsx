/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { RegisterServiceDialog } from "@/components/register-service-dialog";
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ArrowRight, Github, FolderOpen, Globe, Zap, LayoutDashboard, Search } from "lucide-react";
import Link from "next/link";

interface OnboardingViewProps {
    onServiceRegistered: () => void;
}

export function OnboardingView({ onServiceRegistered }: OnboardingViewProps) {
    return (
        <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center p-8 bg-gradient-to-b from-background to-muted/20 animate-in fade-in duration-700">
            <div className="max-w-4xl w-full space-y-12">
                <div className="text-center space-y-4">
                    <div className="inline-flex items-center justify-center p-3 rounded-2xl bg-primary/10 mb-4 ring-1 ring-primary/20 shadow-lg shadow-primary/5">
                        <Zap className="h-8 w-8 text-primary" />
                    </div>
                    <h1 className="text-4xl md:text-6xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
                        Welcome to MCP Any
                    </h1>
                    <p className="text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
                        The definitive management console for the Model Context Protocol.
                        <br />
                        Connect your first service to get started.
                    </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <Card className="group hover:border-primary/50 transition-all duration-300 hover:shadow-xl hover:-translate-y-1 bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <div className="h-12 w-12 rounded-lg bg-blue-500/10 flex items-center justify-center mb-4 group-hover:bg-blue-500/20 transition-colors">
                                <FolderOpen className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                            </div>
                            <CardTitle>Connect Filesystem</CardTitle>
                            <CardDescription>
                                Expose local files and directories to your AI agents securely.
                            </CardDescription>
                        </CardHeader>
                        <CardFooter>
                            <RegisterServiceDialog
                                initialTemplateId="filesystem"
                                onSuccess={onServiceRegistered}
                                trigger={
                                    <Button className="w-full group/btn">
                                        Start Setup <ArrowRight className="ml-2 h-4 w-4 group-hover/btn:translate-x-1 transition-transform" />
                                    </Button>
                                }
                            />
                        </CardFooter>
                    </Card>

                    <Card className="group hover:border-primary/50 transition-all duration-300 hover:shadow-xl hover:-translate-y-1 bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <div className="h-12 w-12 rounded-lg bg-purple-500/10 flex items-center justify-center mb-4 group-hover:bg-purple-500/20 transition-colors">
                                <Github className="h-6 w-6 text-purple-600 dark:text-purple-400" />
                            </div>
                            <CardTitle>Connect GitHub</CardTitle>
                            <CardDescription>
                                Give your agent access to repositories, issues, and pull requests.
                            </CardDescription>
                        </CardHeader>
                        <CardFooter>
                            <RegisterServiceDialog
                                initialTemplateId="github"
                                onSuccess={onServiceRegistered}
                                trigger={
                                    <Button className="w-full group/btn" variant="secondary">
                                        Start Setup <ArrowRight className="ml-2 h-4 w-4 group-hover/btn:translate-x-1 transition-transform" />
                                    </Button>
                                }
                            />
                        </CardFooter>
                    </Card>

                    <Card className="group hover:border-primary/50 transition-all duration-300 hover:shadow-xl hover:-translate-y-1 bg-card/50 backdrop-blur-sm">
                        <CardHeader>
                            <div className="h-12 w-12 rounded-lg bg-orange-500/10 flex items-center justify-center mb-4 group-hover:bg-orange-500/20 transition-colors">
                                <Globe className="h-6 w-6 text-orange-600 dark:text-orange-400" />
                            </div>
                            <CardTitle>Web Search</CardTitle>
                            <CardDescription>
                                Enable internet access for your agents using Brave Search.
                            </CardDescription>
                        </CardHeader>
                        <CardFooter>
                            <RegisterServiceDialog
                                initialTemplateId="web-search"
                                onSuccess={onServiceRegistered}
                                trigger={
                                    <Button className="w-full group/btn" variant="secondary">
                                        Start Setup <ArrowRight className="ml-2 h-4 w-4 group-hover/btn:translate-x-1 transition-transform" />
                                    </Button>
                                }
                            />
                        </CardFooter>
                    </Card>
                </div>

                <div className="flex flex-col items-center gap-6 pt-8">
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-2">
                            <LayoutDashboard className="h-4 w-4" />
                            <span>Enterprise Grade</span>
                        </span>
                        <span className="w-1 h-1 rounded-full bg-border" />
                        <span className="flex items-center gap-2">
                            <Search className="h-4 w-4" />
                            <span>Full Observability</span>
                        </span>
                    </div>

                    <div className="flex gap-4">
                        <Link href="/marketplace">
                            <Button variant="link" className="text-muted-foreground hover:text-foreground">
                                Browse Marketplace
                            </Button>
                        </Link>
                        <span className="text-muted-foreground py-2">•</span>
                        <RegisterServiceDialog
                            onSuccess={onServiceRegistered}
                            trigger={
                                <Button variant="link" className="text-muted-foreground hover:text-foreground">
                                    Manual Configuration
                                </Button>
                            }
                        />
                    </div>
                </div>
            </div>
        </div>
    );
}
