#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="${BACKEND_DIR:-${ROOT_DIR}/services/backend}"
GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-${BACKEND_DIR}/bin/golangci-lint}"

go_wrap() {
  if command -v ya >/dev/null 2>&1; then
    ya tool go "$@"
  else
    go "$@"
  fi
}

cd "${BACKEND_DIR}"

go_wrap test ./...

if [ ! -x "${GOLANGCI_LINT_BIN}" ]; then
  echo "golangci-lint binary not found at ${GOLANGCI_LINT_BIN}" >&2
  exit 1
fi

"${GOLANGCI_LINT_BIN}" run ./... --config=.golangci.yml
go_wrap build ./cmd/api
