#!/usr/bin/env bash
#
# codex-worker-quality-gate.sh
# Quality verification of Worker output by the Orchestrator
#
# Usage:
#   ./scripts/codex-worker-quality-gate.sh --worktree PATH [--skip-gate GATE --reason TEXT]
#

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=lib/codex-worker-common.sh
source "$SCRIPT_DIR/lib/codex-worker-common.sh"

# ============================================
# Local configuration
# ============================================
GATE_SKIP_LOG=".claude/state/gate-skips.log"
AGENTS_SUMMARY_PATTERN='AGENTS_SUMMARY:[[:space:]]*(.+?)[[:space:]]*\|[[:space:]]*HASH:([A-Fa-f0-9]{8})'

# Record skip log
log_skip() {
    local gate="$1"
    local reason="$2"
    local user="${USER:-unknown}"

    mkdir -p "$(dirname "$GATE_SKIP_LOG")"
    printf '%s\t%s\t%s\t%s\n' "$(now_utc)" "$gate" "$reason" "$user" >> "$GATE_SKIP_LOG"
}

# Resolve diff base (merge-base with default branch, fallback to HEAD~1, then HEAD)
resolve_diff_base() {
    local worktree="$1"
    local default_branch
    default_branch=$(get_default_branch)

    local diff_base
    diff_base=$(cd "$worktree" && git merge-base HEAD "origin/$default_branch" 2>/dev/null || git rev-parse HEAD~1 2>/dev/null || echo "")

    if [[ -z "$diff_base" ]]; then
        diff_base=$(cd "$worktree" && git rev-parse HEAD 2>/dev/null || echo "")
    fi

    echo "$diff_base"
}

# Get added lines in the specified range (returned as-is with newlines)
collect_added_lines() {
    local worktree="$1"
    local diff_base="$2"

    cd "$worktree" && git diff "$diff_base"..HEAD --unified=0 -- \
        '*.ts' '*.tsx' '*.js' '*.jsx' '*.mjs' '*.cjs' \
        '*.sh' '*.md' '*.json' '*.yaml' '*.yml' '*.toml' \
        '*.py' '*.go' '*.rs' '*.txt' 2>/dev/null | grep '^+' | grep -v '^+++ ' || true
}

# Get changed files in the specified range
collect_changed_files() {
    local worktree="$1"
    local diff_base="$2"

    cd "$worktree" && git diff --name-only "$diff_base"..HEAD -- \
        '*.ts' '*.tsx' '*.js' '*.jsx' '*.mjs' '*.cjs' \
        '*.sh' '*.md' '*.json' '*.yaml' '*.yml' '*.toml' \
        '*.py' '*.go' '*.rs' '*.txt' 2>/dev/null || true
}

# Gate 1: Evidence verification
gate_evidence() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 1: evidence verification"

    # Compute AGENTS.md hash
    local agents_file="$worktree/AGENTS.md"
    if [[ ! -f "$agents_file" ]]; then
        echo '{"status": "critical", "details": "AGENTS.md not found"}' > "$output_file"
        return 2  # Critical: AGENTS.md is required
    fi

    # Compute hash (identical to engine: strip BOM, strip CR, first 8 chars of SHA256)
    local expected_hash
    expected_hash=$(calculate_file_hash "$agents_file" 8)

    # Search for AGENTS_SUMMARY in Worker output
    # 1. Worker output log (priority)
    # 2. Latest commit message (fallback)
    local worker_output_log="$worktree/.claude/state/worker-output.log"
    local search_content=""
    local found_in_log=false

    if [[ -f "$worker_output_log" ]]; then
        search_content=$(cat "$worker_output_log")
        if echo "$search_content" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
            found_in_log=true
        fi
    fi

    # If not found in log, fall back to commit message
    if [[ "$found_in_log" == "false" ]]; then
        local commit_msg
        commit_msg=$(cd "$worktree" && git log -1 --pretty=%B 2>/dev/null || echo "")
        if echo "$commit_msg" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
            search_content="$commit_msg"
        fi
    fi

    # Pattern match (extract HASH only from AGENTS_SUMMARY lines)
    if echo "$search_content" | grep -qE "$AGENTS_SUMMARY_PATTERN"; then
        local found_hash
        # Extract HASH only from lines containing AGENTS_SUMMARY (avoid unrelated HASHes)
        found_hash=$(echo "$search_content" | grep -E 'AGENTS_SUMMARY' | grep -oE 'HASH:[A-Fa-f0-9]{8}' | head -1 | cut -d: -f2)

        if [[ "${found_hash,,}" == "${expected_hash,,}" ]]; then
            echo '{"status": "passed", "details": "evidence OK"}' > "$output_file"
            return 0
        else
            echo "{\"status\": \"failed\", \"details\": \"hash mismatch: expected=$expected_hash, found=$found_hash\"}" > "$output_file"
            return 1  # High: hash mismatch is retryable
        fi
    else
        echo '{"status": "critical", "details": "AGENTS_SUMMARY evidence not found (immediate failure)"}' > "$output_file"
        return 2  # Critical: missing evidence is an immediate failure
    fi
}

