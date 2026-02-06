/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Github, KeyRound, ExternalLink } from "lucide-react";

interface GitHubTokenFieldProps {
  id: string;
  value: string;
  onChange: (value: string) => void;
  title?: string;
  description?: string;
  required?: boolean;
}

/**
 * GitHubTokenField component.
 * Renders a specialized input for GitHub Personal Access Tokens with a generator button.
 */
export function GitHubTokenField({ id, value, onChange, title, description, required }: GitHubTokenFieldProps) {
  const generateUrl = "https://github.com/settings/tokens/new?scopes=repo,user&description=MCP+Any+Server";

  return (
    <div className="grid gap-2 p-4 border rounded-md bg-muted/30">
        <div className="flex items-center justify-between">
            <Label htmlFor={id} className="flex items-center gap-1 font-semibold">
                <Github className="h-4 w-4" />
                {title || "GitHub Personal Access Token"}
                {required && <span className="text-destructive">*</span>}
            </Label>
        </div>

        <div className="flex gap-2">
            <div className="relative flex-1">
                <KeyRound className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    id={id}
                    value={value}
                    onChange={(e) => onChange(e.target.value)}
                    placeholder="ghp_..."
                    type="password"
                    className="pl-9 font-mono"
                />
            </div>
            <Button variant="secondary" onClick={() => window.open(generateUrl, '_blank')} type="button">
                Generate Token <ExternalLink className="ml-2 h-3 w-3" />
            </Button>
        </div>

        <p className="text-xs text-muted-foreground">
            {description || "A Personal Access Token (PAT) is required to access your private repositories and user data."}
            {" "}
            The token will be stored securely in the local configuration.
        </p>
    </div>
  );
}
