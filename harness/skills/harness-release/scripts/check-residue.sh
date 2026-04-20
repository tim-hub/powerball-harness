#!/usr/bin/env bash
# check-residue.sh — thin launcher for check-residue.py
# Kept for backward compatibility with existing references in validate-plugin.sh and SKILL.md.
exec python3 "$(dirname "${BASH_SOURCE[0]}")/check-residue.py" "$@"
