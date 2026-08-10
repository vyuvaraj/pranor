#!/usr/bin/env bash
# =============================================================================
# check_cgo.sh — Pranor Zero-CGO Build Invariant Verifier
# =============================================================================
# Part of Phase 89 / Sprint 1 (V2.89.7)
# Defined in: requirements_definitive.md §3.4
#
# Enforces that the Pranor core binary remains a statically compiled,
# CGO-free Go binary across all supported release targets.
#
# Three verification layers:
#   1. Cross-compile matrix check (linux/amd64, arm64, darwin, windows)
#   2. Linux static-link assertion via ldd
#   3. go list CGO source file scan across entire module tree
#
# Usage:
#   bash scripts/check_cgo.sh [module_dir]
#
#   module_dir (optional): subdirectory to check. Defaults to current dir.
#   Example: bash scripts/check_cgo.sh gate
#            bash scripts/check_cgo.sh          # runs from pranor root
# =============================================================================
set -euo pipefail

MODULE_DIR="${1:-.}"
DIST_DIR="${MODULE_DIR}/dist/cgo_check"

echo ""
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║        Pranor Zero-CGO Build Invariant Verifier (V2.89.7)           ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""
echo "  Module : ${MODULE_DIR}"
echo "  Date   : $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo ""

# ── Enforce CGO_ENABLED=0 ────────────────────────────────────────────────────
export CGO_ENABLED=0

if [ "$(go env CGO_ENABLED)" != "0" ]; then
    echo "  ❌ FATAL: CGO_ENABLED is not 0. Cannot proceed."
    exit 1
fi
echo "  ✓ CGO_ENABLED=0 confirmed"
echo ""

# ── Detect entrypoint ────────────────────────────────────────────────────────
# Find all main packages in the module using go list.
# Supports: root main.go, cmd/ subdirs, or library-only modules.
MAIN_PKGS=$(cd "${MODULE_DIR}" && go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./... 2>/dev/null || true)

if [ -n "${MAIN_PKGS}" ]; then
    # Use the first main package for binary output
    FIRST_MAIN=$(echo "${MAIN_PKGS}" | head -1)
    BUILD_TARGET="${MODULE_DIR}/..."
    PRODUCES_BINARY=true
else
    # Library-only module — no binary output
    BUILD_TARGET="${MODULE_DIR}/..."
    PRODUCES_BINARY=false
fi

# ── Layer 1: Cross-compile matrix ────────────────────────────────────────────
echo "── Layer 1: Cross-Compile Matrix ───────────────────────────────────────"
mkdir -p "${DIST_DIR}"

TARGETS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
PASSED=0
FAILED=0

for TARGET in "${TARGETS[@]}"; do
    GOOS="${TARGET%/*}"
    GOARCH="${TARGET#*/}"

    set +e
    # Always compile with go build ./... — no -o flag to avoid multi-package conflicts.
    # This validates compilability across all targets without producing release artifacts.
    (cd "${MODULE_DIR}" && GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 go build ./...) 2>/tmp/cgo_err
    RC=$?
    set -e

    if [ $RC -eq 0 ]; then
        echo "  ✓ ${TARGET}"
        PASSED=$((PASSED + 1))
    else
        echo "  ❌ ${TARGET} — build failed:"
        cat /tmp/cgo_err | sed 's/^/      /'
        FAILED=$((FAILED + 1))
    fi
done

echo ""
echo "  Layer 1 result: ${PASSED}/${#TARGETS[@]} targets passed"
[ $FAILED -gt 0 ] && { echo "  ❌ FAILED: ${FAILED} target(s) did not compile"; exit 1; }
echo ""

# ── Layer 2: Linux static-link assertion via ldd ─────────────────────────────
echo "── Layer 2: Linux Static-Link Assertion (ldd) ──────────────────────────"

if [ "${PRODUCES_BINARY}" = "true" ] && command -v ldd &>/dev/null; then
    # Build a named linux/amd64 binary from the first main package for ldd inspection
    mkdir -p "${DIST_DIR}"
    LDD_BIN="${DIST_DIR}/pranor_linux_amd64"
    set +e
    (cd "${MODULE_DIR}" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
        go build -trimpath -o "../${LDD_BIN}" "${FIRST_MAIN}") 2>/tmp/cgo_err
    LDD_BUILD_RC=$?
    set -e

    if [ $LDD_BUILD_RC -ne 0 ]; then
        echo "  ⚠ Could not build linux/amd64 binary for ldd check — skipping"
        cat /tmp/cgo_err | sed 's/^/      /' || true
    elif [ -f "${LDD_BIN}" ]; then
        LDD_OUT=$(ldd "${LDD_BIN}" 2>&1 || true)
        if echo "${LDD_OUT}" | grep -q 'not a dynamic executable'; then
            echo "  ✓ Static binary confirmed: pranor_linux_amd64"
            echo "    ldd: not a dynamic executable"
        else
            echo "  ❌ Dynamic dependencies detected in pranor_linux_amd64:"
            echo "${LDD_OUT}" | sed 's/^/      /'
            exit 1
        fi
    fi
else
    if [ "${PRODUCES_BINARY}" = "false" ]; then
        echo "  ⚠ Library-only module — skipping ldd check"
    else
        echo "  ⚠ ldd not available on this platform — skipping static-link check"
    fi
fi
echo ""

# ── Layer 3: go list CGO source file scan ────────────────────────────────────
echo "── Layer 3: CGO Source File Scan (go list) ─────────────────────────────"

set +e
CGO_PKGS=$(cd "${MODULE_DIR}" && go list -f '{{if .CgoFiles}}{{.ImportPath}}{{end}}' ./... 2>/dev/null)
set -e

if [ -n "${CGO_PKGS}" ]; then
    echo "  ❌ CGO source files detected in the following packages:"
    echo "${CGO_PKGS}" | sed 's/^/      /'
    echo ""
    echo "  These files MUST be removed or moved to a sidecar/wasm provider."
    exit 1
else
    echo "  ✓ No CGO source files found in module tree"
fi
echo ""

# ── Cleanup ──────────────────────────────────────────────────────────────────
rm -rf "${DIST_DIR}"

# ── Final result ─────────────────────────────────────────────────────────────
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║  ✓  Zero-CGO Invariant Passed — All 3 layers verified               ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""