# Gate 2: Structure check
gate_structure() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 2: structure check"

    local lint_result=0
    local type_result=0
    local details=""

    # Skip if package.json does not exist
    if [[ ! -f "$worktree/package.json" ]]; then
        echo '{"status": "passed", "details": "no package.json (skipped)"}' > "$output_file"
        return 0
    fi

    # Quality: auto-detect package manager
    local pm
    pm=$(detect_package_manager "$worktree")
    local pm_run
    pm_run=$(get_pm_run_command "$pm")
    log_info "Package manager detected: $pm"

    # Accurately check scripts keys with jq
    # lint check
    if jq -e '.scripts.lint' "$worktree/package.json" > /dev/null 2>&1; then
        if ! (cd "$worktree" && $pm_run lint --silent 2>&1); then
            lint_result=1
            details="lint error"
        fi
    fi

    # type-check
    if jq -e '.scripts["type-check"]' "$worktree/package.json" > /dev/null 2>&1; then
        if ! (cd "$worktree" && $pm_run type-check --silent 2>&1); then
            type_result=1
            details="${details:+$details, }type error"
        fi
    fi

    if [[ $lint_result -eq 0 ]] && [[ $type_result -eq 0 ]]; then
        echo '{"status": "passed", "details": "structure check OK"}' > "$output_file"
        return 0
    else
        echo "{\"status\": \"failed\", \"details\": \"$details\"}" > "$output_file"
        return 1
    fi
}

