package main

import (
	"path/filepath"
	"testing"
)

// TestSync_NoPhantomFields verifies that pluginJSON struct does not emit
// fields outside the official plugins-reference schema.
// Phase 45.3.1 (Phase 44 follow-up) — guards against re-introducing
// monitors/agents fields in plugin.json that were removed in 45.1.1/45.2.1.
func TestSync_NoPhantomFields(t *testing.T) {
	dir := setupProjectDir(t, fullTOML)
	runSync([]string{dir})

	v := readJSON(t, filepath.Join(dir, ".claude-plugin", "plugin.json"))

	// monitors must NOT appear in plugin.json (use monitors/monitors.json instead)
	if _, ok := v["monitors"]; ok {
		t.Error("plugin.json must not contain monitors field — use monitors/monitors.json (official SSOT) instead")
	}
	// agents must NOT appear in plugin.json (use agents/ auto-discovery instead)
	if _, ok := v["agents"]; ok {
		t.Error("plugin.json must not contain agents field — agents/ directory auto-discovery is the official method")
	}
}
