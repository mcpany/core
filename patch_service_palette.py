import re

with open('ui/src/components/stacks/service-palette.tsx', 'r') as f:
    content = f.read()

# Add import yaml from 'js-yaml'
if "import yaml" not in content:
    content = content.replace('import { apiClient, ServiceTemplate } from "@/lib/client";', "import { apiClient, ServiceTemplate } from \"@/lib/client\";\nimport yaml from 'js-yaml';")

# Replace generateYamlSnippet
old_func = """    const generateYamlSnippet = (t: ServiceTemplate): string => {
        // Construct a YAML snippet based on the template config
        // This is a simplified generation.
        let snippet = `  - name: ${t.serviceConfig.name || t.name.toLowerCase().replace(/\\s+/g, '-')}\\n`;

        if (t.serviceConfig.commandLineService) {
            snippet += `    command: ${t.serviceConfig.commandLineService.command}\\n`;
            if (t.serviceConfig.commandLineService.workingDirectory) {
                snippet += `    working_dir: ${t.serviceConfig.commandLineService.workingDirectory}\\n`;
            }
            if (t.serviceConfig.commandLineService.env && Object.keys(t.serviceConfig.commandLineService.env).length > 0) {
                snippet += `    environment:\\n`;
                for (const [k, v] of Object.entries(t.serviceConfig.commandLineService.env)) {
                     // Handle EnvVarValue or string? Client type says string map usually for simple config,
                     // but UpstreamServiceConfig uses EnvVarValue?
                     // client.ts: environment: { [key: string]: string }; in commandLineService mapping.
                     // wait, client.ts mapping:
                     // environment: config.commandLineService.env (which is map<string, string>)
                     snippet += `      ${k}: ${v}\\n`;
                }
            }
        } else if (t.serviceConfig.httpService) {
             snippet += `    url: ${t.serviceConfig.httpService.address}\\n`;
        }

        return snippet;
    };"""

new_func = """    const generateYamlSnippet = (t: ServiceTemplate): string => {
        // Use proper YAML marshaling instead of manual concatenation
        const baseName = t.serviceConfig.name || t.name.toLowerCase().replace(/\\s+/g, '-');

        let serviceObj: any = { name: baseName };

        if (t.serviceConfig.commandLineService) {
            serviceObj.command = t.serviceConfig.commandLineService.command;
            if (t.serviceConfig.commandLineService.workingDirectory) {
                serviceObj.working_dir = t.serviceConfig.commandLineService.workingDirectory;
            }
            if (t.serviceConfig.commandLineService.env && Object.keys(t.serviceConfig.commandLineService.env).length > 0) {
                serviceObj.environment = {};
                for (const [k, v] of Object.entries(t.serviceConfig.commandLineService.env)) {
                     serviceObj.environment[k] = v;
                }
            }
        } else if (t.serviceConfig.httpService) {
             serviceObj.url = t.serviceConfig.httpService.address;
        }

        // Return formatted as an array item string but with proper 2-space indentation
        const yamlStr = yaml.dump([serviceObj], { indent: 2 });
        // The editor expects it without the leading array dash to match the list level,
        // wait, if we dump an array, it has "- name: ...". The old one had "  - name: ...".
        // Let's just indent each line by 2 spaces.
        return yamlStr.split('\\n').map(line => line ? '  ' + line : line).join('\\n');
    };"""

content = content.replace(old_func, new_func)

# Remove the TODO comment block
todo_comment = """                // TODO: proper YAML marshaling. For now, we might rely on the `description` or `name` to pick a snippet
                // if we want to match the old behavior, OR we simply serialize the config.
                // But the Stack Editor expects a YAML snippet to insert into the stack config.
                // Stack config is YAML."""
content = content.replace(todo_comment, "")

with open('ui/src/components/stacks/service-palette.tsx', 'w') as f:
    f.write(content)
