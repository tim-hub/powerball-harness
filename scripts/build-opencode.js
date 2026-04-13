#!/usr/bin/env node
/**
 * build-opencode.js
 *
 * Script to convert Harness commands to opencode.ai compatible format.
 *
 * What it does:
 * - Copies commands/ → opencode/commands/
 * - Removes description-en from frontmatter
 * - Generates CLAUDE.md as AGENTS.md
 *
 * Usage:
 *   node scripts/build-opencode.js
 */

const fs = require('fs');
const path = require('path');

const ROOT_DIR = path.join(__dirname, '..');
const COMMANDS_DIR = path.join(ROOT_DIR, 'commands');
const SKILLS_DIR = path.join(ROOT_DIR, 'skills');
const OPENCODE_DIR = path.join(ROOT_DIR, 'opencode');
const OPENCODE_COMMANDS_DIR = path.join(OPENCODE_DIR, 'commands');
const OPENCODE_SKILLS_DIR = path.join(OPENCODE_DIR, 'skills');
const OPENCODE_TEMPLATES_DIR = path.join(ROOT_DIR, 'templates', 'opencode', 'commands');
const OPENCODE_PM_DIR = path.join(OPENCODE_COMMANDS_DIR, 'pm');

/**
 * Create a directory recursively
 */
function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

/**
 * Clear a directory recursively
 */
function clearDir(dir) {
  if (fs.existsSync(dir)) {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

/**
 * Parse frontmatter
 */
function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) {
    return { frontmatter: null, body: content };
  }

  const frontmatterStr = match[1];
  const body = content.slice(match[0].length);

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

  return { frontmatter, body };
}

/**
 * Serialize frontmatter to a string
 */
function stringifyFrontmatter(frontmatter) {
  const lines = Object.entries(frontmatter)
    .map(([key, value]) => `${key}: ${value}`);
  return `---\n${lines.join('\n')}\n---\n`;
}

/**
 * Convert a Harness command to opencode format
 */
function convertCommand(content) {
  const { frontmatter, body } = parseFrontmatter(content);

  if (!frontmatter) {
    // No frontmatter — return as-is
    return content;
  }

  // Remove fields not needed by opencode
  const opencodeFields = ['description-en', 'name'];
  for (const field of opencodeFields) {
    delete frontmatter[field];
  }

  // If frontmatter is now empty, return only the body
  if (Object.keys(frontmatter).length === 0) {
    return body;
  }

  return stringifyFrontmatter(frontmatter) + body;
}

/**
 * Recursively process files in a directory
 */
function processDirectory(srcDir, destDir) {
  ensureDir(destDir);

  const entries = fs.readdirSync(srcDir, { withFileTypes: true });
  let processedCount = 0;

  for (const entry of entries) {
    const srcPath = path.join(srcDir, entry.name);
    const destPath = path.join(destDir, entry.name);

    if (entry.isDirectory()) {
      processedCount += processDirectory(srcPath, destPath);
    } else if (entry.name.endsWith('.md')) {
      const content = fs.readFileSync(srcPath, 'utf8');
      const converted = convertCommand(content);
      fs.writeFileSync(destPath, converted);
      processedCount++;
      console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)}`);
    }
  }

  return processedCount;
}

/**
 * Generate AGENTS.md (full copy of CLAUDE.md)
 *
 * opencode.ai recognizes AGENTS.md as the rules file and
 * supports CLAUDE.md as a fallback.
 * Here we output the contents of CLAUDE.md as-is for AGENTS.md.
 */
function generateAgentsMd() {
  const claudeMdPath = path.join(ROOT_DIR, 'CLAUDE.md');

  if (!fs.existsSync(claudeMdPath)) {
    console.log(`  ⚠ CLAUDE.md not found, skipping AGENTS.md generation`);
    return;
  }

  let claudeMdContent = fs.readFileSync(claudeMdPath, 'utf8');

  // Convert title from CLAUDE.md to AGENTS.md
  // Handles patterns "# CLAUDE.md" or "# CLAUDE.md - ..."
  claudeMdContent = claudeMdContent.replace(
    /^# CLAUDE\.md(\s*-\s*.*)?$/m,
    (match, suffix) => `# AGENTS.md${suffix || ''}`
  );

  // Prepend an opencode-compatible header
  const header = `<!-- Generated from CLAUDE.md by build-opencode.js -->
<!-- opencode.ai compatible version of Claude Code Harness -->

`;

  const agentsMd = header + claudeMdContent;

  const destPath = path.join(OPENCODE_DIR, 'AGENTS.md');
  fs.writeFileSync(destPath, agentsMd);
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)} (from CLAUDE.md)`);
}

/**
 * Generate a sample opencode.json
 */
function generateOpencodeJson() {
  const config = {
    "$schema": "https://opencode.ai/config.json",
    "mcp": {
      "harness": {
        "type": "local",
        "enabled": true,
        "command": ["node", "./path/to/claude-code-harness/mcp-server/dist/index.js"]
      }
    }
  };

  const destPath = path.join(OPENCODE_DIR, 'opencode.json');
  fs.writeFileSync(destPath, JSON.stringify(config, null, 2));
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)}`);
}