# Gate 3: Tests
gate_test() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 3: test"

    # Tampering detection patterns (for added lines)
    # Critical: clear tampering — test disabling patterns
    local tamper_critical_patterns=(
        # JS/TS: skip patterns
        'it\.skip\s*\('
        'test\.skip\s*\('
        'describe\.skip\s*\('
        'xit\s*\('
        'xdescribe\s*\('
        # .only patterns (prevent other tests from running)
        '(it|test|describe)\.only\s*\('
        'fit\s*\('
        'fdescribe\s*\('
        # Python: skip patterns
        '@pytest\.mark\.skip'
        '@unittest\.skip'
        'self\.skipTest\s*\('
    )
    # Warning: may have legitimate use but worth flagging
    local tamper_warn_patterns=(
        'eslint-disable'
        '@ts-ignore'
        '@ts-expect-error'
        '@ts-nocheck'
    )

    # Patterns for removed lines (assertion deletion detection)
    local tamper_remove_patterns=(
        'expect\s*\('
        'assert\s*\('
        '\.should\s*\('
        '\.to\.\w+'
        'self\.assert'
    )

    # Diff-based tampering detection (all changes from default branch)
    local default_branch
    default_branch=$(get_default_branch)
    local merge_base
    merge_base=$(cd "$worktree" && git merge-base HEAD "origin/$default_branch" 2>/dev/null || git rev-parse HEAD~1 2>/dev/null || echo "")

    if [[ -n "$merge_base" ]]; then
        # Detect added lines (lines starting with +, excluding +++ headers)
        # Also include Python test files (test_*.py, *_test.py)
        local added_lines
        added_lines=$(cd "$worktree" && git diff "$merge_base"..HEAD --unified=0 -- '*.ts' '*.tsx' '*.js' '*.jsx' '*.spec.*' '*.test.*' '*.py' 2>/dev/null | grep '^+' | grep -v '^+++ ' || echo "")

        # Detect critical patterns (skip-type — clear tampering)
        for pattern in "${tamper_critical_patterns[@]}"; do
            if echo "$added_lines" | grep -qE "$pattern"; then
                echo "{\"status\": \"critical\", \"details\": \"tampering detected: added lines match '$pattern' pattern\"}" > "$output_file"
                return 2  # Critical: tampering detected
            fi
        done

        # Detect warning patterns (eslint-disable — may be legitimate)
        for pattern in "${tamper_warn_patterns[@]}"; do
            if echo "$added_lines" | grep -qE "$pattern"; then
                log_warn "Review needed: added lines match '$pattern' pattern (possible tampering)"
                # Record as warning, not critical
            fi
        done

        # Detect removed lines (lines starting with -) — test files only
        local removed_lines
        removed_lines=$(cd "$worktree" && git diff "$merge_base"..HEAD --unified=0 -- '*.spec.*' '*.test.*' 'test_*.py' '*_test.py' 2>/dev/null | grep '^-' | grep -v '^---' || echo "")

        for pattern in "${tamper_remove_patterns[@]}"; do
            local removed_count
            removed_count=$(echo "$removed_lines" | grep -cE "$pattern" 2>/dev/null || echo 0)

            if [[ "$removed_count" -gt 2 ]]; then
                echo "{\"status\": \"critical\", \"details\": \"tampering detected: '$pattern' removed $removed_count times from tests\"}" > "$output_file"
                return 2  # Critical: mass assertion deletion
            fi
        done

        # Catch-all assertion detection (meaningless assertions that always pass)
        # e.g. expect(true).toBe(true), expect(1).toBe(1)
        if echo "$added_lines" | grep -qE 'expect\((true|false|1|0|null|undefined)\)\.(toBe|toEqual|toStrictEqual)\((true|false|1|0|null|undefined)\)'; then
            log_warn "Review needed: catch-all assertion detected (e.g. expect(true).toBe(true))"
        fi
        # Weak assertions against constants: expect(false).toBeFalsy() etc.
        if echo "$added_lines" | grep -qE 'expect\((true|false|null|undefined|0)\)\.(toBeUndefined|toBeNull|toBeFalsy|toBeTruthy)\(\)'; then
            log_warn "Review needed: weak assertion against constant detected"
        fi

        # Detect large timeout increases (>=30000ms)
        local timeout_hit
        timeout_hit=$(echo "$added_lines" | grep -E 'jest\.setTimeout\(|jasmine\.DEFAULT_TIMEOUT_INTERVAL|[[:space:]]timeout[[:space:]]*:' | grep -oE '[0-9]+' | awk '$1 >= 30000 {found=1} END {print found+0}' 2>/dev/null || echo 0)
        if [[ "${timeout_hit:-0}" -gt 0 ]]; then
            log_warn "Review needed: large timeout increase detected (>=30000ms)"
        fi

        # Detect configuration relaxation (lint/CI/TypeScript strict)
        local config_diff
        config_diff=$(cd "$worktree" && git diff "$merge_base"..HEAD --unified=0 -- '.eslintrc*' 'eslint.config.*' 'tsconfig.json' 'tsconfig.*.json' 'biome.json' 'jest.config.*' 'vitest.config.*' '.github/workflows/*.yml' '.github/workflows/*.yaml' 2>/dev/null | grep '^+' | grep -v '^+++ ' || echo "")
        if [[ -n "$config_diff" ]]; then
            # lint rule disabling
            if echo "$config_diff" | grep -qE '"off"|:[[:space:]]*0'; then
                log_warn "Review needed: lint rule disabling detected in config file"
            fi
            # CI continue-on-error
            if echo "$config_diff" | grep -qE 'continue-on-error:[[:space:]]*true'; then
                log_warn "Review needed: CI continue-on-error addition detected"
            fi
            # TypeScript strict mode relaxation
            if echo "$config_diff" | grep -qE '"strict"[[:space:]]*:[[:space:]]*false|"noImplicitAny"[[:space:]]*:[[:space:]]*false'; then
                echo '{"status": "critical", "details": "tampering detected: TypeScript strict mode relaxed"}' > "$output_file"
                return 2
            fi
        fi
    fi

    # Run tests (auto-detect package manager)
    if [[ -f "$worktree/package.json" ]]; then
        if jq -e '.scripts.test' "$worktree/package.json" > /dev/null 2>&1; then
            local pm
            pm=$(detect_package_manager "$worktree")
            local pm_run
            pm_run=$(get_pm_run_command "$pm")

            if ! (cd "$worktree" && $pm_run test 2>&1); then
                echo '{"status": "failed", "details": "tests failed"}' > "$output_file"
                return 1
            fi
        fi
    fi

    echo '{"status": "passed", "details": "tests OK"}' > "$output_file"
    return 0
}

