# Memory Optimization

## Memory Optimization (CC 2.1.49+)

Since Claude Code 2.1.49, memory usage on session resume has been **reduced by 68%**.

### Best Practices for Long Session Management

| Workload | Recommended Strategy |
|----------|---------------------|
| **Normal implementation** | Resume with `--resume` every 1-2 hours |
| **Large-scale refactoring** | Split sessions by feature unit, use `--resume` for each |
| **Parallel tasks** | Run in parallel with `harness-work --parallel`, use `--resume` midway for long sessions |
| **Memory warning** | Resume immediately with `--resume` (faster than before) |

### Auto-generated Session Names (CC 2.1.41+)

Running `/rename` without arguments auto-generates a session name from the conversation context.
This makes it easier to identify sessions in long-running or `--resume`-heavy workflows.

### Efficient Workflow Example

```bash
# Implementation phase 1
claude "Implement authentication feature"
# -> 1 hour later

# Resume session (memory-efficient)
claude --resume "Add password reset feature"
# -> 1 hour later

# Resume again
claude --resume "Add tests"
```

### Memory Management Recommendations

| Recommendation | Reason |
|---------------|--------|
| **Actively resume sessions** | Low resume cost with 68% memory reduction |
| **Resume periodically** | Keeps context organized and maintains focus |
| **Split by feature unit** | Break large tasks into smaller chunks for resuming |
| **Use Plans.md** | Smooth handoff when resuming |

> 💡 Memory efficiency has been significantly improved, so actively take advantage of session resumption.
