# Makefile — powerball-harness local development helpers
#
# Usage:
#   make              → show this help
#   make test         → run validate + consistency check (recommended pre-commit)
#   make validate     → full plugin validation (tests/validate-plugin.sh)
#   make check        → consistency check (local-scripts/check-consistency.sh)
#   make lint         → migration residue + skill description audit
#   make build        → build Go guardrail binary for current platform (into harness/bin/)
#   make build-all    → cross-compile for darwin-arm64, darwin-amd64, linux-amd64
#   make bench        → run breezing benchmark suite (BENCH_TASK=1 BENCH_ITER=3 by default)
#   make test-all       → run every tests/test-*.sh (includes test-harness.sh for the harness group)
#   make test-harness   → run harness-internal tests (harness/tests/test-harness.sh)
#   make check-version      → check VERSION / all metadata files in sync
#   make sync-version       → sync all metadata files to VERSION
#   make check-version-bump → release metadata policy check (CI gate)
#   make codex-test   → multi-agent migration guards (tests/test-codex-package.sh)

.PHONY: help test test-all test-harness validate check lint build build-all bench check-version sync-version check-version-bump codex-test

# Default target: show help
help:
	@echo "powerball-harness dev targets"
	@echo ""
	@echo "  make test          Run validate + check (recommended before commit)"
	@echo "  make validate      Full plugin validation (tests/validate-plugin.sh)"
	@echo "  make check         Consistency check (harness/skills/harness-release/scripts/check-consistency.sh)"
	@echo "  make lint          Residue scan + skill description audit"
	@echo "  make build         Build Go binary for current platform (into harness/bin/)"
	@echo "  make build-all     Cross-compile for darwin-arm64, darwin-amd64, linux-amd64"
	@echo "  make bench         Run breezing benchmark (BENCH_TASK=1 BENCH_ITER=3 by default)"
	@echo "  make test-all      Run every tests/test-*.sh (test-harness.sh covers harness group)"
	@echo "  make test-harness  Run harness-internal tests (/tests/test-harness.sh)"
	@echo "  make check-version       Check VERSION / all metadata files are in sync"
	@echo "  make sync-version        Sync all metadata files to VERSION"
	@echo "  make check-version-bump  Release metadata policy check (for CI / pre-release)"
	@echo "  make codex-test    Multi-agent migration guard tests"

# Run every tests/test-*.sh and harness/tests/test-*.sh and report a pass/fail summary
test-all:
	@passed=0; failed=0; failures=""; \
	for script in tests/test-*.sh; do \
		[ -f "$$script" ] || continue; \
		if [ "$$CI" = "true" ]; then \
			case "$$script" in \
				tests/test-agent-telemetry-summary.sh|tests/test-structured-handoff-artifact.sh) \
					printf "  %-55s" "$$script"; echo "SKIP (CI)"; continue ;; \
			esac; \
		fi; \
		printf "  %-55s" "$$script"; \
		if bash "$$script" >/dev/null 2>&1; then \
			echo "PASS"; passed=$$((passed + 1)); \
		else \
			echo "FAIL"; failed=$$((failed + 1)); failures="$$failures\n    $$script"; \
		fi; \
	done; \
	echo ""; \
	echo "Results: $$passed passed, $$failed failed"; \
	if [ $$failed -gt 0 ]; then \
		printf "Failed scripts:$$failures\n"; \
		exit 1; \
	fi

# Run harness-internal tests (tests for scripts in harness/scripts/)
test-harness:
	@echo "▶ Running tests/test-harness.sh…"
	bash ./tests/test-harness.sh

# Run both validate and consistency check — recommended pre-submit gate
test: validate check

# Full plugin validation: structure, hooks, skill format, residue, CI consistency
validate:
	@echo "▶ Running validate-plugin.sh…"
	bash ./tests/validate-plugin.sh

# Consistency check: templates, versions, skill references, docs
check:
	@echo "▶ Running check-consistency.sh…"
	bash ./harness/skills/harness-release/scripts/check-consistency.sh

# Lint: migration residue scan + skill description format audit
lint:
	@echo "▶ Running check-residue.sh…"
	bash ./harness/skills/harness-release/scripts/check-residue.sh
	@echo ""
	@echo "▶ Running audit-skill-descriptions.sh…"
	bash ./local-scripts/audit-skill-descriptions.sh

# Build the Go guardrail binary for the current platform (into harness/bin/)
build:
	@echo "▶ Building harness binary for current platform…"
	bash ./local-scripts/build-binary.sh

# Cross-compile for all supported platforms (darwin-arm64, darwin-amd64, linux-amd64)
build-all:
	@echo "▶ Building harness binaries for all platforms…"
	cd go && bash scripts/build-all.sh

# Run breezing benchmark — override defaults with: make bench BENCH_TASK=2 BENCH_ITER=5
BENCH_TASK ?= 1
BENCH_ITER ?= 3
bench:
	@echo "▶ Running breezing benchmark (task $(BENCH_TASK), $(BENCH_ITER) iterations)…"
	bash ./benchmarks/breezing-bench/run.sh --task $(BENCH_TASK) --iterations $(BENCH_ITER) --mode both

# Check VERSION / all metadata files are in sync
check-version:
	@echo "▶ Checking version sync…"
	bash ./harness/skills/harness-release/scripts/sync-version.sh check

# Sync all metadata files to VERSION (harness.toml, template-registry.json, *.md.template)
sync-version:
	@echo "▶ Syncing metadata files to VERSION…"
	bash ./harness/skills/harness-release/scripts/sync-version.sh sync

# Release metadata policy check (runs in CI on every PR; use locally before release)
check-version-bump:
	@echo "▶ Checking release metadata policy…"
	bash ./local-scripts/check-version-bump.sh

# Multi-agent migration guard tests
codex-test:
	@echo "▶ Running codex package tests…"
	bash ./tests/test-codex-package.sh
