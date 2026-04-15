#!/bin/bash
# review-ai-residuals.sh
# Statically detect AI implementation residue candidates from a diff or target files.
#
# Usage:
#   bash scripts/review-ai-residuals.sh --base-ref <git-ref>
#   bash scripts/review-ai-residuals.sh path/to/file.ts path/to/config.sh
#
# Exit:
#   0: Normal exit regardless of whether issues are detected (the review side determines the verdict)
#   2: Usage error

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash scripts/review-ai-residuals.sh --base-ref <git-ref>
  bash scripts/review-ai-residuals.sh <file> [<file> ...]

Options:
  --base-ref <git-ref>  Automatically collect changed files via git diff
  --help                Show this help message

Output:
  Stable JSON:
  {
    "tool": "review-ai-residuals",
    "scan_mode": "diff|files",
    "base_ref": "HEAD~1" | null,
    "files_scanned": ["src/app.ts"],
    "summary": {
      "verdict": "APPROVE|REQUEST_CHANGES",
      "major": 0,
      "minor": 0,
      "recommendation": 0,
      "total": 0
    },
    "observations": []
  }
EOF
}

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\t'/\\t}"
  value="${value//$'\r'/\\r}"
  value="${value//$'\n'/\\n}"
  printf '%s' "$value"
}

trim_match_text() {
  local value="$1"
  value="$(printf '%s' "$value" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"
  if [ "${#value}" -gt 180 ]; then
    printf '%s...' "${value:0:177}"
  else
    printf '%s' "$value"
  fi
}

redact_secret_line() {
  printf '%s' "$1" | sed -E \
    "s/((api[_-]?key|secret|token|password|passwd|client[_-]?secret)[^:=]{0,20}[:=][[:space:]]*['\"]).+(['\"])/\1<redacted>\3/I"
}