/**
 * Generate README.md (skipped if it already exists)
 */
function generateReadme() {
  const destPath = path.join(OPENCODE_DIR, 'README.md');

  // Skip if README.md already exists
  if (fs.existsSync(destPath)) {
    console.log(`  ⏭ ${path.relative(ROOT_DIR, destPath)} (already exists, skipped)`);
    return;
  }

  const readme = `# Harness for OpenCode

The opencode.ai-compatible version of Claude Code Harness.

## Setup

### 1. Copy commands and skills to your project

\`\`\`bash
# Clone Harness
git clone https://github.com/tim-hub/powerball-harness.git

# Copy opencode files
cp -r claude-code-harness/opencode/commands/ your-project/.opencode/commands/
cp -r claude-code-harness/opencode/skills/ your-project/.claude/skills/
cp claude-code-harness/opencode/AGENTS.md your-project/AGENTS.md
\`\`\`

### 2. Set up the MCP server (optional)

\`\`\`bash
# Build the MCP server
cd claude-code-harness/mcp-server
npm install
npm run build

# Copy opencode.json to your project and adjust the path
cp claude-code-harness/opencode/opencode.json your-project/
# Edit the path in opencode.json to the actual path
\`\`\`

### 3. Start using it

\`\`\`bash
cd your-project
opencode
\`\`\`

## Available commands

| Command | Description |
|----------|------|
| \`/harness-init\` | Project setup |
| \`/plan-with-agent\` | Create a development plan |
| \`/work\` | Execute tasks |
| \`/harness-review\` | Code review |

## Available skills

opencode.ai automatically recognizes skills in the \`.claude/skills/\` directory:

| Skill | Description |
|--------|------|
| \`notebookLM\` | Document generation (NotebookLM YAML, slides) |
| \`impl\` | Feature implementation |
| \`harness-review\` | Code review |
| \`verify\` | Build verification and error recovery |
| \`auth\` | Authentication and payments (Clerk, Stripe) |
| \`deploy\` | Deployment (Vercel, Netlify) |
| \`ui\` | UI component generation |

## MCP tools

The following tools are available via the MCP server:

| Tool | Description |
|--------|------|
| \`harness_workflow_plan\` | Create a plan |
| \`harness_workflow_work\` | Execute tasks |
| \`harness_workflow_review\` | Code review |
| \`harness_session_broadcast\` | Cross-session notifications |
| \`harness_status\` | Check status |

## Limitations

- The Harness plugin system (\`.claude-plugin/\`) cannot be used with opencode
- Hooks must be configured separately on the opencode side

## Related links

- [Claude Code Harness](https://github.com/tim-hub/powerball-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
`;

  fs.writeFileSync(destPath, readme);
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)}`);
}

/**
 * Copy skills (.claude/skills/ compatible format)
 *
 * opencode.ai recognizes .claude/skills/<name>/SKILL.md.
 * Copy harness skills as-is.
 */
function copySkills() {
  if (!fs.existsSync(SKILLS_DIR)) {
    console.log(`  ⚠ skills/ directory not found, skipping`);
    return 0;
  }

  // Clear the existing skills directory
  clearDir(OPENCODE_SKILLS_DIR);
  ensureDir(OPENCODE_SKILLS_DIR);

  const entries = fs.readdirSync(SKILLS_DIR, { withFileTypes: true });
  let copiedCount = 0;

  const skipSkills = new Set([
    'allow1',
    'breezing',
    'cc-update-review',
    'claude-codex-upstream-update',
    'zz-review-empty',
    'zz-review-escape',
  ]);

  for (const entry of entries) {
    if (!entry.isDirectory()) continue;

    const skillName = entry.name;
    const srcSkillDir = path.join(SKILLS_DIR, skillName);
    const destSkillDir = path.join(OPENCODE_SKILLS_DIR, skillName);

    // Skip test, development, and opencode-incompatible skills
    if (skillName.startsWith('test-') || skillName.startsWith('x-') || skipSkills.has(skillName)) {
      console.log(`  ⏭ ${skillName}/ (dev/test/unsupported skill, skipped)`);
      continue;
    }

    // Check that SKILL.md exists
    const skillMdPath = path.join(srcSkillDir, 'SKILL.md');
    if (!fs.existsSync(skillMdPath)) {
      console.log(`  ⏭ ${skillName}/ (no SKILL.md, skipped)`);
      continue;
    }

    // Recursively copy the skill directory
    copyDirectoryRecursive(srcSkillDir, destSkillDir);
    copiedCount++;
    console.log(`  ✓ ${skillName}/`);
  }

  return copiedCount;
}

/**
 * Recursively copy a directory
 */
function copyDirectoryRecursive(src, dest) {
  ensureDir(dest);

  const entries = fs.readdirSync(src, { withFileTypes: true });

  // Directory/file patterns to exclude
  const excludePatterns = [
    'CLAUDE.md',           // Auto-generated memory context
    'node_modules',        // npm dependencies
    'coverage',            // Test coverage
    '.claude',             // Claude session state
  ];

  // File name prefixes to exclude (startsWith)
  const excludePrefixes = [
    'IMPLEMENTATION_',     // In-progress implementation documents
    'TASK_',               // Task-related documents
  ];

  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    // Exact-match exclusions
    if (excludePatterns.includes(entry.name)) {
      continue;
    }

    // Prefix-based exclusions
    if (excludePrefixes.some(prefix => entry.name.startsWith(prefix))) {
      continue;
    }

    if (entry.isDirectory()) {
      copyDirectoryRecursive(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

/**
 * Main entry point
 */
function main() {
  console.log('🔄 Building opencode version...\n');

  // Clear the opencode directory
  clearDir(OPENCODE_COMMANDS_DIR);
  clearDir(OPENCODE_SKILLS_DIR);
  ensureDir(OPENCODE_DIR);

  // Convert commands (v2.17.0+: commands/ has been migrated to Skills; process only if present)
  console.log('📁 Converting commands:');
  let commandCount = 0;
  if (fs.existsSync(COMMANDS_DIR)) {
    const commandEntries = fs.readdirSync(COMMANDS_DIR);
    if (commandEntries.length === 0) {
      console.log('  ⏭ commands/ is empty (migrated to skills in v2.17.0+)');
    } else {
      commandCount = processDirectory(COMMANDS_DIR, OPENCODE_COMMANDS_DIR);
    }
  } else {
    console.log('  ⏭ commands/ not found (migrated to skills in v2.17.0+)');
  }

  // Convert PM commands (from templates/opencode/commands/)
  console.log('\n📁 Processing PM commands (from templates/opencode/):');
  let pmCount = 0;
  if (fs.existsSync(OPENCODE_TEMPLATES_DIR)) {
    pmCount = processDirectory(OPENCODE_TEMPLATES_DIR, OPENCODE_PM_DIR);
    console.log(`   PM Commands: ${pmCount} files`);
  } else {
    console.log('   ⚠ templates/opencode/commands/ not found, skipping PM commands');
  }

  // Copy skills
  console.log('\n📁 Copying skills:');
  const skillCount = copySkills();

  // Generate additional files
  console.log('\n📄 Generating additional files:');
  generateAgentsMd();
  generateOpencodeJson();
  generateReadme();

  console.log(`\n✅ Done!`);
  console.log(`   Commands: ${commandCount} files`);
  console.log(`   PM Commands: ${pmCount} files`);
  console.log(`   Skills: ${skillCount} directories`);
  console.log(`   Output: ${path.relative(process.cwd(), OPENCODE_DIR)}/`);
}

main();
