#!/usr/bin/env node
/**
 * build-opencode.js
 *
 * Script to convert Harness commands to opencode.ai compatible format
 *
 * Conversion:
 * - commands/ -> opencode/commands/ copy
 * - Remove description-en from frontmatter
 * - Generate CLAUDE.md as AGENTS.md
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
 * Create directory recursively
 */
function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

/**
 * Clear directory recursively
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
 * Convert frontmatter to string
 */
function stringifyFrontmatter(frontmatter) {
  const lines = Object.entries(frontmatter)
    .map(([key, value]) => `${key}: ${value}`);
  return `---\n${lines.join('\n')}\n---\n`;
}

/**
 * Convert Harness command to opencode format
 */
function convertCommand(content) {
  const { frontmatter, body } = parseFrontmatter(content);

  if (!frontmatter) {
    // Return as-is if no frontmatter
    return content;
  }

  // Remove fields unnecessary for opencode
  const opencodeFields = ['description-en', 'name'];
  for (const field of opencodeFields) {
    delete frontmatter[field];
  }

  // If frontmatter becomes empty
  if (Object.keys(frontmatter).length === 0) {
    return body;
  }

  return stringifyFrontmatter(frontmatter) + body;
}

/**
 * Process files in directory recursively
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
 * opencode.ai recognizes AGENTS.md as a rules file,
 * with CLAUDE.md as fallback.
 * Here we output CLAUDE.md content as-is as AGENTS.md.
 */
function generateAgentsMd() {
  const claudeMdPath = path.join(ROOT_DIR, 'CLAUDE.md');

  if (!fs.existsSync(claudeMdPath)) {
    console.log(`  ⚠ CLAUDE.md not found, skipping AGENTS.md generation`);
    return;
  }

  let claudeMdContent = fs.readFileSync(claudeMdPath, 'utf8');

  // Convert title from CLAUDE.md to AGENTS.md
  // Handle "# CLAUDE.md" or "# CLAUDE.md - ..." patterns
  claudeMdContent = claudeMdContent.replace(
    /^# CLAUDE\.md(\s*-\s*.*)?$/m,
    (match, suffix) => `# AGENTS.md${suffix || ''}`
  );

  // Add opencode-compatible header
  const header = `<!-- Generated from CLAUDE.md by build-opencode.js -->
<!-- opencode.ai compatible version of Claude Code Harness -->

`;

  const agentsMd = header + claudeMdContent;

  const destPath = path.join(OPENCODE_DIR, 'AGENTS.md');
  fs.writeFileSync(destPath, agentsMd);
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)} (from CLAUDE.md)`);
}

/**
 * Generate opencode.json sample
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
 * Generate README.md (skip if exists)
 */
function generateReadme() {
  const destPath = path.join(OPENCODE_DIR, 'README.md');

  // Skip if README.md already exists
  if (fs.existsSync(destPath)) {
    console.log(`  ⏭ ${path.relative(ROOT_DIR, destPath)} (already exists, skipped)`);
    return;
  }

  const readme = `# Harness for OpenCode

opencode.ai compatible version of Claude Code Harness.

## Setup

### 1. Copy commands and skills to your project

\`\`\`bash
# Clone Harness
git clone https://github.com/Chachamaru127/claude-code-harness.git

# Copy files for opencode
cp -r claude-code-harness/opencode/commands/ your-project/.opencode/commands/
cp -r claude-code-harness/opencode/skills/ your-project/.claude/skills/
cp claude-code-harness/opencode/AGENTS.md your-project/AGENTS.md
\`\`\`

### 2. Set up MCP server (optional)

\`\`\`bash
# Build MCP server
cd claude-code-harness/mcp-server
npm install
npm run build

# Copy opencode.json to project and adjust paths
cp claude-code-harness/opencode/opencode.json your-project/
# Update paths in opencode.json to actual paths
\`\`\`

### 3. Start using

\`\`\`bash
cd your-project
opencode
\`\`\`

## Available Commands

| Command | Description |
|---------|-------------|
| \`/harness-init\` | Project setup |
| \`/plan-with-agent\` | Development plan creation |
| \`/work\` | Task execution |
| \`/harness-review\` | Code review |

## Available Skills

opencode.ai auto-discovers skills in the \`.claude/skills/\` directory:

| Skill | Description |
|-------|-------------|
| \`notebookLM\` | Document generation (NotebookLM YAML, slides) |
| \`impl\` | Feature implementation |
| \`harness-review\` | Code review |
| \`verify\` | Build verification and error recovery |
| \`auth\` | Auth and payments (Clerk, Stripe) |
| \`deploy\` | Deploy (Vercel, Netlify) |
| \`ui\` | UI component generation |

## MCP Tools

The following tools are available via the MCP server:

| Tool | Description |
|-------|-------------|
| \`harness_workflow_plan\` | Plan creation |
| \`harness_workflow_work\` | Task execution |
| \`harness_workflow_review\` | Code review |
| \`harness_session_broadcast\` | Inter-session notification |
| \`harness_status\` | Status check |

## Limitations

- Harness plugin system (\`.claude-plugin/\`) cannot be used with opencode
- Hooks need separate configuration on the opencode side

## Related Links

- [Claude Code Harness](https://github.com/Chachamaru127/claude-code-harness)
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

  // Clear existing skills directory
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

    // Skip test/dev/unsupported skills for opencode
    if (skillName.startsWith('test-') || skillName.startsWith('x-') || skipSkills.has(skillName)) {
      console.log(`  ⏭ ${skillName}/ (dev/test/unsupported skill, skipped)`);
      continue;
    }

    // Check if SKILL.md exists
    const skillMdPath = path.join(srcSkillDir, 'SKILL.md');
    if (!fs.existsSync(skillMdPath)) {
      console.log(`  ⏭ ${skillName}/ (no SKILL.md, skipped)`);
      continue;
    }

    // Copy skill directory recursively
    copyDirectoryRecursive(srcSkillDir, destSkillDir);
    copiedCount++;
    console.log(`  ✓ ${skillName}/`);
  }

  return copiedCount;
}

/**
 * Copy directory recursively
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

  // File name patterns to exclude (startsWith)
  const excludePrefixes = [
    'IMPLEMENTATION_',     // In-progress documents
    'TASK_',               // Task-related documents
  ];

  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    // Exclude by exact match
    if (excludePatterns.includes(entry.name)) {
      continue;
    }

    // Exclude by prefix
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
 * Main processing
 */
function main() {
  console.log('🔄 Building opencode version...\n');

  // Clear opencode directory
  clearDir(OPENCODE_COMMANDS_DIR);
  clearDir(OPENCODE_SKILLS_DIR);
  ensureDir(OPENCODE_DIR);

  // Convert commands (v2.17.0+: commands/ migrated to Skills, process only if present)
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
