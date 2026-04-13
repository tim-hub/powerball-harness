#!/usr/bin/env node
/**
 * sync-skill-mirrors.mjs
 * Sync skills from skills/ (SSOT) to Codex/OpenCode mirrors.
 *
 * Why:
 *   Windows checkout with core.symlinks=false turns repository symlinks into
 *   plain text files. Claude Code ignores those files when building the slash
 *   command list, so the harness-* entry skills disappear before SessionStart
 *   repair hooks can run.
 *
 * This script keeps skills as real directories in:
 *   - codex/.codex/skills/
 *   - opencode/skills/
 *
 * Source of truth: skills/ (the main skills directory)
 *
 * Sync scope:
 *   - All skill directories that exist in BOTH skills/ (SSOT) and a mirror root
 *   - New skills added only to skills/ are NOT auto-propagated (add manually)
 *   - routing-rules.md is synced if present in both
 *
 * Usage:
 *   node .claude/scripts/sync-skill-mirrors.mjs          # overwrite mirrors from skills/
 *   node .claude/scripts/sync-skill-mirrors.mjs --check   # verify mirrors match skills/
 */

import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PLUGIN_ROOT = path.join(__dirname, '..', '..');
const SSOT_DIR = path.join(PLUGIN_ROOT, 'skills');

const MIRROR_ROOTS = [
  'codex/.codex/skills',
  'opencode/skills',
];

const mode = process.argv[2] === '--check' ? 'check' : 'sync';

if (process.argv[2] && process.argv[2] !== '--check') {
  console.error('Usage: node sync-skill-mirrors.mjs [--check]');
  process.exit(2);
}

function diffQuiet(src, dst) {
  try {
    execFileSync('diff', ['-qr', src, dst], { stdio: ['pipe', 'pipe', 'pipe'] });
    return true;
  } catch {
    return false;
  }
}

function diffQuietFile(src, dst) {
  try {
    execFileSync('diff', ['-q', src, dst], { stdio: ['pipe', 'pipe', 'pipe'] });
    return true;
  } catch {
    return false;
  }
}

function copyDirRecursive(src, dst) {
  fs.mkdirSync(dst, { recursive: true });
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const srcPath = path.join(src, entry.name);
    const dstPath = path.join(dst, entry.name);
    if (entry.isDirectory()) {
      copyDirRecursive(srcPath, dstPath);
    } else {
      fs.copyFileSync(srcPath, dstPath);
    }
  }
}

function syncSkill(skill, mirrorRoot) {
  const src = path.join(SSOT_DIR, skill);
  const dstRoot = path.join(PLUGIN_ROOT, mirrorRoot);
  const dst = path.join(dstRoot, skill);

  fs.mkdirSync(dstRoot, { recursive: true });
  fs.rmSync(dst, { recursive: true, force: true });
  copyDirRecursive(src, dst);
  console.log(`synced ${mirrorRoot}/${skill}`);
}

function checkSkill(skill, mirrorRoot) {
  const src = path.join(SSOT_DIR, skill);
  const dst = path.join(PLUGIN_ROOT, mirrorRoot, skill);

  if (!fs.existsSync(dst)) {
    console.error(`missing ${mirrorRoot}/${skill}`);
    return false;
  }

  const stat = fs.lstatSync(dst);
  if (stat.isSymbolicLink()) {
    console.error(`symlink ${mirrorRoot}/${skill}`);
    return false;
  }

  if (!diffQuiet(src, dst)) {
    console.error(`drift ${mirrorRoot}/${skill}`);
    return false;
  }

  console.log(`ok ${mirrorRoot}/${skill}`);
  return true;
}

// Skills excluded per mirror target.
// Iterate SSOT and skip these — mirrors stay in sync with new skills automatically.
const SKIP_PER_MIRROR = {
  'codex/.codex/skills': new Set(['allow1', 'cc-update-review', 'claude-codex-upstream-update']),
  'opencode/skills':     new Set(['allow1', 'breezing', 'cc-update-review', 'claude-codex-upstream-update']),
};

let failures = 0;

for (const mirrorRoot of MIRROR_ROOTS) {
  const mirrorDir = path.join(PLUGIN_ROOT, mirrorRoot);
  if (!fs.existsSync(mirrorDir)) continue;

  const skipSkills = SKIP_PER_MIRROR[mirrorRoot] ?? new Set();

  // Iterate SSOT — not the mirror — so new skills are picked up automatically
  for (const entry of fs.readdirSync(SSOT_DIR, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const skill = entry.name;

    if (skill === 'node_modules' || skill === '.git') continue;
    if (skill.startsWith('test-') || skill.startsWith('x-') || skill.startsWith('zz-')) continue;
    if (skipSkills.has(skill)) continue;
    if (!fs.existsSync(path.join(SSOT_DIR, skill, 'SKILL.md'))) continue;

    if (mode === 'sync') {
      syncSkill(skill, mirrorRoot);
    } else {
      if (!checkSkill(skill, mirrorRoot)) {
        failures++;
      }
    }
  }

  const ssotRules = path.join(SSOT_DIR, 'routing-rules.md');
  const mirrorRules = path.join(mirrorDir, 'routing-rules.md');

  if (fs.existsSync(ssotRules) && fs.existsSync(mirrorRules)) {
    if (mode === 'sync') {
      fs.copyFileSync(ssotRules, mirrorRules);
      console.log(`synced ${mirrorRoot}/routing-rules.md`);
    } else {
      if (!diffQuietFile(ssotRules, mirrorRules)) {
        console.error(`drift ${mirrorRoot}/routing-rules.md`);
        failures++;
      } else {
        console.log(`ok ${mirrorRoot}/routing-rules.md`);
      }
    }
  }
}

if (mode === 'check' && failures > 0) {
  process.exit(1);
}
