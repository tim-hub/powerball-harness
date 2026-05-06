package guardrail

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/tim-hub/powerball-harness/go/pkg/config"
	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

const (
	protectedBranchPushPolicyAsk   = "ask"
	protectedBranchPushPolicyDeny  = "deny"
	protectedBranchPushPolicyAllow = "allow"
)

func normalizeProtectedBranchPushPolicy(value string) string {
	normalized := strings.ToLower(strings.Trim(strings.TrimSpace(value), `"'`))
	switch normalized {
	case protectedBranchPushPolicyAsk, "confirm":
		return protectedBranchPushPolicyAsk
	case protectedBranchPushPolicyDeny, "block":
		return protectedBranchPushPolicyDeny
	case protectedBranchPushPolicyAllow, "approve":
		return protectedBranchPushPolicyAllow
	default:
		return protectedBranchPushPolicyAsk
	}
}

func resolveProtectedBranchPushPolicy(input hookproto.HookInput, projectRoot string) string {
	for _, envName := range []string{
		"HARNESS_PROTECTED_BRANCH_PUSH_POLICY",
		"HARNESS_DIRECT_PUSH_POLICY",
	} {
		if value := os.Getenv(envName); value != "" {
			return normalizeProtectedBranchPushPolicy(value)
		}
	}

	if value := readProtectedBranchPushPolicyFromYAML(projectRoot); value != "" {
		return normalizeProtectedBranchPushPolicy(value)
	}

	if value := readProtectedBranchPushPolicyFromHarnessTOML(filepath.Join(projectRoot, "harness.toml")); value != "" {
		return normalizeProtectedBranchPushPolicy(value)
	}

	if input.PluginRoot != "" && input.PluginRoot != projectRoot {
		if value := readProtectedBranchPushPolicyFromHarnessTOML(filepath.Join(input.PluginRoot, "harness.toml")); value != "" {
			return normalizeProtectedBranchPushPolicy(value)
		}
	}

	return protectedBranchPushPolicyAsk
}

func readProtectedBranchPushPolicyFromHarnessTOML(path string) string {
	cfg, err := config.ParseFile(path)
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.Safety.Permissions.ProtectedBranchPush
}

func readProtectedBranchPushPolicyFromYAML(projectRoot string) string {
	configPath := filepath.Join(projectRoot, ".claude-code-harness.config.yaml")
	f, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inSafety := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if value, ok := parseSimpleYAMLValue(trimmed, "protected_branch_push"); ok {
				return value
			}
			if value, ok := parseSimpleYAMLValue(trimmed, "protectedBranchPush"); ok {
				return value
			}
			inSafety = strings.HasPrefix(trimmed, "safety:")
			continue
		}
		if !inSafety {
			continue
		}
		if value, ok := parseSimpleYAMLValue(trimmed, "protected_branch_push"); ok {
			return value
		}
		if value, ok := parseSimpleYAMLValue(trimmed, "protectedBranchPush"); ok {
			return value
		}
	}
	return ""
}

func parseSimpleYAMLValue(line, key string) (string, bool) {
	prefix := key + ":"
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if idx := strings.Index(value, "#"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	return strings.Trim(value, `"'`), true
}
