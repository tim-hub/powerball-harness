#!/usr/bin/env node
/**
 * build-opencode.js
 *
 * Harness コマンドを opencode.ai 互換形式に変換するスクリプト
 *
 * 変換内容:
 * - commands/ → opencode/commands/ にコピー
 * - frontmatter から description-en を削除
 * - CLAUDE.md → AGENTS.md として生成
 *
 * 使用方法:
 *   node scripts/build-opencode.js
 */

const fs = require('fs');
const path = require('path');

const ROOT_DIR = path.join(__dirname, '..');
const COMMANDS_DIR = path.join(ROOT_DIR, 'commands');
const OPENCODE_DIR = path.join(ROOT_DIR, 'opencode');
const OPENCODE_COMMANDS_DIR = path.join(OPENCODE_DIR, 'commands');

/**
 * ディレクトリを再帰的に作成
 */
function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

/**
 * ディレクトリを再帰的にクリア
 */
function clearDir(dir) {
  if (fs.existsSync(dir)) {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

/**
 * frontmatter を解析
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
 * frontmatter を文字列に変換
 */
function stringifyFrontmatter(frontmatter) {
  const lines = Object.entries(frontmatter)
    .map(([key, value]) => `${key}: ${value}`);
  return `---\n${lines.join('\n')}\n---\n`;
}

/**
 * Harness コマンドを opencode 形式に変換
 */
function convertCommand(content) {
  const { frontmatter, body } = parseFrontmatter(content);

  if (!frontmatter) {
    // frontmatter がない場合はそのまま返す
    return content;
  }

  // opencode で不要なフィールドを削除
  const opencodeFields = ['description-en', 'name'];
  for (const field of opencodeFields) {
    delete frontmatter[field];
  }

  // frontmatter が空になった場合
  if (Object.keys(frontmatter).length === 0) {
    return body;
  }

  return stringifyFrontmatter(frontmatter) + body;
}

/**
 * ディレクトリ内のファイルを再帰的に処理
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
 * AGENTS.md を生成
 */
function generateAgentsMd() {
  const agentsMd = `# AGENTS.md

This project uses [claude-code-harness](https://github.com/Chachamaru127/claude-code-harness) workflow.

## Available Commands

### Core Commands

| Command | Description |
|---------|-------------|
| \`/harness-init\` | Project setup |
| \`/plan-with-agent\` | Create development plan |
| \`/work\` | Execute tasks |
| \`/harness-review\` | Code review |
| \`/sync-status\` | Check project status |

### Optional Commands

| Command | Description |
|---------|-------------|
| \`/harness-update\` | Update Harness |
| \`/mcp-setup\` | Setup MCP server |
| \`/lsp-setup\` | Setup LSP |

## Workflow

\`\`\`
/plan-with-agent → /work → /harness-review → commit
\`\`\`

## MCP Integration

This project includes an MCP server for cross-client communication.
See \`opencode.json\` for configuration.

## More Information

- [Harness Documentation](https://github.com/Chachamaru127/claude-code-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
`;

  const destPath = path.join(OPENCODE_DIR, 'AGENTS.md');
  fs.writeFileSync(destPath, agentsMd);
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)}`);
}

/**
 * opencode.json サンプルを生成
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
 * README.md を生成（既存の場合はスキップ）
 */
function generateReadme() {
  const destPath = path.join(OPENCODE_DIR, 'README.md');

  // 既存の README.md がある場合はスキップ
  if (fs.existsSync(destPath)) {
    console.log(`  ⏭ ${path.relative(ROOT_DIR, destPath)} (already exists, skipped)`);
    return;
  }

  const readme = `# Harness for OpenCode

Claude Code Harness の opencode.ai 互換版です。

## セットアップ

### 1. コマンドをプロジェクトにコピー

\`\`\`bash
# Harness をクローン
git clone https://github.com/Chachamaru127/claude-code-harness.git

# opencode 用コマンドをコピー
cp -r claude-code-harness/opencode/commands/ your-project/.opencode/commands/
cp claude-code-harness/opencode/AGENTS.md your-project/AGENTS.md
\`\`\`

### 2. MCP サーバーをセットアップ（オプション）

\`\`\`bash
# MCP サーバーをビルド
cd claude-code-harness/mcp-server
npm install
npm run build

# opencode.json をプロジェクトにコピーしてパスを調整
cp claude-code-harness/opencode/opencode.json your-project/
# opencode.json 内のパスを実際のパスに変更
\`\`\`

### 3. 利用開始

\`\`\`bash
cd your-project
opencode
\`\`\`

## 利用可能なコマンド

| コマンド | 説明 |
|----------|------|
| \`/harness-init\` | プロジェクトセットアップ |
| \`/plan-with-agent\` | 開発プラン作成 |
| \`/work\` | タスク実行 |
| \`/harness-review\` | コードレビュー |

## MCP ツール

MCP サーバー経由で以下のツールが利用可能です：

| ツール | 説明 |
|--------|------|
| \`harness_workflow_plan\` | プラン作成 |
| \`harness_workflow_work\` | タスク実行 |
| \`harness_workflow_review\` | コードレビュー |
| \`harness_session_broadcast\` | セッション間通知 |
| \`harness_status\` | 状態確認 |

## 制限事項

- Harness プラグインシステム（\`.claude-plugin/\`）は opencode では使用できません
- フックは opencode 側で別途設定が必要です

## 関連リンク

- [Claude Code Harness](https://github.com/Chachamaru127/claude-code-harness)
- [OpenCode Documentation](https://opencode.ai/docs/)
`;

  fs.writeFileSync(destPath, readme);
  console.log(`  ✓ ${path.relative(ROOT_DIR, destPath)}`);
}

/**
 * メイン処理
 */
function main() {
  console.log('🔄 Building opencode version...\n');

  // opencode ディレクトリをクリア
  clearDir(OPENCODE_COMMANDS_DIR);
  ensureDir(OPENCODE_DIR);

  // コマンドを変換
  console.log('📁 Converting commands:');
  const count = processDirectory(COMMANDS_DIR, OPENCODE_COMMANDS_DIR);

  // 追加ファイルを生成
  console.log('\n📄 Generating additional files:');
  generateAgentsMd();
  generateOpencodeJson();
  generateReadme();

  console.log(`\n✅ Done! Converted ${count} command files.`);
  console.log(`   Output: ${path.relative(process.cwd(), OPENCODE_DIR)}/`);
}

main();
