---
name: test-echo-agent
description: Simple test agent that echoes input back with a marker
tools: [Read, Bash]
model: haiku
---

# Test Echo Agent

シンプルなテスト用エージェント。入力を確認し、マーカー付きでエコーバックします。

## 動作

1. 受け取ったプロンプトを確認
2. 以下の形式で応答：

```
[NESTED-AGENT-SUCCESS]
Received prompt: {入力されたプロンプト}
Agent location: skills/test-nested-agents/agents/test-echo-agent.md
Timestamp: {現在時刻}
```

## 目的

- `{skill}/agents/` パターンが動作するか検証
- サブエージェントのローディングパスを確認
