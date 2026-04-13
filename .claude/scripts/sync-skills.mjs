#!/usr/bin/env node
/**
 * sync-skills.mjs
 * PostToolUse hook: sync all skill mirrors when a file under skills/ is edited.
 *
 * Reads the hook JSON from stdin, extracts tool_input.file_path,
 * and syncs both mirrors when the edited file is under skills/
 * (excluding opencode/skills/, skills-codex/, and codex/.codex/skills/).
 *
 * Mirrors synced:
 *   - opencode/skills/  (via build-opencode.mjs)
 *   - codex/.codex/skills/  (via sync-skill-mirrors.mjs)
 *
 * Usage: Called by .claude/settings.json PostToolUse (Write|Edit matcher)
 */

import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.join(__dirname, '..', '..');

// Read hook input from stdin
let input = '';
try {
  input = fs.readFileSync(0, 'utf8');
} catch {
  process.exit(0);
}

// Extract file_path from tool_input
let filePath = '';
try {
  const parsed = JSON.parse(input);
  filePath = parsed?.tool_input?.file_path ?? '';
} catch {
  process.exit(0);
}

if (!filePath) process.exit(0);

// Only trigger for files under skills/ (the SSOT directory)
if (!filePath.includes('/skills/')) process.exit(0);

// Exclude mirror directories to avoid infinite loops
if (
  filePath.includes('/opencode/skills/') ||
  filePath.includes('/skills-codex/') ||
  filePath.includes('/.codex/skills/')
) {
  process.exit(0);
}

// Sync both mirrors
try {
  execFileSync('node', [path.join(__dirname, 'build-opencode.mjs')], {
    stdio: ['pipe', 'pipe', 'pipe'],
  });
} catch { /* non-fatal */ }

try {
  execFileSync('node', [path.join(__dirname, 'sync-skill-mirrors.mjs')], {
    stdio: ['pipe', 'pipe', 'pipe'],
  });
} catch { /* non-fatal */ }
