#!/usr/bin/env node
/**
 * record-usage.js - Harness resource usage recording script
 *
 * Usage:
 *   node record-usage.js <type> <name> [options]
 *
 * Types: skill, command, agent, hook
 * Options: --blocked (for hooks that blocked an action)
 *
 * Example:
 *   node record-usage.js skill impl
 *   node record-usage.js hook test-quality-guard --blocked
 */

const fs = require('fs');
const path = require('path');

// Usage file location - project-local
const USAGE_FILE = path.join(process.cwd(), '.claude', 'state', 'harness-usage.json');

// Schema version for future migrations
const SCHEMA_VERSION = '1.0';

/**
 * Initialize empty usage data
 */
function createEmptyUsage() {
  return {
    version: SCHEMA_VERSION,
    updatedAt: new Date().toISOString(),
    skills: {},
    commands: {},
    agents: {},
    hooks: {}
  };
}

/**
 * Normalize an agent key to a harness role.
 */
function normalizeAgentRole(name) {
  const value = String(name || '').trim().toLowerCase();
  if (!value) {
    return 'unknown';
  }
  if (value.includes('review')) {
    return 'reviewer';
  }
  if (value.includes('lead') || value.includes('planner')) {
    return 'lead';
  }
  if (value.includes('worker') || value.includes('impl')) {
    return 'worker';
  }
  return value;
}

/**
 * Load usage data from file
 */
function loadUsage() {
  try {
    if (fs.existsSync(USAGE_FILE)) {
      const data = JSON.parse(fs.readFileSync(USAGE_FILE, 'utf-8'));
      // Ensure all required sections exist
      return {
        version: data.version || SCHEMA_VERSION,
        updatedAt: data.updatedAt || new Date().toISOString(),
        skills: data.skills || {},
        commands: data.commands || {},
        agents: data.agents || {},
        hooks: data.hooks || {}
      };
    }
  } catch (err) {
    // If file is corrupted, start fresh
    console.error(`[record-usage] Warning: Could not load usage file: ${err.message}`);
  }
  return createEmptyUsage();
}

/**
 * Save usage data to file
 */
function saveUsage(usage) {
  try {
    // Ensure directory exists
    const dir = path.dirname(USAGE_FILE);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    usage.updatedAt = new Date().toISOString();
    fs.writeFileSync(USAGE_FILE, JSON.stringify(usage, null, 2), 'utf-8');
  } catch (err) {
    console.error(`[record-usage] Error saving usage file: ${err.message}`);
    process.exit(1);
  }
}

/**
 * Record usage for a resource
 */
function recordUsage(type, name, options = {}) {
  if (!type || !name) {
    console.error('[record-usage] Error: type and name are required');
    process.exit(1);
  }

  const validTypes = ['skill', 'command', 'agent', 'hook'];
  if (!validTypes.includes(type)) {
    console.error(`[record-usage] Error: Invalid type "${type}". Must be one of: ${validTypes.join(', ')}`);
    process.exit(1);
  }

  const usage = loadUsage();
  const section = type + 's'; // skill -> skills, etc.
  const now = new Date().toISOString();

  if (type === 'hook') {
    // Hooks have triggered/blocked counts
    if (!usage[section][name]) {
      usage[section][name] = { triggered: 0, blocked: 0, lastTriggered: null };
    }
    usage[section][name].triggered += 1;
    usage[section][name].lastTriggered = now;
    if (options.blocked) {
      usage[section][name].blocked += 1;
    }
  } else {
    // Skills, commands, agents have count/lastUsed
    if (!usage[section][name]) {
      usage[section][name] = { count: 0, lastUsed: null };
    }
    usage[section][name].count += 1;
    usage[section][name].lastUsed = now;
  }

  saveUsage(usage);

  // Output success for logging
  console.log(`[record-usage] Recorded ${type}: ${name}`);
}

/**
 * Get cleanup suggestions based on usage data
 */
