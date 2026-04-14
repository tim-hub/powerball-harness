# Makefile — powerball-harness local development helpers
#
# Usage:
#   make           → show this help
#   make test      → run validate + consistency check (recommended pre-commit)
#   make validate  → full plugin validation (tests/validate-plugin.sh)
#   make check     → consistency check (local-scripts/check-consistency.sh)
#   make lint      → migration residue + skill description audit
#   make build     → build Go guardrail binary for current platform
#   make bench     → run breezing benchmark suite (BENCH_TASK=1 BENCH_ITER=3 by default)
#   make version   → check VERSION / harness.toml sync

.PHONY: help test validate check lint build bench version

# Default target: show help
help:
	@echo "powerball-harness dev targets"
	@echo ""
	@echo "  make test       Run validate + check (recommended before commit)"
	@echo "  make validate   Full plugin validation (tests/validate-plugin.sh)"
	@echo "  make check      Consistency check (local-scripts/check-consistency.sh)"
	@echo "  make lint       Residue scan + skill description audit"
	@echo "  make build      Build Go guardrail binary for current platform"
	@echo "  make bench      Run breezing benchmark (BENCH_TASK=1 BENCH_ITER=3 by default)"
	@echo "  make version    Check VERSION / harness.toml are in sync"

# Run both validate and consistency check — recommended pre-submit gate
test: validate check

# Full plugin validation: structure, hooks, skill format, residue, CI consistency
validate:
	@echo "▶ Running validate-plugin.sh…"
	bash ./tests/validate-plugin.sh

# Consistency check: templates, versions, skill references, docs
check:
	@echo "▶ Running check-consistency.sh…"
	bash ./local-scripts/check-consistency.sh

# Lint: migration residue scan + skill description format audit
lint:
	@echo "▶ Running check-residue.sh…"
	bash ./local-scripts/check-residue.sh
	@echo ""
	@echo "▶ Running audit-skill-descriptions.sh…"
	bash ./local-scripts/audit-skill-descriptions.sh

# Build the Go guardrail binary for the current platform
build:
	@echo "▶ Building harness binary…"
	bash ./harness/skills/harness-setup/scripts/build-binary.sh

# Run breezing benchmark — override defaults with: make bench BENCH_TASK=2 BENCH_ITER=5
BENCH_TASK ?= 1
BENCH_ITER ?= 3
bench:
	@echo "▶ Running breezing benchmark (task $(BENCH_TASK), $(BENCH_ITER) iterations)…"
	bash ./benchmarks/breezing-bench/run.sh --task $(BENCH_TASK) --iterations $(BENCH_ITER) --mode both

# Check VERSION / harness.toml version sync
version:
	@echo "▶ Checking version sync…"
	bash ./harness/scripts/sync-version.sh check
