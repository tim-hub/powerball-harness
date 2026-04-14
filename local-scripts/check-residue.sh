#!/usr/bin/env bash
# check-residue.sh — Migration Residue Scanner (Phase 40)
#
# Purpose:
#   Load .claude/rules/deleted-concepts.yaml and detect whether deleted paths or
#   concepts still exist anywhere in the repository.
#   Exits 0 if none found, exits 1 if one or more are found.
#
# Usage:
#   bash local-scripts/check-residue.sh
#
# Python3 acts as the primary parser; bash is used only as a launcher.

set -euo pipefail

# Determine the repository root (resolved relative to the script location)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

export REPO_ROOT_PY="${REPO_ROOT}"

exec python3 - "$@" <<'PYEOF'
import subprocess
import sys
import os
import time
import re

try:
    import yaml
    def _load_yaml(path):
        with open(path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)
except ImportError:
    # PyYAML not installed — fall back to an indent-aware minimal parser.
    # Handles the known schema: version, deleted_paths[], deleted_concepts[]
    # each with path/term, deleted_in, reason, allowlist[], optional fields.
    def _load_yaml(path):
        """Indent-aware YAML loader for deleted-concepts.yaml (stdlib only)."""
        with open(path, "r", encoding="utf-8") as f:
            lines = f.readlines()

        result = {"version": 1, "deleted_paths": [], "deleted_concepts": []}
        current_section = None   # 'deleted_paths' or 'deleted_concepts'
        current_entry = None     # dict being built
        in_allowlist = False     # True when inside an allowlist: block
        allowlist_indent = -1    # indent level of allowlist items

        def _strip_value(s):
            """Strip inline comments, then quotes from a YAML scalar value."""
            # Remove inline comment (space + #) — but only outside quoted strings
            # Simple heuristic: if starts with " or ', strip until matching close quote
            s = s.strip()
            if s.startswith('"'):
                end = s.find('"', 1)
                return s[1:end] if end != -1 else s[1:]
            if s.startswith("'"):
                end = s.find("'", 1)
                return s[1:end] if end != -1 else s[1:]
            # No quotes: strip trailing inline comment
            if '  #' in s:
                s = s[:s.index('  #')]
            elif s.startswith('#'):
                return ''
            return s.strip()

        def flush_entry():
            if current_entry is not None and current_section is not None:
                result[current_section].append(current_entry)

        for raw in lines:
            if not raw.strip() or raw.strip().startswith('#'):
                continue
            indent = len(raw) - len(raw.lstrip())
            stripped = raw.strip()

            # Top-level keys (indent == 0)
            if indent == 0:
                if stripped.startswith('version:'):
                    m = re.match(r'version:\s*(\d+)', stripped)
                    if m:
                        result['version'] = int(m.group(1))
                elif stripped.startswith('deleted_paths:'):
                    flush_entry()
                    current_section = 'deleted_paths'
                    current_entry = None
                    in_allowlist = False
                elif stripped.startswith('deleted_concepts:'):
                    flush_entry()
                    current_section = 'deleted_concepts'
                    current_entry = None
                    in_allowlist = False
                continue

            if current_section is None:
                continue

            # New list entry: starts with '- ' at indent 2
            if stripped.startswith('- ') and indent <= 4:
                flush_entry()
                current_entry = {'allowlist': []}
                in_allowlist = False
                rest = stripped[2:]
                if ':' in rest:
                    k, v = rest.split(':', 1)
                    val = _strip_value(v)
                    current_entry[k.strip()] = val if val else None
                continue

            if current_entry is None:
                continue

            # Inside an entry
            if in_allowlist and indent > allowlist_indent - 4:
                # This is an allowlist item
                if stripped.startswith('- '):
                    item = _strip_value(stripped[2:])
                    if item:
                        current_entry['allowlist'].append(item)
                    continue
                else:
                    in_allowlist = False

            if stripped.startswith('allowlist:'):
                in_allowlist = True
                allowlist_indent = indent + 2
                continue

            if ':' in stripped and not stripped.startswith('-'):
                k, v = stripped.split(':', 1)
                k = k.strip()
                v = _strip_value(v)
                if k == '_scan_disabled':
                    current_entry[k] = v.lower() in ('true', '1', 'yes')
                elif v:
                    current_entry[k] = v
                else:
                    current_entry[k] = None

        flush_entry()
        return result

REPO_ROOT = os.path.normpath(
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
) if False else os.environ.get("REPO_ROOT_PY", "")

