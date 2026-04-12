#!/usr/bin/env bash
# generate-skill-manifest.sh
# Convert skill frontmatter in the repo to machine-readable JSON.

set -euo pipefail

usage() {
  cat >&2 <<'EOF'
Usage: scripts/generate-skill-manifest.sh [--output PATH]
EOF
  exit 1
}

OUTPUT_FILE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --output)
      OUTPUT_FILE="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

node - "$OUTPUT_FILE" <<'NODE'
const fs = require('fs');
const path = require('path');

process.stdout.on('error', (error) => {
  if (error.code === 'EPIPE') process.exit(0);
  throw error;
});

const [outputFile] = process.argv.slice(2);
const repoRoot = process.cwd();
const roots = ['skills', 'skills-v3', 'skills-v3-codex', 'codex/.codex/skills', 'opencode/skills'];

function walk(dirPath, entries) {
  if (!fs.existsSync(dirPath)) return;
  for (const entry of fs.readdirSync(dirPath, { withFileTypes: true })) {
    if (entry.name === '.git' || entry.name === 'node_modules') continue;
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      walk(fullPath, entries);
    } else if (entry.isFile() && entry.name === 'SKILL.md') {
      entries.push(fullPath);
    }
  }
}

function parseScalar(value) {
  const trimmed = value.trim();
  if (trimmed === 'true') return true;
  if (trimmed === 'false') return false;
  if (/^-?\d+(?:\.\d+)?$/.test(trimmed)) return Number(trimmed);
  if ((trimmed.startsWith('[') && trimmed.endsWith(']')) || (trimmed.startsWith('{') && trimmed.endsWith('}'))) {
    try {
      return JSON.parse(trimmed);
    } catch {
      return trimmed;
    }
  }
  if ((trimmed.startsWith('"') && trimmed.endsWith('"')) || (trimmed.startsWith("'") && trimmed.endsWith("'"))) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function parseFrontmatter(filePath) {
  const raw = fs.readFileSync(filePath, 'utf8');
  const lines = raw.split(/\r?\n/);
  if (lines[0] !== '---') return null;

  const frontmatter = {};
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i];
    if (line === '---') break;
    const match = line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/);
    if (!match) continue;
    const [, key, value] = match;
    frontmatter[key] = parseScalar(value);
  }
  return frontmatter;
}

function parseDoNotUseFor(frontmatter) {
  const candidates = [
    frontmatter['description-en'],
    frontmatter['description-ja'],
    frontmatter.description,
  ].filter(Boolean);

  for (const candidate of candidates) {
    const englishMatch = String(candidate).match(/Do NOT (?:use|load) for:\s*(.+)$/i);
    if (englishMatch) {
      return englishMatch[1]
        .split(',')
        .map((item) => item.trim().replace(/^(and|or)\s+/i, '').replace(/[.。]+$/g, ''))
        .filter(Boolean);
    }

    // Legacy Japanese pattern support removed (codebase is now English-only)
  }

  return [];
}

function rootNameFor(relativePath) {
  for (const root of roots) {
    const normalized = `${root}/`;
    if (relativePath.startsWith(normalized)) return root;
  }
  return 'unknown';
}

const skillFiles = [];
for (const root of roots) {
  walk(path.join(repoRoot, root), skillFiles);
}

const skills = skillFiles
  .map((filePath) => {
    const frontmatter = parseFrontmatter(filePath);
    if (!frontmatter) return null;

    const relativePath = path.relative(repoRoot, filePath).split(path.sep).join('/');
    const directory = path.dirname(relativePath).split(path.sep).join('/');
    const surface = rootNameFor(relativePath);
    return {
      path: relativePath,
      directory,
      surface,
      name: frontmatter.name || path.basename(path.dirname(filePath)),
      description: frontmatter.description || null,
      description_en: frontmatter['description-en'] || null,
      description_ja: frontmatter['description-ja'] || null,
      allowed_tools: Array.isArray(frontmatter['allowed-tools']) ? frontmatter['allowed-tools'] : [],
      argument_hint: frontmatter['argument-hint'] || null,
      context: frontmatter.context || null,
      effort: frontmatter.effort || null,
      user_invocable: typeof frontmatter['user-invocable'] === 'boolean' ? frontmatter['user-invocable'] : null,
      do_not_use_for: parseDoNotUseFor(frontmatter),
    };
  })
  .filter(Boolean)
  .sort((a, b) => a.path.localeCompare(b.path));

const surfaceIndex = new Map();
for (const skill of skills) {
  const surfaces = surfaceIndex.get(skill.name) || new Set();
  surfaces.add(skill.surface);
  surfaceIndex.set(skill.name, surfaces);
}

for (const skill of skills) {
  skill.related_surfaces = Array.from(surfaceIndex.get(skill.name) || []).sort();
}

const manifest = {
  schema_version: 'skill-manifest.v1',
  generated_at: new Date().toISOString(),
  repo_root: repoRoot,
  roots,
  skill_count: skills.length,
  skills,
};

const output = `${JSON.stringify(manifest, null, 2)}\n`;
if (outputFile) {
  fs.mkdirSync(path.dirname(outputFile), { recursive: true });
  fs.writeFileSync(outputFile, output);
}
process.stdout.write(output);
NODE