should_ignore_path() {
  case "$1" in
    *.md|*.mdx|*.txt|*.rst|*.adoc) return 0 ;;
    docs/*|*/docs/*) return 0 ;;
    examples/*|*/examples/*) return 0 ;;
    tests/fixtures/*|*/tests/fixtures/*) return 0 ;;
    */node_modules/*|node_modules/*) return 0 ;;
    .git/*|*/.git/*) return 0 ;;
  esac
  return 1
}

is_scannable_file() {
  case "$1" in
    *.sh|*.bash|*.zsh|*.js|*.jsx|*.mjs|*.cjs|*.ts|*.tsx|*.py|*.rb|*.php|*.go|*.rs|*.java|*.kt|*.kts|*.swift|*.json|*.yml|*.yaml|*.toml|*.ini|*.cfg|*.conf|*.env)
      return 0
      ;;
  esac
  return 1
}

append_json_string_array() {
  local file="$1"
  local first=1
  printf '['
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    if [ "$first" -eq 0 ]; then
      printf ','
    fi
    printf '"%s"' "$(json_escape "$line")"
    first=0
  done < "$file"
  printf ']'
}

append_json_object_array() {
  local file="$1"
  local first=1
  printf '['
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    if [ "$first" -eq 0 ]; then
      printf ','
    fi
    printf '%s' "$line"
    first=0
  done < "$file"
  printf ']'
}

SEARCH_TOOL=""
if command -v rg >/dev/null 2>&1; then
  SEARCH_TOOL="rg"
else
  echo '{"tool":"review-ai-residuals","scan_mode":"files","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[],"warning":"rg_not_found"}'
  exit 0
fi

SCAN_MODE="files"
BASE_REF_INPUT=""
POSITIONAL_FILES=()

while [ $# -gt 0 ]; do
  case "$1" in
    --base-ref)
      if [ $# -lt 2 ]; then
        echo "error: --base-ref requires a value" >&2
        usage >&2
        exit 2
      fi
      SCAN_MODE="diff"
      BASE_REF_INPUT="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      echo "error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      POSITIONAL_FILES+=("$1")
      shift
      ;;
  esac
done

if [ "$SCAN_MODE" = "diff" ] && [ ${#POSITIONAL_FILES[@]} -gt 0 ]; then
  echo "error: --base-ref and explicit files cannot be combined" >&2
  usage >&2
  exit 2
fi

if [ "$SCAN_MODE" = "files" ] && [ ${#POSITIONAL_FILES[@]} -eq 0 ]; then
  if [ -n "${BASE_REF:-}" ]; then
    SCAN_MODE="diff"
    BASE_REF_INPUT="${BASE_REF}"
  else
    SCAN_MODE="diff"
    BASE_REF_INPUT="HEAD~1"
  fi
fi

TMP_FILES="$(mktemp /tmp/harness-tmp.XXXXXX)"
TMP_OBS="$(mktemp /tmp/harness-tmp.XXXXXX)"
TMP_DIFF="$(mktemp /tmp/harness-tmp.XXXXXX)"
cleanup() {
  rm -f "$TMP_FILES" "$TMP_OBS" "$TMP_DIFF"
}
trap cleanup EXIT

collect_diff_files() {
  local base_ref="$1"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    return 1
  fi
  git diff --name-only --diff-filter=ACMR "$base_ref" -- 2>/dev/null || return 1
}

queue_file_if_scannable() {
  local path="$1"
  path="${path#./}"
  [ -f "$path" ] || return 0
  should_ignore_path "$path" && return 0
  is_scannable_file "$path" || return 0
  printf '%s\n' "$path" >> "$TMP_FILES"
}

if [ "$SCAN_MODE" = "diff" ]; then
  if collect_diff_files "$BASE_REF_INPUT" >"$TMP_DIFF" 2>/dev/null; then
    while IFS= read -r path; do
      queue_file_if_scannable "$path"
    done < "$TMP_DIFF"
  fi
else
  for path in "${POSITIONAL_FILES[@]}"; do
    queue_file_if_scannable "$path"
  done
fi

sort -u "$TMP_FILES" -o "$TMP_FILES"

MAJOR_COUNT=0
MINOR_COUNT=0
RECOMMENDATION_COUNT=0

append_observation() {
  local severity="$1"
  local rule="$2"
  local location="$3"
  local issue="$4"
  local suggestion="$5"
  local match_text="$6"

  case "$severity" in
    major) MAJOR_COUNT=$((MAJOR_COUNT + 1)) ;;
    minor) MINOR_COUNT=$((MINOR_COUNT + 1)) ;;
    recommendation) RECOMMENDATION_COUNT=$((RECOMMENDATION_COUNT + 1)) ;;
  esac

  printf '{"severity":"%s","category":"AI Residuals","rule":"%s","location":"%s","issue":"%s","suggestion":"%s","match":"%s"}\n' \
    "$(json_escape "$severity")" \
    "$(json_escape "$rule")" \
    "$(json_escape "$location")" \
    "$(json_escape "$issue")" \
    "$(json_escape "$suggestion")" \
    "$(json_escape "$match_text")" \
    >> "$TMP_OBS"
}

scan_file() {
  local file="$1"
  while IFS=$'\t' read -r rule severity pattern issue suggestion; do
    [ -n "$rule" ] || continue
    while IFS= read -r hit; do
      [ -n "$hit" ] || continue
      local line_num line_text location match_text
      line_num="${hit%%:*}"
      line_text="${hit#*:}"
      location="${file}:${line_num}"
      match_text="$(trim_match_text "$line_text")"
      if [ "$rule" = "hardcoded-secret" ]; then
        match_text="$(trim_match_text "$(redact_secret_line "$match_text")")"
      fi
      append_observation "$severity" "$rule" "$location" "$issue" "$suggestion" "$match_text"
    done < <("${SEARCH_TOOL}" --no-config -n -I --pcre2 "$pattern" -- "$file" 2>/dev/null || true)
  done <<'EOF'
test-skip	major	\b(it|describe|test)\.skip\s*\(	A disabled test remains in the code. It may allow issues to slip through review.	Remove the skip, or if absolutely necessary leave a reason in a comment and issue.
localhost-reference	major	\b(localhost|127\.0\.0\.1|0\.0\.0\.0)\b	A local-only connection target remains. This is prone to misconfiguration in production or shared environments.	Inject the URL/host from an environment variable or a public configuration source.
hardcoded-secret	major	(?i)\b(api[_-]?key|secret|token|password|passwd|client[_-]?secret)\b[^:=\n]{0,20}[:=][[:space:]]*['"][^'"]{8,}['"]	A value that looks like a secret is hardcoded. This is dangerous both for leakage and environment lock-in.	Replace it with an environment variable, a secrets store, or a safe configuration injection mechanism.
hardcoded-env-url	major	https?://(dev|staging|internal|sandbox)[.-][A-Za-z0-9._/-]+	An environment-specific URL is hardcoded. This can lead to incorrect connections in the wrong environment.	Move it to per-environment configuration.
mock-data	minor	\bmockData\b	A mock value name remains. Verify whether placeholder data has been carried into production code.	Replace with real data, or clearly mark it as test-only if intentional.
dummy-value	minor	\bdummy[A-Za-z0-9_]*\b	A placeholder named "dummy" remains.	Replace with a real value or rename to a name that conveys intent.
fake-data	minor	\bfake(Data)?\b	A name derived from fake data remains.	If in production code replace with a real implementation; if in test code, make the purpose explicit.
todo-fixme	minor	\b(TODO|FIXME)\b	An unresolved TODO / FIXME remains.	Resolve it before shipping, or leave a tracking reference in a comment.
provisional-comment	recommendation	(?i)(temporary implementation|stub implementation|placeholder implementation|replace later|hardcoded for now|wire real service)	A provisional implementation comment remains. It may not cause an immediate incident, but clarifying intent is safer.	Leave a deadline, tracking reference, and a plan for the permanent fix in a comment or issue.
EOF
}

while IFS= read -r file; do
  [ -n "$file" ] || continue
  scan_file "$file"
done < "$TMP_FILES"

TOTAL_COUNT=$((MAJOR_COUNT + MINOR_COUNT + RECOMMENDATION_COUNT))
VERDICT="APPROVE"
if [ "$MAJOR_COUNT" -gt 0 ]; then
  VERDICT="REQUEST_CHANGES"
fi

if [ -n "$BASE_REF_INPUT" ] && [ "$SCAN_MODE" = "diff" ]; then
  BASE_REF_JSON="\"$(json_escape "$BASE_REF_INPUT")\""
else
  BASE_REF_JSON="null"
fi

printf '{'
printf '"tool":"review-ai-residuals",'
printf '"scan_mode":"%s",' "$(json_escape "$SCAN_MODE")"
printf '"base_ref":%s,' "$BASE_REF_JSON"
printf '"files_scanned":%s,' "$(append_json_string_array "$TMP_FILES")"
printf '"summary":{"verdict":"%s","major":%s,"minor":%s,"recommendation":%s,"total":%s},' \
  "$VERDICT" \
  "$MAJOR_COUNT" \
  "$MINOR_COUNT" \
  "$RECOMMENDATION_COUNT" \
  "$TOTAL_COUNT"
printf '"observations":%s' "$(append_json_object_array "$TMP_OBS")"
printf '}\n'