# When called via bash exec python3 -, __file__ is not available.
# Instead of passing through environment variables, infer from argv[0].
# However, in a heredoc exec case sys.argv[0] == '-', so resolve from getcwd().
if not REPO_ROOT:
    # The script is expected to be called from local-scripts/. cwd is arbitrary,
    # so use sys.argv if provided, otherwise resolve from cwd.
    REPO_ROOT = os.getcwd()
    # If cwd is local-scripts/, go up to the repo root.
    if os.path.basename(REPO_ROOT) in ("scripts", "local-scripts"):
        REPO_ROOT = os.path.dirname(REPO_ROOT)

YAML_PATH = os.path.join(REPO_ROOT, ".claude/rules/deleted-concepts.yaml")

start_time = time.time()

# ─── Load YAML ─────────────────────────────────────────────────────────────
if not os.path.exists(YAML_PATH):
    print(f"ERROR: {YAML_PATH} not found", file=sys.stderr)
    sys.exit(2)

config = _load_yaml(YAML_PATH)

deleted_paths    = config.get("deleted_paths", [])
deleted_concepts = config.get("deleted_concepts", [])

# Skip entries that have the scan_disabled flag set
deleted_concepts = [c for c in deleted_concepts if not c.get("_scan_disabled", False)]

n_paths    = len(deleted_paths)
n_concepts = len(deleted_concepts)

print("=== Migration Residue Scan ===")
print(f"Loaded: .claude/rules/deleted-concepts.yaml")
print(f"Entries: {n_paths} deleted_paths + {n_concepts} deleted_concepts")
print()

# ─── Allowlist check ─────────────────────────────────────────────────────────
def is_allowlisted(filepath: str, allowlist: list) -> bool:
    """
    Determine whether filepath matches any prefix in the allowlist.
    Allowlist entries are prefix-matched.
    filepath is relative to the repository root (no leading ./).
    """
    # Strip leading ./ and normalize
    rel = filepath.lstrip("./")
    for entry in allowlist:
        entry_clean = entry.lstrip("./")
        if rel.startswith(entry_clean):
            return True
    return False

# ─── grep utility ─────────────────────────────────────────────────────────────
def grep_files(term: str, repo_root: str) -> list:
    """
    Search repo_root recursively for term as a fixed string using grep -rln -F.
    Returns a list of relative paths of matched files.
    """
    try:
        result = subprocess.run(
            ["grep", "-rln", "-F",
             "--exclude-dir=.git",
             term, "."],
            cwd=repo_root,
            capture_output=True,
            text=True,
        )
        if result.returncode not in (0, 1):
            # Ignore grep errors (returncode=2)
            return []
        files = [f.strip() for f in result.stdout.splitlines() if f.strip()]
        return files
    except Exception as e:
        print(f"  WARNING: grep execution error: {e}", file=sys.stderr)
        return []

def grep_line_numbers(term: str, filepath: str, repo_root: str) -> list:
    """
    Return the line numbers and content of lines in filepath that match term.
    Returns: list of (lineno, line_content)
    """
    try:
        result = subprocess.run(
            ["grep", "-n", "-F", term, filepath],
            cwd=repo_root,
            capture_output=True,
            text=True,
        )
        lines = []
        for line in result.stdout.splitlines():
            # Format: "27:    grep 'core/src/guardrails/rules.ts'"
            m = re.match(r"^(\d+):(.*)$", line)
            if m:
                lines.append((int(m.group(1)), m.group(2).strip()))
        return lines
    except Exception:
        return []

def grep_h1_v3_files(repo_root: str) -> list:
    """
    Search SKILL.md / agents/*.md files whose H1 title has a '(v3)' suffix.
    Pattern: Lines starting with '# ' and containing '(v3)'.
    Use grep -rl because grep -rln is not available here.
    """
    try:
        result = subprocess.run(
            ["grep", "-rln", "--include=*.md",
             "--exclude-dir=.git",
             r"^# .*(v3)", "."],
            cwd=repo_root,
            capture_output=True,
            text=True,
        )
        if result.returncode not in (0, 1):
            return []
        files = [f.strip() for f in result.stdout.splitlines() if f.strip()]
        return files
    except Exception:
        return []

# ─── Run scan ─────────────────────────────────────────────────────────────────
violations = 0
violation_files = set()

