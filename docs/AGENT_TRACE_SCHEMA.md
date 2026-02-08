# AgentTrace Schema Documentation

## Overview

AgentTrace is a lightweight metadata format for recording AI code generation provenance. It provides attribution tracking for Claude Code plugins.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.1.0 | 2026-02-03 | Initial implementation with basic metadata |
| 0.2.0 | 2026-02-04 | Added Attribution field for plugin tracking |
| 0.3.0 | 2026-02-04 | Added Task tool metrics (CC 2.1.30+) |

---

## Schema v0.3.0 (Current)

### Overview

AgentTrace v0.3.0 adds support for Task tool execution metrics, enabling cost tracking and performance analysis for parallel subagent workflows.

### Full Schema

```typescript
interface AgentTrace {
  version: "0.3.0";
  id: string;                  // UUID v4
  timestamp: string;           // ISO 8601 UTC
  tool: "Edit" | "Write" | "Task";  // Tool that triggered the trace
  files: FileChange[];              // Files modified (empty for Task tool)

  // v0.3.0: Task tool metrics (CC 2.1.30+)
  metrics?: {
    tokenCount?: number;       // Total tokens consumed
    toolUses?: number;         // Number of tool invocations
    duration?: number;         // Execution time (milliseconds)
  };

  // v0.2.0: Attribution
  attribution?: {
    plugin: string;            // Plugin name (e.g., "claude-code-harness")
    version: string;           // Plugin version
    license?: string;          // Plugin license
    author?: string;           // Plugin author
  };

  // VCS information
  vcs?: {
    revision: string;          // Git commit hash
    branch: string;            // Git branch name
    dirty: boolean;            // Uncommitted changes exist
  };

  // Project metadata
  metadata: {
    project: string;           // Project name
    projectType: string;       // Detected project type
    sessionId?: string;        // Claude Code session ID
    taskId?: string;           // Task tool invocation ID (v0.3.0+)
  };
}

interface FileChange {
  path: string;                // Relative to repo root
  action: "create" | "modify"; // Operation type
  range: string;               // Currently "unknown"
}
```

### Storage Format

AgentTrace records are stored in **JSONL** format (one JSON object per line):

```
.claude/state/agent-trace.jsonl
```

**File Permissions**: `0600` (owner read/write only) for security.

**Rotation**: Files are rotated at 10MB to prevent unbounded growth.

---

## v0.3.0 New Feature: Task Tool Metrics

### Overview

Claude Code 2.1.30 introduced metrics in Task tool results:
- `tokenCount` - Total tokens consumed by the subagent
- `toolUses` - Number of tool invocations
- `duration` - Execution time in milliseconds

### How to Capture Metrics

#### From Task Tool Result

When a Task tool call completes, the result includes metrics:

```typescript
// Task tool result structure (CC 2.1.30+)
{
  "content": [...],
  "metrics": {
    "tokenCount": 2340,
    "toolUses": 12,
    "duration": 45000  // 45 seconds
  }
}
```

#### Integration Pattern

```javascript
// In hook handler or skill
const taskResult = await Task({
  subagent_type: "task-worker",
  prompt: "Implement feature X"
});

// Extract metrics
const metrics = taskResult.metrics || {};

// Store in AgentTrace
const trace = {
  version: "0.3.0",
  id: generateUUID(),
  timestamp: new Date().toISOString(),
  tool: "Task",
  metrics: {
    tokenCount: metrics.tokenCount,
    toolUses: metrics.toolUses,
    duration: metrics.duration
  },
  metadata: {
    project: "my-project",
    projectType: "nextjs",
    taskId: "task-123"  // Track which task generated this
  }
};
```

---

## Use Cases

### 1. Cost Tracking (Parallel Workflows)

When using `/work all` with parallel task-workers, aggregate metrics across all workers:

```markdown
🚀 /work all 完了
├─ Worker 1: ✅ 2,340 tokens | 12 tools | 45s
├─ Worker 2: ✅ 1,890 tokens | 8 tools | 32s
├─ Worker 3: ✅ 3,120 tokens | 15 tools | 58s
└─ 合計: 7,350 tokens | 35 tools | 135s (2m 15s)
```

**Implementation**:

```javascript
// Aggregate metrics from multiple task-workers
function aggregateMetrics(taskResults) {
  const totals = { tokenCount: 0, toolUses: 0, duration: 0 };

  for (const result of taskResults) {
    if (result.metrics) {
      totals.tokenCount += result.metrics.tokenCount || 0;
      totals.toolUses += result.metrics.toolUses || 0;
      totals.duration += result.metrics.duration || 0;
    }
  }

  return totals;
}
```

### 2. Performance Analysis

Track execution time trends for different task types:

