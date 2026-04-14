#!/usr/bin/env node
/**
 * check-japanese.mjs
 * Scans git diff (staged + unstaged changes) for Japanese characters.
 * Run as a TaskCompleted hook — alerts if Japanese is found in changed lines.
 */

import { execSync } from 'node:child_process';

const JAPANESE_RE = /[\u3000-\u9fff\uff00-\uffef\u3040-\u30ff]/;

// Exclude lines that are intentionally Japanese (grep patterns, Unicode escapes in source)
const ALLOWLIST_RE = /\\u[0-9a-f]{4}|grep.*\[\\x|awk.*\[\\x/i;

let diff;
try {
  // Get both staged and unstaged changes vs HEAD
  diff = execSync('git diff HEAD 2>/dev/null || git diff 2>/dev/null', {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  });
} catch {
  process.exit(0); // No git repo or no diff — silent pass
}

const hits = [];

for (const line of diff.split('\n')) {
  // Only check added lines (prefix +) but not the +++ file header
  if (!line.startsWith('+') || line.startsWith('+++')) continue;
  const content = line.slice(1);
  if (JAPANESE_RE.test(content) && !ALLOWLIST_RE.test(content)) {
    hits.push(line);
  }
}

if (hits.length === 0) process.exit(0);

const preview = hits.slice(0, 5).map(l => `  ${l}`).join('\n');
const more = hits.length > 5 ? `\n  ... and ${hits.length - 5} more line(s)` : '';

// Output a warning that Claude Code will show to the user
process.stdout.write(JSON.stringify({
  message: `[Japanese detected] ${hits.length} line(s) with Japanese characters found in changes:\n${preview}${more}\n\nPlease translate or remove Japanese before completing.`,
}));
process.exit(2); // exit 2 = block/alert
