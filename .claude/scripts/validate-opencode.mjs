#!/usr/bin/env node
/**
 * validate-opencode.mjs
 *
 * Validate that files converted for opencode are in the correct format.
 *
 * What it checks:
 * - No opencode-incompatible fields in frontmatter
 * - Required files exist
 * - JSON files are valid
 *
 * Usage:
 *   node .claude/scripts/validate-opencode.mjs
 *
 * Exit codes:
 *   0: Validation succeeded
 *   1: Validation failed
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT_DIR = path.join(__dirname, '..', '..');
const OPENCODE_DIR = path.join(ROOT_DIR, 'opencode');

// Frontmatter fields that are invalid for opencode
const INVALID_FIELDS = ['description-en', 'name'];

// Required files (v2.17.0+: commands have been migrated to Skills; skills are now required)
const REQUIRED_FILES = [
  'opencode/AGENTS.md',
  'opencode/opencode.json',
  'opencode/README.md',
  'opencode/skills',  // Skills are now the primary mechanism
];

let errors = [];
let warnings = [];

/**
 * Parse frontmatter
 */
function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) {
    return null;
  }

  const frontmatterStr = match[1];
  const frontmatter = {};
  const lines = frontmatterStr.split('\n');

  for (const line of lines) {
    const colonIndex = line.indexOf(':');
    if (colonIndex > 0) {
      const key = line.slice(0, colonIndex).trim();
      const value = line.slice(colonIndex + 1).trim();
      frontmatter[key] = value;
    }
  }

  return frontmatter;
}

/**
 * Validate a command file
 */
function validateCommandFile(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  const frontmatter = parseFrontmatter(content);
  const relativePath = path.relative(ROOT_DIR, filePath);

  if (!frontmatter) {
    // Files without frontmatter produce a warning only
    warnings.push(`${relativePath}: No frontmatter found`);
    return;
  }

  // Check for invalid fields
  for (const field of INVALID_FIELDS) {
    if (frontmatter[field]) {
      errors.push(`${relativePath}: Invalid field '${field}' found in frontmatter`);
    }
  }

  // Warn when description is missing
  if (!frontmatter.description) {
    warnings.push(`${relativePath}: Missing 'description' field`);
  }
}

/**
 * Recursively validate files in a directory
 */
function validateDirectory(dir) {
  if (!fs.existsSync(dir)) {
    errors.push(`Directory not found: ${path.relative(ROOT_DIR, dir)}`);
    return;
  }

  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      validateDirectory(fullPath);
    } else if (entry.name.endsWith('.md')) {
      validateCommandFile(fullPath);
    }
  }
}

/**
 * Validate a JSON file
 */
function validateJsonFile(filePath) {
  const relativePath = path.relative(ROOT_DIR, filePath);

  if (!fs.existsSync(filePath)) {
    errors.push(`File not found: ${relativePath}`);
    return;
  }

  try {
    const content = fs.readFileSync(filePath, 'utf8');
    JSON.parse(content);
  } catch (e) {
    errors.push(`${relativePath}: Invalid JSON - ${e.message}`);
  }
}

/**
 * Verify that required files exist
 */
function validateRequiredFiles() {
  for (const file of REQUIRED_FILES) {
    const fullPath = path.join(ROOT_DIR, file);
    if (!fs.existsSync(fullPath)) {
      errors.push(`Required file/directory not found: ${file}`);
    }
  }
}

/**
 * Validate the structure of opencode.json
 */
function validateOpencodeConfig() {
  const configPath = path.join(OPENCODE_DIR, 'opencode.json');

  if (!fs.existsSync(configPath)) {
    return; // Already reported in required file check above
  }

  try {
    const content = fs.readFileSync(configPath, 'utf8');
    const config = JSON.parse(content);

    // Check for $schema field
    if (!config.$schema) {
      warnings.push('opencode/opencode.json: Missing $schema field');
    }

    // Check for mcp configuration
    if (config.mcp && config.mcp.harness) {
      const harness = config.mcp.harness;
      if (harness.type !== 'local' && harness.type !== 'remote') {
        errors.push('opencode/opencode.json: Invalid mcp.harness.type (must be "local" or "remote")');
      }
    }
  } catch (e) {
    // JSON parse errors are already reported above
  }
}

/**
 * Main entry point
 */
function main() {
  console.log('🔍 Validating opencode files...\n');

  // Check that required files exist
  console.log('📁 Checking required files...');
  validateRequiredFiles();

  // Validate command files
  console.log('📄 Validating command files...');
  const commandsDir = path.join(OPENCODE_DIR, 'commands');
  if (fs.existsSync(commandsDir)) {
    validateDirectory(commandsDir);
  }

  // Validate JSON files
  console.log('📋 Validating JSON files...');
  validateJsonFile(path.join(OPENCODE_DIR, 'opencode.json'));
  validateOpencodeConfig();

  // Output results
  console.log('\n' + '='.repeat(50));

  if (warnings.length > 0) {
    console.log('\n⚠️  Warnings:');
    for (const warning of warnings) {
      console.log(`   ${warning}`);
    }
  }

  if (errors.length > 0) {
    console.log('\n❌ Errors:');
    for (const error of errors) {
      console.log(`   ${error}`);
    }
    console.log(`\n❌ Validation failed with ${errors.length} error(s).`);
    process.exit(1);
  }

  console.log('\n✅ Validation passed!');
  if (warnings.length > 0) {
    console.log(`   (${warnings.length} warning(s))`);
  }
  process.exit(0);
}

main();
