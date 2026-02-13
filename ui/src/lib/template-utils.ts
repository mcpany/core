/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceTemplate as ClientServiceTemplate, UpstreamServiceConfig } from "@/lib/client";
import { ServiceTemplate as FrontendServiceTemplate } from "@/lib/templates";
import {
    Activity,
    Calendar,
    CheckCircle2,
    Clock,
    Cloud,
    Database,
    FileText,
    Github,
    Globe,
    Map,
    MessageSquare,
    Server,
    Zap
} from "lucide-react";

/**
 * Applies user-provided field values to a service template's configuration.
 *
 * @param template The service template.
 * @param fieldValues A map of field names to their values.
 * @returns A new UpstreamServiceConfig object with the values applied.
 */
export function applyTemplateFields(
  template: FrontendServiceTemplate,
  fieldValues: Record<string, string>
): Partial<UpstreamServiceConfig> {
  // Deep clone the config to avoid mutating the original
  const config = JSON.parse(JSON.stringify(template.config));

  if (!template.fields) {
    return config;
  }

  for (const field of template.fields) {
    const value = fieldValues[field.name];
    if (value === undefined || value === "") {
        continue; // Skip empty values, or handle defaults if we had them
    }

    applyValueToConfig(config, field.key, value, field.replaceToken);
  }

  return config;
}

/**
 * Helper to set a value in a nested object based on a dot-notation path.
 * If replaceToken is provided, it performs a string replacement on the target value.
 */
function applyValueToConfig(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: Record<string, any>,
  path: string,
  value: string,
  replaceToken?: string
) {
  const parts = path.split(".");
  let current = obj;

  for (let i = 0; i < parts.length - 1; i++) {
    const part = parts[i];
    if (!(part in current)) {
      // Create nested object if it doesn't exist
      current[part] = {};
    }
    current = current[part];
  }

  const lastPart = parts[parts.length - 1];

  if (replaceToken) {
    const currentValue = current[lastPart];
    if (typeof currentValue === "string") {
      current[lastPart] = currentValue.replace(replaceToken, value);
    } else {
        // Fallback: if target isn't a string, just set it (shouldn't happen for replaceToken use cases)
        current[lastPart] = value;
    }
  } else {
    // Direct assignment
    current[lastPart] = value;
  }
}

export function getIconForName(name: string) {
    const map: Record<string, any> = {
        "activity": Activity,
        "calendar": Calendar,
        "check-circle-2": CheckCircle2,
        "clock": Clock,
        "cloud": Cloud,
        "database": Database,
        "file-text": FileText,
        "github": Github,
        "globe": Globe,
        "map": Map,
        "message-square": MessageSquare,
        "server": Server,
        "zap": Zap,
        // Aliases
        "linear": CheckCircle2,
        "jira": Activity,
        "gitlab": Github,
        "slack": MessageSquare,
        "notion": FileText,
    };
    return map[name.toLowerCase()] || Server;
}

export function convertBackendTemplateToFrontend(backend: ClientServiceTemplate): FrontendServiceTemplate {
    const fields = parseSchemaToFields(backend.serviceConfig.configurationSchema);

    // Determine category based on tags
    let category = "Other";
    if (backend.tags?.includes("productivity")) category = "Productivity";
    else if (backend.tags?.includes("development")) category = "Dev Tools";
    else if (backend.tags?.includes("database")) category = "Database";
    else if (backend.tags?.includes("cloud")) category = "Cloud";
    else if (backend.tags?.includes("web")) category = "Web";

    return {
        id: backend.id,
        name: backend.name,
        description: backend.description,
        icon: getIconForName(backend.icon || "server"),
        category: category,
        config: backend.serviceConfig,
        fields: fields
    };
}

function parseSchemaToFields(schemaStr?: string): FrontendServiceTemplate["fields"] {
    if (!schemaStr) return [];
    try {
        const schema = JSON.parse(schemaStr);
        if (schema.type !== "object" || !schema.properties) return [];

        const fields: NonNullable<FrontendServiceTemplate["fields"]> = [];
        for (const [key, prop] of Object.entries(schema.properties as Record<string, any>)) {
            // Heuristic mapping: Map schema property to deep config path.
            // Since we don't have the mapping in the schema, we rely on convention or assumptions?
            // The seeds.go templates used environment variables usually.
            // e.g. GITHUB_PERSONAL_ACCESS_TOKEN -> commandLineService.env.GITHUB_PERSONAL_ACCESS_TOKEN
            // But HTTP services might use upstreamAuth.oauth2.clientId etc.

            // For OAuth (implicit config), we might not need fields if we rely on "Connect" button.
            // But if there are fields (like in CLI templates), we assume they map to Env Vars if capitalized?
            // Or we check the `upstreamAuth` config?

            // Let's assume for now:
            // 1. If key looks like ENV_VAR, map to commandLineService.env.KEY
            // 2. If schema title says "Client ID", maybe upstreamAuth.oauth2.clientId?

            // Simplification: We only support generating fields for CLI templates that use Env Vars for now,
            // OR if the template manually defined the mapping (which we don't have in backend proto yet).

            // Wait, the seeds.go I wrote has "ConfigurationSchema".
            // It has properties like GITHUB_PERSONAL_ACCESS_TOKEN.
            // And the command uses `npx ...`.
            // So these map to `commandLineService.env`.

            // For OAuth templates (Google Calendar), the schema was empty!
            // So fields will be empty. The "Connect" logic (OAuth) is handled separately by the UI based on authType?
            // `FrontendServiceTemplate` doesn't have `authType`.

            // The `ServiceList` component handles "Login" button.
            // `ServiceEditor` or `TemplateConfigForm` handles fields.

            let configKey = "";
            if (/^[A-Z0-9_]+$/.test(key)) {
                configKey = `commandLineService.env.${key}`;
            } else {
                // Fallback or skip
                continue;
            }

            fields.push({
                name: key,
                label: prop.title || key,
                placeholder: prop.description || "",
                key: configKey,
                type: prop.format === "password" ? "password" : "text"
            });
        }
        return fields;
    } catch (e) {
        console.error("Failed to parse schema", e);
        return [];
    }
}
