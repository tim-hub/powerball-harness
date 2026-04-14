#!/bin/bash
# path-utils.sh
# Cross-platform path utility functions for Windows/Mac/Linux compatibility.
#
# Usage:
#   source "${SCRIPT_DIR}/path-utils.sh"
#   detect_os
#   is_absolute_path "/some/path"
#   normalize_path "C:\Users\name"
#
# Supports:
#   - macOS (darwin)
#   - Linux
#   - Windows (Git Bash/MSYS2/Cygwin/WSL)

# ============================================================================
# OS Detection
# ============================================================================

# Cache variable for detect_os result (set on first call)
_DETECTED_OS=""

# Detect the current operating system
# Returns: darwin, linux, windows, or unknown
# Note: Result is cached for performance
detect_os() {
  # Return cached result if available
  if [ -n "$_DETECTED_OS" ]; then
    echo "$_DETECTED_OS"
    return 0
  fi

  local result
  case "${OSTYPE:-}" in
    darwin*) result="darwin" ;;
    linux*)
      # Check for WSL
      if grep -qiE "(microsoft|wsl)" /proc/version 2>/dev/null; then
        result="windows"
      else
        result="linux"
      fi
      ;;
    msys*|cygwin*|mingw*)
      result="windows"
      ;;
    *)
      # Fallback: check for Windows environment variables
      if [ -n "${WINDIR:-}" ] || [ -n "${SYSTEMROOT:-}" ]; then
        result="windows"
      elif [ "$(uname -s 2>/dev/null)" = "Darwin" ]; then
        result="darwin"
      elif [ "$(uname -s 2>/dev/null)" = "Linux" ]; then
        result="linux"
      else
        result="unknown"
      fi
      ;;
  esac

  # Cache and return the result
  _DETECTED_OS="$result"
  echo "$result"
}

# Check if running on Windows (any variant)
is_windows() {
  [ "$(detect_os)" = "windows" ]
}

# Check if running on macOS
is_macos() {
  [ "$(detect_os)" = "darwin" ]
}

# Check if running on Linux (excluding WSL)
is_linux() {
  [ "$(detect_os)" = "linux" ]
}

# ============================================================================
# Path Detection
# ============================================================================