# ── Scan deleted_paths ──
print("[scanning deleted_paths...]")
for entry in deleted_paths:
    path_term = entry["path"]
    allowlist  = entry.get("allowlist", [])
    reason     = entry.get("reason", "")

    # Add default allowlist entries (common to all entries)
    default_allowlist = [
        "CHANGELOG.md",
        ".claude/memory/archive/",
        ".claude/worktrees/",
        ".claude/state/",
        "out/",
        "output/",          # Diagnostic output and generated artifacts
        "benchmarks/",
        "tests/validate-plugin-v3.sh",  # v3 compatibility test (intentionally kept)
        ".claude/rules/deleted-concepts.yaml",  # This file itself
        "local-scripts/check-residue.sh",             # The scanner itself
    ]
    effective_allowlist = list(set(allowlist + default_allowlist))

    matched_files = grep_files(path_term, REPO_ROOT)

    # Filter by allowlist
    filtered = [f for f in matched_files if not is_allowlisted(f, effective_allowlist)]

    if filtered:
        violations += len(filtered)
        violation_files.update(filtered)
        print(f"  ✗ {path_term}")
        for f in filtered:
            lines = grep_line_numbers(path_term, f, REPO_ROOT)
            if lines:
                for lineno, content in lines[:3]:  # Show up to 3 lines
                    print(f"    {f}:L{lineno} — \"{content}\"")
            else:
                print(f"    {f}")
        print(f"    (matched entry: {path_term}, reason: \"{reason[:60]}...\")" if len(reason) > 60 else f"    (matched entry: {path_term}, reason: \"{reason}\")")
        print()

# ── Scan deleted_concepts ──
print("[scanning deleted_concepts...]")
for entry in deleted_concepts:
    if entry.get("_scan_disabled", False):
        continue

    term       = entry["term"]
    term_ja    = entry.get("term_ja")
    replacement = entry.get("replacement", "")
    reason     = entry.get("reason", "")
    allowlist  = entry.get("allowlist", [])

    default_allowlist = [
        "CHANGELOG.md",
        ".claude/memory/archive/",
        ".claude/worktrees/",
        ".claude/state/",
        "out/",
        "output/",          # Diagnostic output and generated artifacts
        "benchmarks/",
        ".claude/rules/deleted-concepts.yaml",  # Exclude this file itself
        "local-scripts/check-residue.sh",             # Exclude the scanner itself
        "tests/validate-plugin-v3.sh",          # v3 compatibility test (intentionally kept)
    ]
    effective_allowlist = list(set(allowlist + default_allowlist))

    # Scan for the English term
    terms_to_scan = [term]
    if term_ja:
        terms_to_scan.append(term_ja)

    for scan_term in terms_to_scan:
        matched_files = grep_files(scan_term, REPO_ROOT)
        filtered = [f for f in matched_files if not is_allowlisted(f, effective_allowlist)]

        if filtered:
            violations += len(filtered)
            violation_files.update(filtered)
            display_term = scan_term
            display_replacement = f" → {replacement}" if replacement else ""
            print(f"  ✗ \"{display_term}\"")
            for f in filtered:
                lines = grep_line_numbers(scan_term, f, REPO_ROOT)
                if lines:
                    for lineno, content in lines[:3]:
                        print(f"    {f}:L{lineno} — \"{content}\"")
                else:
                    print(f"    {f}")
            print(f"    (matched entry: {display_term}{display_replacement})")
            print()

# ── Scan for H1 (v3) suffix (special handling) ──
print("[scanning H1 (v3) suffix in skills/ and agents/...]")
h1_allowlist = [
    "CHANGELOG.md",
    ".claude/memory/archive/",
    ".claude/worktrees/",
    ".claude/state/",
    "out/",
    "output/",
    "benchmarks/",
    ".claude/rules/",  # Historical documents inside rules/
    "local-scripts/check-residue.sh",
    ".claude/rules/deleted-concepts.yaml",
    "tests/validate-plugin-v3.sh",  # v3 compatibility test (intentionally kept)
]

h1_files = grep_h1_v3_files(REPO_ROOT)
h1_filtered = [f for f in h1_files if not is_allowlisted(f, h1_allowlist)]

if h1_filtered:
    violations += len(h1_filtered)
    violation_files.update(h1_filtered)
    print(f"  ✗ H1 title with (v3) suffix")
    for f in h1_filtered:
        # Show matching lines
        try:
            result = subprocess.run(
                ["grep", "-n", r"^# .*(v3)", f],
                cwd=REPO_ROOT,
                capture_output=True,
                text=True,
            )
            for line in result.stdout.splitlines()[:3]:
                m = re.match(r"^(\d+):(.*)$", line)
                if m:
                    print(f"    {f}:L{m.group(1)} — \"{m.group(2).strip()}\"")
        except Exception:
            print(f"    {f}")
    print("    (matched entry: H1 (v3) suffix → remove version suffix from H1 titles)")
    print()

# ─── Summary output ───────────────────────────────────────────────────────────
elapsed = time.time() - start_time

print("=== Summary ===")
if violations == 0:
    print("  ✓ No migration residue detected")
    print(f"  Scan duration: {elapsed:.1f}s")
    print("  Exit: 0")
    sys.exit(0)
else:
    print(f"  Violations: {violations} (in {len(violation_files)} files)")
    print(f"  Scan duration: {elapsed:.1f}s")
    print("  Exit: 1 (residue detected)")
    sys.exit(1)

PYEOF
