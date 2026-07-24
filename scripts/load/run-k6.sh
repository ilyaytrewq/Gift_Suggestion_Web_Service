#!/usr/bin/env bash
# Run k6 load tests against a running Gift Suggestion stack.
#
# Prerequisites:
#   - k6 installed: https://grafana.com/docs/k6/latest/set-up/install-k6/
#   - backend reachable (default http://localhost:8080)
#   - catalog imported for meaningful catalog/recommendation results
#
# Examples:
#   ./scripts/load/run-k6.sh smoke
#   BASE_URL=http://localhost ./scripts/load/run-k6.sh mixed
#   K6_VUS=50 K6_DURATION=5m ./scripts/load/run-k6.sh mixed
#   K6_TEST_EMAIL=user@example.com K6_TEST_PASSWORD=secret ./scripts/load/run-k6.sh smoke

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCENARIO="${1:-smoke}"

case "$SCENARIO" in
  smoke | mixed | recommendations) ;;
  *)
    echo "Usage: $0 [smoke|mixed|recommendations]" >&2
    exit 1
    ;;
esac

if ! command -v k6 >/dev/null 2>&1; then
  echo "k6 is not installed. See https://grafana.com/docs/k6/latest/set-up/install-k6/" >&2
  exit 1
fi

export BASE_URL="${BASE_URL:-http://localhost:8080}"

cd "$ROOT"
exec k6 run "loadtests/${SCENARIO}.js"