# Check if a path is absolute (supports Unix and Windows formats)
# Examples:
#   /home/user       -> true (Unix)
#   C:/Users/name    -> true (Windows with forward slash)
#   C:\Users\name    -> true (Windows with backslash)
#   D:\              -> true (Windows drive root)
#   ./relative       -> false
#   relative/path    -> false
is_absolute_path() {
  local p="$1"
  [ -z "$p" ] && return 1

  # Unix absolute path (starts with /)
  [[ "$p" == /* ]] && return 0

  # Windows absolute path (drive letter: C:/ or C:\)
  # Also handles uppercase and lowercase drive letters
  [[ "$p" =~ ^[A-Za-z]:[\\/] ]] && return 0

  # Windows UNC path (\\server\share or //server/share)
  [[ "$p" =~ ^[\\/][\\/] ]] && return 0

  return 1
}

# Check if a path is a Windows-style path
is_windows_path() {
  local p="$1"
  [ -z "$p" ] && return 1

  # Drive letter format
  [[ "$p" =~ ^[A-Za-z]:[\\/] ]] && return 0

  # UNC path
  [[ "$p" =~ ^[\\/][\\/] ]] && return 0

  return 1
}

# ============================================================================
# Path Normalization
# ============================================================================

# Normalize a path:
#   - Convert backslashes to forward slashes
#   - Remove trailing slashes (except for root paths)
#   - Collapse multiple slashes
# Note: Does NOT resolve symlinks or relative components (use realpath for that)
normalize_path() {
  local p="$1"
  [ -z "$p" ] && echo "" && return 0

  # Convert backslashes to forward slashes using parameter expansion
  # This is much faster than character-by-character loop
  p="${p//\\//}"

  # Collapse multiple slashes to single slash using tr -s (squeeze repeats)
  # This is more reliable than parameter expansion for this operation
  if [[ "$p" == //* ]]; then
    # Preserve leading // for UNC paths
    local prefix="//"
    p="${p#//}"
    p="$(printf '%s' "$p" | tr -s '/')"
    p="${prefix}${p}"
  else
    p="$(printf '%s' "$p" | tr -s '/')"
  fi

  # Remove trailing slash (unless it's root "/" or "C:/")
  if [ "${#p}" -gt 1 ]; then
    # Check if it's a Windows drive root (e.g., "C:/")
    if [[ "$p" =~ ^[A-Za-z]:/$ ]]; then
      : # Keep the trailing slash for drive roots
    else
      p="${p%/}"
    fi
  fi

  echo "$p"
}

# Convert path to native format for the current OS
# On Windows Git Bash: converts to Windows path if needed
# On Unix: returns the normalized path as-is
to_native_path() {
  local p="$1"
  [ -z "$p" ] && echo "" && return 0

  p="$(normalize_path "$p")"

  if is_windows; then
    # In Git Bash/MSYS2, we typically want Unix-style paths
    # But if the caller needs a Windows path, they should use cygpath
    if command -v cygpath >/dev/null 2>&1; then
      cygpath -u "$p" 2>/dev/null || echo "$p"
    else
      echo "$p"
    fi
  else
    echo "$p"
  fi
}

# Convert path to Windows format (if on Windows)
to_windows_path() {
  local p="$1"
  [ -z "$p" ] && echo "" && return 0

  if is_windows && command -v cygpath >/dev/null 2>&1; then
    cygpath -w "$p" 2>/dev/null || echo "$p"
  else
    echo "$p"
  fi
}

# ============================================================================
# Path Comparison
# ============================================================================

# Compare two paths for equality after normalization
# Returns 0 (true) if paths are equivalent, 1 (false) otherwise
paths_equal() {
  local p1="$1"
  local p2="$2"

  p1="$(normalize_path "$p1")"
  p2="$(normalize_path "$p2")"

  # Case-insensitive comparison on Windows
  if is_windows; then
    [[ "${p1,,}" == "${p2,,}" ]]
  else
    [ "$p1" = "$p2" ]
  fi
}

# Check if path1 is under path2 (path2 is a parent of path1)
# Example: is_path_under "/home/user/project/file.txt" "/home/user" -> true
is_path_under() {
  local child="$1"
  local parent="$2"

  child="$(normalize_path "$child")"
  parent="$(normalize_path "$parent")"

  # Ensure parent ends with / for comparison
  [[ "$parent" != */ ]] && parent="${parent}/"

  if is_windows; then
    [[ "${child,,}/" == "${parent,,}"* ]] || [[ "${child,,}" == "${parent%/}" ]]
  else
    [[ "${child}/" == "${parent}"* ]] || [ "$child" = "${parent%/}" ]
  fi
}

# ============================================================================
# Utility Functions
# ============================================================================

# Get the directory portion of a path (cross-platform dirname)
get_dirname() {
  local p="$1"
  p="$(normalize_path "$p")"

  # Use bash parameter expansion for portability
  local dir="${p%/*}"

  # Handle edge cases
  if [ "$dir" = "$p" ]; then
    # No slash found - current directory
    echo "."
  elif [ -z "$dir" ]; then
    # Path was just "/something" - root
    echo "/"
  else
    echo "$dir"
  fi
}

# Get the filename portion of a path (cross-platform basename)
get_basename() {
  local p="$1"
  p="$(normalize_path "$p")"

  # Remove trailing slash first
  p="${p%/}"

  # Get the last component
  echo "${p##*/}"
}

# Get file extension (without the dot)
get_extension() {
  local p="$1"
  local filename
  filename="$(get_basename "$p")"

  # Check if there's a dot (and it's not at the beginning)
  if [[ "$filename" == *.* ]] && [[ "$filename" != .* ]]; then
    echo "${filename##*.}"
  else
    echo ""
  fi
}

# ============================================================================
# sed Compatibility
# ============================================================================

# Cross-platform sed in-place editing
# Usage: sed_inplace 's/old/new/' file.txt
sed_inplace() {
  local expr="$1"
  local file="$2"

  if is_macos; then
    sed -i '' "$expr" "$file"
  else
    sed -i "$expr" "$file"
  fi
}

# ============================================================================
# Export Functions (for sourcing)
# ============================================================================

# Make functions available when sourced
export -f detect_os is_windows is_macos is_linux 2>/dev/null || true
export -f is_absolute_path is_windows_path 2>/dev/null || true
export -f normalize_path to_native_path to_windows_path 2>/dev/null || true
export -f paths_equal is_path_under 2>/dev/null || true
export -f get_dirname get_basename get_extension 2>/dev/null || true
export -f sed_inplace 2>/dev/null || true