# Gate 4: Hardening parity
gate_hardening() {
    local worktree="$1"
    local output_file="$2"

    log_gate "Gate 4: hardening parity"

    local codex_state_dir="$worktree/.claude/state/codex-worker"
    local base_instructions_file="$codex_state_dir/base-instructions.txt"
    local prompt_file="$codex_state_dir/prompt.txt"
    local contract_file="$codex_state_dir/hardening-contract.txt"
    local marker="HARNESS_HARDENING_CONTRACT_V1"

    local status="passed"
    local violations=()

    # 1. Verify injected contract artifacts
    local contract_files=(
        "$base_instructions_file"
        "$prompt_file"
        "$contract_file"
    )

    for file in "${contract_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            violations+=("Hardening contract artifact missing: $file")
            status="critical"
            continue
        fi

        if ! grep -Fq "$marker" "$file" 2>/dev/null; then
            violations+=("Hardening contract marker missing: $file")
            status="critical"
        fi
    done

    # 2. Diff-based hardening violation check
    local diff_base
    diff_base=$(resolve_diff_base "$worktree")

    if [[ -n "$diff_base" ]]; then
        local changed_files
        local added_lines
        changed_files=$(collect_changed_files "$worktree" "$diff_base")
        added_lines=$(collect_added_lines "$worktree" "$diff_base")

        # 2-1. bypass flags check
        if echo "$added_lines" | grep -qE -- '--no-verify|--no-gpg-sign'; then
            violations+=("Added lines contain bypass flags: --no-verify or --no-gpg-sign")
            [[ "$status" == "passed" ]] && status="failed"
        fi

        # 2-2. protected branch reset check
        if echo "$added_lines" | grep -qE 'git[[:space:]]+reset[[:space:]]+--hard'; then
            if echo "$added_lines" | grep -qE '(origin/)?(main|master)'; then
                violations+=("Added lines contain protected hard reset command against main/master")
                [[ "$status" == "passed" ]] && status="failed"
            fi
        fi

        # 2-3. protected files check
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            case "$file" in
                package.json|Dockerfile|docker-compose.yml|schema.prisma|wrangler.toml|index.html|.github/workflows/*.yml|.github/workflows/*.yaml)
                    violations+=("Protected file changed: $file")
                    [[ "$status" == "passed" ]] && status="failed"
                    ;;
            esac
        done <<< "$changed_files"

        # 2-4. secrets / credentials check
        if echo "$added_lines" | grep -qE '(api[_-]?key|secret[_-]?key|auth[_-]?token|access[_-]?token|password|credential|private[_-]?key)[[:space:]]*[:=]'; then
            violations+=("Added lines contain hardcoded secret-like assignment")
            [[ "$status" == "passed" ]] && status="failed"
        fi

        if echo "$added_lines" | grep -qE '(postgres://|mysql://|mongodb://|redis://|amqp://|DATABASE_URL|REDIS_URL)'; then
            violations+=("Added lines contain hardcoded service/database connection string")
            [[ "$status" == "passed" ]] && status="failed"
        fi

        if echo "$added_lines" | grep -qE '(192\.168\.[0-9]+\.[0-9]+|10\.[0-9]+\.[0-9]+\.[0-9]+|172\.(1[6-9]|2[0-9]|3[01])\.[0-9]+\.[0-9]+)'; then
            violations+=("Added lines contain private IP address")
            [[ "$status" == "passed" ]] && status="failed"
        fi
    else
        violations+=("Hardening diff base could not be resolved")
        status="critical"
    fi

    local violations_json='[]'
    if [[ ${#violations[@]} -gt 0 ]]; then
        violations_json=$(printf '%s\n' "${violations[@]}" | jq -R . | jq -s '.')
    fi

    local details="hardening parity OK"
    case "$status" in
        critical) details="hardening contract missing or invalid" ;;
        failed) details="hardening violations detected" ;;
    esac

    jq -n \
        --arg status "$status" \
        --arg details "$details" \
        --arg marker "$marker" \
        --argjson violations "$violations_json" \
        '{
            status: $status,
            details: $details,
            marker: $marker,
            violations: $violations
        }' > "$output_file"

    case "$status" in
        passed) return 0 ;;
        failed) return 1 ;;
        critical) return 2 ;;
    esac
}

# Main processing
main() {
    check_dependencies

    local worktree=""
    local skip_gates=()
    local skip_reason=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --worktree)
                if [[ -z "${2:-}" ]]; then
                    log_error "--worktree requires a value"
                    exit 1
                fi
                worktree="$2"; shift 2 ;;
            --skip-gate)
                if [[ -z "${2:-}" ]]; then
                    log_error "--skip-gate requires a value"
                    exit 1
                fi
                skip_gates+=("$2"); shift 2 ;;
            --reason)
                if [[ -z "${2:-}" ]]; then
                    log_error "--reason requires a value"
                    exit 1
                fi
                skip_reason="$2"; shift 2 ;;
            -h|--help) usage; exit 0 ;;
            *) log_error "Unknown option: $1"; exit 1 ;;
        esac
    done

    # Check required parameters
    if [[ -z "$worktree" ]]; then
        log_error "--worktree is required"
        exit 1
    fi

    if [[ ! -d "$worktree" ]]; then
        log_error "Worktree does not exist: $worktree"
        exit 1
    fi

    # Security: verify it is a worktree of the same repository
    if ! validate_worktree_path "$worktree"; then
        exit 1
    fi

    # Reason is required when skipping
    if [[ ${#skip_gates[@]} -gt 0 ]] && [[ -z "$skip_reason" ]]; then
        log_error "--reason is required when using --skip-gate"
        exit 1
    fi

    # Security: check skip allowlist (default: deny)
    if [[ ${#skip_gates[@]} -gt 0 ]]; then
        local allowlist
        allowlist=$(get_config "gate_skip_allowlist")

        # If allowlist is empty or [], deny all skips
        if [[ -z "$allowlist" || "$allowlist" == "[]" || "$allowlist" == "null" ]]; then
            log_error "Gate skipping is not permitted (allowlist is empty)"
            log_error "To allow skipping, edit gate_skip_allowlist in the configuration file"
            exit 1
        fi

        for gate in "${skip_gates[@]}"; do
            # Check if gate is in the allowlist
            if ! echo "$allowlist" | jq -e --arg g "$gate" 'index($g) != null' >/dev/null 2>&1; then
                log_error "Gate '$gate' is not in the allowlist"
                log_error "Allowed gates: $allowlist"
                exit 1
            fi
            log_warn "Skipping gate '$gate': $skip_reason (user=${USER:-unknown})"
        done
    fi

    # Temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d /tmp/harness-tmp.XXXXXX)
    trap 'rm -rf "$tmp_dir"' EXIT

    # Store results
    local overall_status="passed"
    local gates_json="{}"
    local skipped_json="[]"
    local errors_json="[]"

    # Gate 1: evidence verification
    if [[ " ${skip_gates[*]} " =~ " evidence " ]]; then
        log_warn "Skipping Gate 1 (evidence)"
        log_skip "evidence" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["evidence"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.evidence = {"status": "skipped", "details": $reason}')
    else
        local evidence_exit_code=0
        gate_evidence "$worktree" "$tmp_dir/evidence.json" || evidence_exit_code=$?

        if [[ $evidence_exit_code -eq 0 ]]; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
        elif [[ $evidence_exit_code -eq 2 ]]; then
            # Critical: missing evidence
            overall_status="critical"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["CRITICAL: evidence missing"]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/evidence.json" '.evidence = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 1 failed: hash mismatch"]')
        fi
    fi

    # Gate 2: structure check
    if [[ " ${skip_gates[*]} " =~ " structure " ]]; then
        log_warn "Skipping Gate 2 (structure)"
        log_skip "structure" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["structure"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.structure = {"status": "skipped", "details": $reason}')
    else
        if gate_structure "$worktree" "$tmp_dir/structure.json"; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/structure.json" '.structure = $g[0]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/structure.json" '.structure = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 2 failed"]')
        fi
    fi

    # Gate 3: tests
    if [[ " ${skip_gates[*]} " =~ " test " ]]; then
        log_warn "Skipping Gate 3 (test)"
        log_skip "test" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["test"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.test = {"status": "skipped", "details": $reason}')
    else
        local test_exit_code=0
        gate_test "$worktree" "$tmp_dir/test.json" || test_exit_code=$?

        if [[ $test_exit_code -eq 0 ]]; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
        elif [[ $test_exit_code -eq 2 ]]; then
            # Critical: tampering detected
            overall_status="critical"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["CRITICAL: tampering detected"]')
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/test.json" '.test = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 3 failed"]')
        fi
    fi

    # Gate 4: Hardening parity
    if [[ " ${skip_gates[*]} " =~ " hardening " ]]; then
        log_warn "Skipping Gate 4 (hardening)"
        log_skip "hardening" "$skip_reason"
        skipped_json=$(echo "$skipped_json" | jq '. + ["hardening"]')
        gates_json=$(echo "$gates_json" | jq --arg reason "$skip_reason" '.hardening = {"status": "skipped", "details": $reason}')
    else
        local hardening_exit_code=0
        gate_hardening "$worktree" "$tmp_dir/hardening.json" || hardening_exit_code=$?

        if [[ $hardening_exit_code -eq 0 ]]; then
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/hardening.json" '.hardening = $g[0]')
        elif [[ $hardening_exit_code -eq 2 ]]; then
            overall_status="critical"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/hardening.json" '.hardening = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["CRITICAL: hardening contract missing or invalid"]')
            while IFS= read -r violation; do
                [[ -z "$violation" ]] && continue
                errors_json=$(echo "$errors_json" | jq --arg v "$violation" '. + [$v]')
            done < <(jq -r '.violations[]' "$tmp_dir/hardening.json" 2>/dev/null || true)
        else
            overall_status="failed"
            gates_json=$(echo "$gates_json" | jq --slurpfile g "$tmp_dir/hardening.json" '.hardening = $g[0]')
            errors_json=$(echo "$errors_json" | jq '. + ["Gate 4 failed: hardening violations"]')
            while IFS= read -r violation; do
                [[ -z "$violation" ]] && continue
                errors_json=$(echo "$errors_json" | jq --arg v "$violation" '. + [$v]')
            done < <(jq -r '.violations[]' "$tmp_dir/hardening.json" 2>/dev/null || true)
        fi
    fi

    # Output final result
    local result
    result=$(jq -n \
        --arg status "$overall_status" \
        --argjson gates "$gates_json" \
        --argjson skipped "$skipped_json" \
        --argjson errors "$errors_json" \
        '{
            status: $status,
            gates: $gates,
            skipped: $skipped,
            errors: $errors
        }')

    # Security: centrally manage gate result (saved outside worktree)
    local details_summary
    details_summary=$(echo "$result" | jq -c '.errors')
    save_gate_result "$worktree" "$overall_status" "$details_summary"

    echo "$result"

    # Exit code
    case "$overall_status" in
        passed) exit 0 ;;
        failed) exit 1 ;;
        critical) exit 2 ;;
    esac
}

# Usage
usage() {
    cat << EOF
Usage: $0 --worktree PATH [OPTIONS]

Options:
  --worktree PATH       Worktree to inspect (required)
  --skip-gate GATE      Skip a specific gate (evidence, structure, test, hardening)
  --reason TEXT         Skip reason (required with --skip-gate)
  -h, --help            Show help

Gates:
  evidence   - AGENTS_SUMMARY evidence verification
  structure  - lint, type-check
  test       - test execution, tampering detection
  hardening  - Codex parity hardening (contract, bypass flags, protected files, secrets)

Examples:
  $0 --worktree ../worktrees/worker-1
  $0 --worktree ../worktrees/worker-1 --skip-gate test --reason "test environment not configured"
EOF
}

main "$@"
