/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SecretValue } from "@proto/config/v1/auth";

/**
 * Result of analyzing a repository.
 */
export interface AnalysisResult {
  /** The detected or fallback name of the service. */
  name: string;
  /** The suggested run command. */
  command: string;
  /** Detected environment variables. */
  envVars: Record<string, SecretValue>;
  /** The type of project detected. */
  detectedType: "node" | "python" | "unknown";
  /** Description extracted from metadata. */
  description: string;
  /** Error message if analysis failed. */
  error?: string;
}

const RAW_GITHUB_BASE = "https://raw.githubusercontent.com";

interface RepoInfo {
    owner: string;
    repo: string;
    branch?: string;
    path: string; // Base path, empty for root
}

/**
 * Parses a GitHub URL to extract owner, repo, branch, and path.
 */
export function parseGithubUrl(url: string): RepoInfo | null {
  try {
    const u = new URL(url);
    if (u.hostname !== "github.com") return null;

    // Path: /owner/repo/tree/branch/path/to/dir
    const parts = u.pathname.split("/").filter(Boolean);
    if (parts.length < 2) return null;

    const owner = parts[0];
    const repo = parts[1];

    let branch: string | undefined = undefined;
    let path = "";

    if (parts.length >= 4 && parts[2] === "tree") {
        branch = parts[3];
        if (parts.length > 4) {
            path = parts.slice(4).join("/");
        }
    }

    return { owner, repo, branch, path };
  } catch {
    return null;
  }
}

/**
 * Fetches raw content from a GitHub repository.
 */
async function fetchRawFile(info: RepoInfo, filename: string): Promise<string | null> {
  const { owner, repo, branch, path } = info;
  const filePath = path ? `${path}/${filename}` : filename;

  // If branch is known, fetch directly
  if (branch) {
      try {
        const res = await fetch(`${RAW_GITHUB_BASE}/${owner}/${repo}/${branch}/${filePath}`);
        if (res.ok) return await res.text();
        return null;
      } catch {
          return null;
      }
  }

  // Fallback: Try 'main' then 'master'
  try {
      const resMain = await fetch(`${RAW_GITHUB_BASE}/${owner}/${repo}/main/${filePath}`);
      if (resMain.ok) return await resMain.text();
  } catch {}

  try {
      const resMaster = await fetch(`${RAW_GITHUB_BASE}/${owner}/${repo}/master/${filePath}`);
      if (resMaster.ok) return await resMaster.text();
  } catch {}

  return null;
}

/**
 * Scans text for environment variable patterns.
 */
function scanForEnvVars(text: string): Record<string, SecretValue> {
  const envVars: Record<string, SecretValue> = {};

  // Pattern 1: export VAR_NAME=value
  const exportRegex = /export\s+([A-Z_][A-Z0-9_]*)=/g;
  let match;
  while ((match = exportRegex.exec(text)) !== null) {
    envVars[match[1]] = { plainText: "", validationRegex: "" };
  }

  // Pattern 2: process.env.VAR_NAME (Node)
  const nodeEnvRegex = /process\.env\.([A-Z_][A-Z0-9_]*)/g;
  while ((match = nodeEnvRegex.exec(text)) !== null) {
    envVars[match[1]] = { plainText: "", validationRegex: "" };
  }

  // Pattern 3: os.environ.get("VAR_NAME") or os.getenv("VAR_NAME") (Python)
  const pythonEnvRegex = /os\.(?:environ\.get|getenv)\(["']([A-Z_][A-Z0-9_]*)["']\)/g;
  while ((match = pythonEnvRegex.exec(text)) !== null) {
    envVars[match[1]] = { plainText: "", validationRegex: "" };
  }

  // Pattern 4: VAR_NAME=... in code blocks (often in README)
  // We allow optional whitespace at start of line to handle indented code blocks
  const strictEnvLineRegex = /^\s*([A-Z_][A-Z0-9_]*)=/gm;
  while ((match = strictEnvLineRegex.exec(text)) !== null) {
    envVars[match[1]] = { plainText: "", validationRegex: "" };
  }

  // Filter out common false positives
  const ignored = new Set(["NODE_ENV", "PATH", "PYTHONPATH", "HOME", "USER", "SHELL", "PIP_CACHE_DIR", "CI"]);
  Object.keys(envVars).forEach(k => {
    if (ignored.has(k)) delete envVars[k];
  });

  return envVars;
}

/**
 * Analyzes a GitHub repository to determine MCP configuration.
 */
export async function analyzeRepository(url: string): Promise<AnalysisResult> {
  const repoInfo = parseGithubUrl(url);
  if (!repoInfo) {
    return {
      name: "",
      command: "",
      envVars: {},
      detectedType: "unknown",
      description: "",
      error: "Invalid GitHub URL"
    };
  }

  let name = repoInfo.repo;
  // If analyzing subpath, append dir name to name?
  if (repoInfo.path) {
      const parts = repoInfo.path.split('/');
      name = parts[parts.length - 1];
  }

  let command = "";
  let detectedType: "node" | "python" | "unknown" = "unknown";
  let description = "";
  let envVars: Record<string, SecretValue> = {};

  // 1. Fetch metadata files in parallel
  const [pkgJsonStr, pyProjectStr, readmeStr] = await Promise.all([
    fetchRawFile(repoInfo, "package.json"),
    fetchRawFile(repoInfo, "pyproject.toml"),
    fetchRawFile(repoInfo, "README.md")
  ]);

  // 2. Scan README for env vars (and description if needed)
  if (readmeStr) {
    envVars = { ...envVars, ...scanForEnvVars(readmeStr) };
  }

  // 3. Detect Node.js
  if (pkgJsonStr) {
    detectedType = "node";
    try {
      const pkg = JSON.parse(pkgJsonStr);
      if (pkg.name) name = pkg.name;
      if (pkg.description) description = pkg.description;

      // Heuristic: npx -y <package-name>
      command = `npx -y ${pkg.name}`;

      envVars = { ...envVars, ...scanForEnvVars(JSON.stringify(pkg.scripts || {})) };
    } catch (e) {
      console.warn("Failed to parse package.json", e);
    }
  }

  // 4. Detect Python (if not Node, or if Node failed to yield good results)
  if (pyProjectStr && detectedType === "unknown") {
    detectedType = "python";
    // Parse pyproject.toml
    const nameMatch = pyProjectStr.match(/name\s*=\s*["']([^"']+)["']/);
    if (nameMatch) {
        const pkgName = nameMatch[1];
        name = pkgName;
        command = `uvx ${pkgName}`;
    } else {
        command = `uvx ${name}`;
    }

    const descMatch = pyProjectStr.match(/description\s*=\s*["']([^"']+)["']/);
    if (descMatch) description = descMatch[1];
  }

  return {
    name,
    command,
    envVars,
    detectedType,
    description
  };
}
