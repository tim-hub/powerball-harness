#!/bin/bash
#
#   cd go && bash scripts/build-all.sh
#
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${GO_DIR}/.." && pwd)"

VERSION=$(cat "${REPO_ROOT}/VERSION" 2>/dev/null || echo "dev")
LDFLAGS="-s -w -X main.version=${VERSION}"
OUTDIR="${REPO_ROOT}/bin"
mkdir -p "${OUTDIR}"

platforms=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
)

echo "Building harness v${VERSION} for ${#platforms[@]} platforms..."

for platform in "${platforms[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "${platform}"
  output="${OUTDIR}/harness-${GOOS}-${GOARCH}"
  echo "  Building ${output}..."
  (cd "${GO_DIR}" && CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build -ldflags="${LDFLAGS}" -o "${output}" ./cmd/harness/)
done

echo ""
echo "Done. Built ${#platforms[@]} binaries:"
ls -lh "${OUTDIR}"/harness-* 2>/dev/null || echo "  (no binaries found in ${OUTDIR})"