function getCleanupSuggestions() {
  const usage = loadUsage();
  const suggestions = {
    unusedSkills: [],
    unusedCommands: [],
    unusedAgents: [],
    inactiveHooks: [],
    summary: {}
  };

  const thirtyDaysAgo = new Date();
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

  // Check each section for unused/inactive resources
  for (const [name, data] of Object.entries(usage.skills)) {
    if (data.count === 0 || (data.lastUsed && new Date(data.lastUsed) < thirtyDaysAgo)) {
      suggestions.unusedSkills.push({ name, ...data });
    }
  }

  for (const [name, data] of Object.entries(usage.commands)) {
    if (data.count === 0 || (data.lastUsed && new Date(data.lastUsed) < thirtyDaysAgo)) {
      suggestions.unusedCommands.push({ name, ...data });
    }
  }

  for (const [name, data] of Object.entries(usage.agents)) {
    if (data.count === 0 || (data.lastUsed && new Date(data.lastUsed) < thirtyDaysAgo)) {
      suggestions.unusedAgents.push({ name, ...data });
    }
  }

  for (const [name, data] of Object.entries(usage.hooks)) {
    if (data.triggered === 0 || (data.lastTriggered && new Date(data.lastTriggered) < thirtyDaysAgo)) {
      suggestions.inactiveHooks.push({ name, ...data });
    }
  }

  // Summary stats
  suggestions.summary = {
    totalSkills: Object.keys(usage.skills).length,
    totalCommands: Object.keys(usage.commands).length,
    totalAgents: Object.keys(usage.agents).length,
    totalHooks: Object.keys(usage.hooks).length,
    unusedSkillsCount: suggestions.unusedSkills.length,
    unusedCommandsCount: suggestions.unusedCommands.length,
    unusedAgentsCount: suggestions.unusedAgents.length,
    inactiveHooksCount: suggestions.inactiveHooks.length
  };

  return suggestions;
}

/**
 * Get full usage report
 */
function getUsageReport() {
  const usage = loadUsage();

  // Sort by usage count (descending)
  const sortByCount = (a, b) => (b[1].count || b[1].triggered || 0) - (a[1].count || a[1].triggered || 0);
  const roleSummary = {};

  for (const [name, data] of Object.entries(usage.agents)) {
    const role = normalizeAgentRole(name);
    if (!roleSummary[role]) {
      roleSummary[role] = {
        count: 0,
        retryCount: 0,
        lastUsed: null,
        agentNames: []
      };
    }

    const count = data.count || 0;
    roleSummary[role].count += count;
    roleSummary[role].retryCount += data.retryCount || 0;
    if (!roleSummary[role].lastUsed || (data.lastUsed && new Date(data.lastUsed) > new Date(roleSummary[role].lastUsed))) {
      roleSummary[role].lastUsed = data.lastUsed || roleSummary[role].lastUsed;
    }
    roleSummary[role].agentNames.push(name);
  }

  return {
    version: usage.version,
    updatedAt: usage.updatedAt,
    skills: Object.entries(usage.skills).sort(sortByCount),
    commands: Object.entries(usage.commands).sort(sortByCount),
    agents: Object.entries(usage.agents).sort(sortByCount),
    roles: Object.entries(roleSummary)
      .map(([role, summary]) => [role, summary])
      .sort((a, b) => (b[1].count || 0) - (a[1].count || 0)),
    hooks: Object.entries(usage.hooks).sort(sortByCount),
    cleanup: getCleanupSuggestions()
  };
}

// Main execution
const args = process.argv.slice(2);

if (args.length === 0) {
  console.error('Usage: node record-usage.js <type> <name> [--blocked]');
  console.error('       node record-usage.js --report');
  console.error('       node record-usage.js --cleanup');
  process.exit(1);
}

if (args[0] === '--report') {
  console.log(JSON.stringify(getUsageReport(), null, 2));
} else if (args[0] === '--cleanup') {
  console.log(JSON.stringify(getCleanupSuggestions(), null, 2));
} else {
  const type = args[0];
  const name = args[1];
  const options = {
    blocked: args.includes('--blocked')
  };
  recordUsage(type, name, options);
}
