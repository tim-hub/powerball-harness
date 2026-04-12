#!/usr/bin/env node
/**
 * run-script.js
 * Cross-platform bash script runner for Windows/Mac/Linux
 *
 * Purpose:
 * - Resolve ${CLAUDE_PLUGIN_ROOT} path issues on Windows
 * - Convert C:\Users\... to /c/Users/... format for bash
 *
 * Usage:
 *   node run-script.js <script-name> [args...]
 *   Example: node run-script.js session-init
 *       node run-script.js posttooluse-log-toolname
 *
 * Usage in hooks.json:
 *   "command": "node ${CLAUDE_PLUGIN_ROOT}/scripts/run-script.js session-init"
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Platform detection
const isWindows = process.platform === 'win32';

/**
 * Convert Windows path to MSYS/Git Bash format
 * C:\Users\foo → /c/Users/foo
 * \\server\share → //server/share
 */
function toMsysPath(windowsPath) {
  if (!windowsPath) return windowsPath;

  // Convert backslashes to forward slashes
  let msysPath = windowsPath.replace(/\\/g, '/');

  // Drive letter conversion: C:/ -> /c/
  const driveMatch = msysPath.match(/^([A-Za-z]):\//);
  if (driveMatch) {
    msysPath = '/' + driveMatch[1].toLowerCase() + msysPath.slice(2);
  }

  return msysPath;
}

/**
 * Detect bash executable path
 */
function findBash() {
  if (!isWindows) {
    return 'bash';
  }

  // Windows: Look for Git Bash's bash
  const possiblePaths = [
    'C:\\Program Files\\Git\\bin\\bash.exe',
    'C:\\Program Files (x86)\\Git\\bin\\bash.exe',
    process.env.PROGRAMFILES + '\\Git\\bin\\bash.exe',
    process.env['PROGRAMFILES(X86)'] + '\\Git\\bin\\bash.exe',
    'C:\\msys64\\usr\\bin\\bash.exe',
    'C:\\msys32\\usr\\bin\\bash.exe',
  ];

  for (const bashPath of possiblePaths) {
    if (bashPath && fs.existsSync(bashPath)) {
      return bashPath;
    }
  }

  // Fallback: use bash from PATH
  return 'bash';
}

/**
 * Main processing
 */
function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    console.error('Usage: node run-script.js <script-name> [args...]');
    console.error('Example: node run-script.js session-init');
    process.exit(1);
  }

  const scriptName = args[0];
  const scriptArgs = args.slice(1);

  // Get script directory
  const scriptsDir = __dirname;

  // Build script path
  let scriptPath = path.join(scriptsDir, scriptName);

  // Add .sh extension if missing
  if (!scriptPath.endsWith('.sh')) {
    scriptPath += '.sh';
  }

  // Verify script exists
  if (!fs.existsSync(scriptPath)) {
    console.error(`Error: Script not found: ${scriptPath}`);
    process.exit(1);
  }

  // Detect bash executable
  const bashPath = findBash();

  // Convert path to MSYS format on Windows
  let bashScriptPath = scriptPath;
  if (isWindows) {
    bashScriptPath = toMsysPath(scriptPath);
  }

  // Prepare environment variables
  const env = { ...process.env };

  if (isWindows) {
    // Disable MSYS path conversion (prevent double conversion)
    env.MSYS_NO_PATHCONV = '1';
    env.MSYS2_ARG_CONV_EXCL = '*';

    // Also convert CLAUDE_PLUGIN_ROOT
    if (env.CLAUDE_PLUGIN_ROOT) {
      env.CLAUDE_PLUGIN_ROOT = toMsysPath(env.CLAUDE_PLUGIN_ROOT);
    }
  }

  // Execute bash script
  const child = spawn(bashPath, [bashScriptPath, ...scriptArgs], {
    env,
    stdio: 'inherit',  // Transparently forward stdin/stdout/stderr
    shell: false,
  });

  child.on('error', (err) => {
    console.error(`Failed to execute bash: ${err.message}`);
    if (isWindows) {
      console.error('Hint: Make sure Git Bash is installed');
    }
    process.exit(1);
  });

  child.on('exit', (code, signal) => {
    if (signal) {
      process.exit(1);
    }
    process.exit(code || 0);
  });
}

main();
