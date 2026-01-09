/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceDefinition, ResourceContent } from "@/lib/client";

export const MOCK_RESOURCES: ResourceDefinition[] = [
    { name: "config.json", uri: "file:///app/config.json", mimeType: "application/json" },
    { name: "README.md", uri: "file:///app/README.md", mimeType: "text/markdown" },
    { name: "script.py", uri: "file:///app/scripts/script.py", mimeType: "text/x-python" },
    { name: "logo.png", uri: "file:///app/assets/logo.png", mimeType: "image/png" },
    { name: "notes.txt", uri: "file:///home/user/notes.txt", mimeType: "text/plain" }
];

export const MOCK_RESOURCE_CONTENTS: Record<string, ResourceContent> = {
    "file:///app/config.json": {
        uri: "file:///app/config.json",
        mimeType: "application/json",
        text: JSON.stringify({
            "key": "value",
            "nested": {
                "data": [1, 2, 3]
            },
            "timestamp": new Date().toISOString()
        }, null, 2)
    },
    "file:///app/README.md": {
        uri: "file:///app/README.md",
        mimeType: "text/markdown",
        text: `# Demo Document\n\nThis is a *simulated* markdown file from **file:///app/README.md**.\n\n## Features\n- Rich text\n- Code blocks\n\n\`\`\`javascript\nconsole.log('Hello MCP');\n\`\`\``
    },
    "file:///app/scripts/script.py": {
        uri: "file:///app/scripts/script.py",
        mimeType: "text/x-python",
        text: `def hello_world():\n    print("Hello from file:///app/scripts/script.py")\n\nif __name__ == "__main__":\n    hello_world()`
    }
};

export function getMockResourceContent(uri: string): ResourceContent {
    if (MOCK_RESOURCE_CONTENTS[uri]) {
        return MOCK_RESOURCE_CONTENTS[uri];
    }
    return {
        uri: uri,
        mimeType: "text/plain",
        text: `This is a simulated plain text content for resource:\n${uri}\n\nThe backend endpoint /api/v1/resources/read is likely not implemented yet.`
    };
}