```sql
-- Example query (if AgentTrace is imported to analytics DB)
SELECT
  metadata->>'projectType' as project_type,
  AVG(metrics->>'duration') as avg_duration_ms,
  AVG(metrics->>'tokenCount') as avg_tokens
FROM agent_traces
WHERE version = '0.3.0'
  AND metrics IS NOT NULL
GROUP BY project_type;
```

### 3. Efficiency Monitoring

Detect inefficient patterns:

```javascript
// Flag tasks with high token-to-duration ratio (token wastage)
function detectInefficiency(trace) {
  if (!trace.metrics) return null;

  const tokensPerSecond = trace.metrics.tokenCount / (trace.metrics.duration / 1000);

  if (tokensPerSecond > 100) {
    return "⚠️ High token consumption rate - consider optimizing prompts";
  }

  if (trace.metrics.toolUses > 50) {
    return "⚠️ Excessive tool uses - task may be stuck in loop";
  }

  return null;
}
```

---

## Migration Guide

### From v0.2.0 to v0.3.0

**Backwards Compatible**: v0.3.0 is backwards compatible. Existing traces without `metrics` remain valid.

**Reading v0.2.0 traces**:

```javascript
function readTrace(line) {
  const trace = JSON.parse(line);

  // Handle both versions
  const metrics = trace.metrics || {
    tokenCount: null,
    toolUses: null,
    duration: null
  };

  return { ...trace, metrics };
}
```

### Schema Version Detection

```javascript
function getSchemaVersion(trace) {
  if (trace.metrics !== undefined) return "0.3.0";
  if (trace.attribution !== undefined) return "0.2.0";
  return "0.1.0";
}
```

---

## Implementation Notes

### Where to Emit AgentTrace

AgentTrace is currently emitted by:

1. **PostToolUse Hook** - Automatic emission on Write/Edit tools
   - Location: `hooks/hooks.json` → `PostToolUse` → `emit-agent-trace.js`
   - Captures: Direct file modifications

2. **Task Tool Hook** - Automatic emission via PostToolUse hook for Task tool
   - Location: `hooks/hooks.json` → `PostToolUse` → `emit-agent-trace.js`
   - Captures: Subagent execution metrics (tokenCount, toolUses, duration)

### emit-agent-trace.js Extension

To support v0.3.0 metrics, `emit-agent-trace.js` needs to:

1. Detect if the tool is `Task`
2. Parse `CLAUDE_TOOL_RESULT` environment variable for metrics
3. Add metrics to the AgentTrace record

**Example**:

```javascript
// In emit-agent-trace.js
function main() {
  const toolName = process.env.CLAUDE_TOOL_NAME;
  const toolResult = process.env.CLAUDE_TOOL_RESULT;

  // ... existing code ...

  const record = {
    version: "0.3.0",
    id: generateUUID(),
    timestamp: getTimestamp(),
    tool: toolName,
    files: files
  };

  // v0.3.0: Extract metrics from Task tool result
  if (toolName === "Task" && toolResult) {
    try {
      const result = JSON.parse(toolResult);
      if (result.metrics) {
        record.metrics = {
          tokenCount: result.metrics.tokenCount,
          toolUses: result.metrics.toolUses,
          duration: result.metrics.duration
        };
      }
    } catch (err) {
      logError("parseMetrics", err);
    }
  }

  // ... rest of the code ...
}
```

---

## Security Considerations

### Metrics Data Privacy

Task tool metrics may indirectly reveal:
- Project complexity (high token counts)
- Development patterns (tool usage patterns)
- Performance characteristics (duration patterns)

**Recommendation**: Treat `agent-trace.jsonl` as sensitive and exclude from public repositories.

### Recommended .gitignore

```gitignore
# AgentTrace files (contains metrics and project metadata)
.claude/state/agent-trace.jsonl*
```

---

## Future Considerations

### Potential v0.4.0 Features

- **Error tracking**: Capture Task tool failures and error messages
- **Cost calculation**: Estimate API costs based on tokenCount
- **Parallel tracking**: Link related traces from parallel workflows
- **Context window usage**: Track context window utilization percentage

### Integration Opportunities

- **harness-ui**: Visualize metrics in dashboard
- **CI/CD**: Track metrics across builds for performance regression detection
- **Analytics**: Export to BI tools for long-term trend analysis

---

## References

- [Claude Code 2.1.30 Release Notes](https://code.claude.com/docs/en/changelog)
- [emit-agent-trace.js](../scripts/emit-agent-trace.js) - Current implementation
- [Task tool documentation](https://code.claude.com/docs/en/sdk/task-tool)
- [CHANGELOG.md](../CHANGELOG.md) - Version history
