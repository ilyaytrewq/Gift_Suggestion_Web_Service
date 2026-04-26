#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="${BACKEND_DIR:-${ROOT_DIR}/services/backend}"
GOLANGCI_LINT_BIN="${GOLANGCI_LINT_BIN:-${BACKEND_DIR}/bin/golangci-lint}"

cd "${BACKEND_DIR}"

go test ./...

if [ ! -x "${GOLANGCI_LINT_BIN}" ]; then
  echo "golangci-lint binary not found at ${GOLANGCI_LINT_BIN}" >&2
  exit 1
fi

"${GOLANGCI_LINT_BIN}" run ./... --config=.golangci.yml
go build ./cmd/api
