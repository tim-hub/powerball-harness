#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

function resolveGoArch() {
  switch (process.arch) {
    case 'x64':
      return 'amd64';
    case 'arm64':
      return 'arm64';
    default:
      return process.arch;
  }
}

function resolveBinary(projectRoot) {
  const binary = path.join(projectRoot, 'bin', `harness-${process.platform}-${resolveGoArch()}`);
  fs.accessSync(binary, fs.constants.X_OK);
  return binary;
}

function runHarnessSubcommand(args) {
  const projectRoot = path.resolve(process.env.CLAUDE_PLUGIN_ROOT || path.resolve(__dirname, '..', '..'));

  let binary;
  try {
    binary = resolveBinary(projectRoot);
  } catch (error) {
    const detail = error && error.message ? error.message : String(error);
    console.error(`[claude-code-harness] platform binary not found or not executable: ${detail}`);
    process.exit(2);
  }

  const result = spawnSync(binary, args, {
    cwd: process.cwd(),
    env: process.env,
    stdio: 'inherit',
  });

  if (result.error) {
    console.error(`[claude-code-harness] failed to run ${path.basename(binary)}: ${result.error.message}`);
    process.exit(1);
  }

  process.exit(result.status ?? 1);
}

module.exports = { runHarnessSubcommand };
