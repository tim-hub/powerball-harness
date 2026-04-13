#!/usr/bin/env node
/**
 * sync-skill-mirrors.mjs
 * Sync all skills from skills/ (SSOT) to both mirror targets.
 *
 * Mirrors:
 *   - codex/.codex/skills/
 *   - opencode/skills/
 *
 * Usage:
 *   node .claude/scripts/sync-skill-mirrors.mjs          # sync
 *   node .claude/scripts/sync-skill-mirrors.mjs --check  # verify (exit 1 if drift)
 */

import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const PLUGIN_ROOT = path.join(__dirname, '..', '..');
const SSOT_DIR = path.join(PLUGIN_ROOT, 'skills');

const MIRRORS = [
  'codex/.codex/skills',
  'opencode/skills',
];

const mode = process.argv[2] === '--check' ? 'check' : 'sync';
if (process.argv[2] && process.argv[2] !== '--check') {
  console.error('Usage: node sync-skill-mirrors.mjs [--check]');
  process.exit(2);
}

function copyDir(src, dst) {
  fs.mkdirSync(dst, { recursive: true });
  for (const entry of fs.readdirSync(src, { withFileTypes: true })) {
    const s = path.join(src, entry.name);
    const d = path.join(dst, entry.name);
    entry.isDirectory() ? copyDir(s, d) : fs.copyFileSync(s, d);
  }
}

function diffQuiet(a, b) {
  try {
    execFileSync('diff', ['-qr', a, b], { stdio: 'pipe' });
    return true;
  } catch { return false; }
}

// Collect all skill directories from SSOT
const skills = fs.readdirSync(SSOT_DIR, { withFileTypes: true })
  .filter(e => e.isDirectory() && fs.existsSync(path.join(SSOT_DIR, e.name, 'SKILL.md')))
  .map(e => e.name);

let failures = 0;

for (const mirror of MIRRORS) {
  const mirrorDir = path.join(PLUGIN_ROOT, mirror);
  if (!fs.existsSync(mirrorDir)) continue;

  for (const skill of skills) {
    const src = path.join(SSOT_DIR, skill);
    const dst = path.join(mirrorDir, skill);

    if (mode === 'sync') {
      fs.rmSync(dst, { recursive: true, force: true });
      copyDir(src, dst);
      console.log(`synced ${mirror}/${skill}`);
    } else {
      if (!fs.existsSync(dst) || fs.lstatSync(dst).isSymbolicLink() || !diffQuiet(src, dst)) {
        console.error(`drift ${mirror}/${skill}`);
        failures++;
      } else {
        console.log(`ok ${mirror}/${skill}`);
      }
    }
  }
}

if (mode === 'check' && failures > 0) process.exit(1);
