#!/usr/bin/env node

const path = require('path');
const { spawnSync } = require('child_process');

const pluginRoot = path.resolve(process.env.CLAUDE_PLUGIN_ROOT || path.resolve(__dirname, '..', '..'));
const harness = path.join(pluginRoot, 'bin', 'harness');
const result = spawnSync(harness, ['hook', 'pre-compact-save'], {
  cwd: process.cwd(),
  env: process.env,
  stdio: 'inherit',
});

if (result.error) {
  console.error(`[pre-compact-save] failed to run harness: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 1);
