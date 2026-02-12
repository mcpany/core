/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceTemplate } from "@/lib/templates";
import { ServiceRegistryItem } from "@/lib/service-registry";
import { CommunityServer } from "@/lib/marketplace-service";
import {
    Database, FileText, Github, Globe, Server, Activity, Cloud,
    MessageSquare, Map as MapIcon, Clock, Zap, CheckCircle2, Box, Command
} from "lucide-react";

/**
 * Maps a ServiceRegistryItem (static structured definition) to a ServiceTemplate.
 * This converts the JSON schema properties into wizard fields.
 */
export function mapRegistryToTemplate(item: ServiceRegistryItem): ServiceTemplate {
    const fields: any[] = [];

    // Parse configurationSchema to generate fields
    if (item.configurationSchema && item.configurationSchema.properties) {
        Object.entries(item.configurationSchema.properties).forEach(([key, prop]: [string, any]) => {
            fields.push({
                name: key,
                label: prop.title || key,
                placeholder: prop.description || "",
                // Heuristic: Map uppercase keys to env vars, others to command args if we had a way to specify
                // For registry items, they usually expect env vars or command args replacers.
                // The registry items currently use "env" in their command (e.g. npx ...).
                // But the schema implies inputs.
                // Let's assume standard Registry items map to Environment Variables by default
                // unless the command has a placeholder.
                key: `commandLineService.env.${key}`,
                defaultValue: prop.default,
                type: prop.format === "password" ? "password" : "text"
            });
        });
    }

    // Check for command placeholders (e.g. {{DB_PATH}})
    // If a field matches a placeholder, update its key to replace the token in command
    fields.forEach(field => {
        const token = `{{${field.name}}}`;
        if (item.command.includes(token)) {
            field.key = "commandLineService.command";
            field.replaceToken = token;
        }
    });

    return {
        id: `registry-${item.id}`,
        name: item.name,
        description: item.description,
        icon: resolveIcon(item.id, "registry"),
        category: "Registry",
        featured: false,
        config: {
            name: item.id, // default service name
            commandLineService: {
                command: item.command,
                env: {},
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            } as any
        },
        fields: fields,
        // @ts-ignore - Adding custom property for UI badge
        source: "verified"
    };
}

/**
 * Maps a CommunityServer (dynamic markdown parse) to a ServiceTemplate.
 * Since we don't have a schema, we provide a generic "Environment Variables" field
 * or just a raw command edit.
 */
export function mapCommunityToTemplate(server: CommunityServer): ServiceTemplate {
    // Generate a safe ID
    const safeId = server.name.toLowerCase().replace(/[^a-z0-9-]/g, '-');

    // Heuristic for command
    let command = "npx -y package-name";
    // Try to guess from URL
    if (server.url.includes("github.com")) {
        const parts = server.url.split("/");
        if (parts.length >= 5) {
            const repo = parts[4];
            // If python tag exists?
            if (server.tags.some(t => t.includes("python") || t.includes("🐍"))) {
                command = `uvx ${repo}`; // Assumption
            } else {
                command = `npx -y ${repo}`; // Assumption
            }
        }
    }

    return {
        id: `community-${safeId}`,
        name: server.name,
        description: server.description,
        icon: resolveIcon(server.category, "community"),
        category: server.category || "Community",
        featured: false,
        config: {
            name: safeId,
            commandLineService: {
                command: command,
                env: {},
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            } as any
        },
        fields: [
            // Generic field to let user edit command if needed,
            // but usually we want specific env vars.
            // Since we don't know them, we might just leave fields empty
            // and let the user configure raw JSON in the wizard step if we supported it?
            // Or just provide a "Command" field override?
            // The Wizard form supports field overriding specific config paths.
            {
                name: "command",
                label: "Command",
                placeholder: "npx -y ...",
                key: "commandLineService.command",
                defaultValue: command
            }
        ],
        // @ts-ignore - Adding custom property for UI badge
        source: "community",
        // @ts-ignore
        url: server.url
    };
}

/**
 * Resolves a string name/category to a Lucide icon component.
 */
export function resolveIcon(key: string, source: string): any {
    const k = key.toLowerCase();

    if (k.includes("database") || k.includes("sql") || k.includes("postgres")) return Database;
    if (k.includes("git") || k.includes("code")) return Github;
    if (k.includes("web") || k.includes("search") || k.includes("browser")) return Globe;
    if (k.includes("cloud") || k.includes("aws") || k.includes("flare")) return Cloud;
    if (k.includes("slack") || k.includes("chat") || k.includes("discord")) return MessageSquare;
    if (k.includes("file") || k.includes("system")) return FileText;
    if (k.includes("map") || k.includes("geo")) return MapIcon;
    if (k.includes("time") || k.includes("date")) return Clock;
    if (k.includes("weather")) return Cloud;

    if (source === "community") return Box;
    if (source === "registry") return CheckCircle2;

    return Command;
}
