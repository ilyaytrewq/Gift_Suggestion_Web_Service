#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/docker-compose.yml}"
BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://127.0.0.1:8080/health/ready}"

cleanup() {
  docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans
}

trap cleanup EXIT

docker compose -f "${COMPOSE_FILE}" config -q
docker compose -f "${COMPOSE_FILE}" build backend frontend
docker compose -f "${COMPOSE_FILE}" up -d postgres backend

backend_container_id="$(docker compose -f "${COMPOSE_FILE}" ps -q backend)"
if [ -z "${backend_container_id}" ]; then
  echo "Backend container was not created" >&2
  exit 1
fi

attempt=1
max_attempts=30
while [ "${attempt}" -le "${max_attempts}" ]; do
  health_status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}unknown{{end}}' "${backend_container_id}")"
  if [ "${health_status}" = "healthy" ]; then
    break
  fi

  if [ "${attempt}" -eq "${max_attempts}" ]; then
    echo "Backend container did not become healthy" >&2
    docker compose -f "${COMPOSE_FILE}" logs postgres backend
    exit 1
  fi

  sleep 2
  attempt=$((attempt + 1))
done

curl --fail --silent --show-error "${BACKEND_HEALTH_URL}" >/dev/null
