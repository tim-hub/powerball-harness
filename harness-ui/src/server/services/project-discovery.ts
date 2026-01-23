import { readdir, readFile, realpath } from 'fs/promises';
import { join, resolve, isAbsolute } from 'path';
import type { Dirent } from 'fs';
import type { Project } from '@shared/types';

// plansDirectory 設定を読み取る（検証付き）
async function getPlansDirectory(projectPath: string): Promise<string> {
  const configPath = join(projectPath, '.claude-code-harness.config.yaml');
  try {
    const content = await readFile(configPath, 'utf-8');
    // 簡易 YAML パース: plansDirectory: の値を取得
    const match = content.match(/^plansDirectory:\s*["']?([^"'\n]+)["']?/m);
    const plansDir = match?.[1]?.trim() || '.';

    // Security: Reject absolute paths and parent directory traversal
    if (isAbsolute(plansDir) || plansDir.includes('..')) {
      return '.';
    }

    return plansDir;
  } catch {
    return '.';
  }
}

// plansPath を解決（パストラバーサル対策付き）
async function resolvePlansPath(projectPath: string): Promise<string> {
  const plansDir = await getPlansDirectory(projectPath);
  const basePath = plansDir === '.' ? projectPath : join(projectPath, plansDir);

  // Helper to check if child path is within parent (handles /project vs /project-evil edge case)
  const isWithinDirectory = (parent: string, child: string): boolean => {
    const normalizedParent = parent.endsWith('/') ? parent : parent + '/';
    return child === parent || child.startsWith(normalizedParent);
  };

  // Security: Ensure resolved path is within project root (symlink protection)
  try {
    const resolvedBase = await realpath(projectPath);
    // Resolve plansDir with realpath to catch symlinks pointing outside
    const resolvedPlansDir = plansDir === '.'
      ? resolvedBase
      : await realpath(join(projectPath, plansDir));

    // Check if the resolved path stays within project root
    if (!isWithinDirectory(resolvedBase, resolvedPlansDir)) {
      return join(projectPath, 'Plans.md');
    }
  } catch {
    // If realpath fails (e.g., path doesn't exist yet), use logical check only
    const resolvedPlansDir = resolve(projectPath, plansDir);
    const resolvedBase = resolve(projectPath);
    if (!isWithinDirectory(resolvedBase, resolvedPlansDir)) {
      return join(projectPath, 'Plans.md');
    }
  }

  return join(basePath, 'Plans.md');
}

const DEFAULT_SKIP_DIRS = new Set([
  'node_modules',
  'dist',
  'build',
  'out',
  'target',
  'coverage',
  '.git',
  '.idea',
  '.vscode',
]);

const PROJECT_MARKERS = new Set([
  '.git',
  'package.json',
  'pyproject.toml',
  'requirements.txt',
  'setup.py',
  'go.mod',
  'Cargo.toml',
  'pom.xml',
  'build.gradle',
  'build.gradle.kts',
  'deno.json',
  'deno.jsonc',
  'bunfig.toml',
  'tsconfig.json',
  'Plans.md',
]);

interface DiscoverOptions {
  maxDepth?: number;
  maxProjects?: number;
  rootPath: string;
}

function hashPath(input: string): string {
  let hash = 0;
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash << 5) - hash + input.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash).toString(36);
}

async function toProject(path: string): Promise<Project> {
  const name = path.split('/').pop() || path;
  const plansPath = await resolvePlansPath(path);
  return {
    id: `proj_${hashPath(path)}`,
    name,
    path,
    plansPath,
    worktreePaths: [],
    isActive: false,
  };
}

function isCandidate(names: Set<string>): boolean {
  if (names.has('.git')) return true;
  for (const marker of PROJECT_MARKERS) {
    if (names.has(marker)) return true;
  }
  return false;
}

async function scanDirectory(
  dir: string,
  depth: number,
  maxDepth: number,
  maxProjects: number,
  results: Project[]
): Promise<void> {
  if (depth > maxDepth || results.length >= maxProjects) return;

  let entries: Dirent[];
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return;
  }

  const names = new Set<string>(entries.map((entry) => entry.name));
  if (isCandidate(names)) {
    results.push(await toProject(dir));
    return;
  }

  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    if (entry.name.startsWith('.')) continue;
    if (DEFAULT_SKIP_DIRS.has(entry.name)) continue;
    await scanDirectory(join(dir, entry.name), depth + 1, maxDepth, maxProjects, results);
    if (results.length >= maxProjects) return;
  }
}

export async function discoverProjects(options: DiscoverOptions): Promise<Project[]> {
  const {
    rootPath,
    maxDepth = 2,
    maxProjects = 200,
  } = options;

  const results: Project[] = [];
  await scanDirectory(rootPath, 0, maxDepth, maxProjects, results);
  return results;
}
