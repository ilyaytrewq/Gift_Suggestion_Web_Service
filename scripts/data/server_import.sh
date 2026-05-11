#!/usr/bin/env bash
# Импорт каталога прямо на сервере (запускать там, где лежат датасеты).
#
# Использование:
#   bash scripts/data/server_import.sh
#   bash scripts/data/server_import.sh --skip-rows 190000
#   bash scripts/data/server_import.sh --max-rows 1000
#   bash scripts/data/server_import.sh --skip-db-init

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib_import.sh
source "${SCRIPT_DIR}/lib_import.sh"

CATALOG="${HOME}/gift-datasets/dataset/catalog_real.csv"
API_URL="http://localhost:8080"
ADMIN_EMAIL="${IMPORT_ADMIN_EMAIL:-admin@gift.local}"
ADMIN_PASSWORD="${IMPORT_ADMIN_PASSWORD:-AdminSecret2024!}"
COMPOSE_PROJECT="gift-suggestion"
DB_USER="gift"
DB_NAME="gift_suggestion"
CHUNK_SIZE="${IMPORT_CHUNK_SIZE:-${IMPORT_DEFAULT_CHUNK_SIZE:-1000}}"
MAX_ROWS=0
SKIP_ROWS=0
SKIP_DB_INIT=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --catalog)         CATALOG="$2";         shift 2 ;;
    --api)             API_URL="$2";         shift 2 ;;
    --email)           ADMIN_EMAIL="$2";     shift 2 ;;
    --password)        ADMIN_PASSWORD="$2";  shift 2 ;;
    --compose-project) COMPOSE_PROJECT="$2"; shift 2 ;;
    --chunk)           CHUNK_SIZE="$2";      shift 2 ;;
    --max-rows)        MAX_ROWS="$2";        shift 2 ;;
    --skip-rows)       SKIP_ROWS="$2";       shift 2 ;;
    --skip-db-init)    SKIP_DB_INIT=true;    shift ;;
    *) echo "Неизвестный параметр: $1"; exit 1 ;;
  esac
done

[[ ! -f "${CATALOG}" ]] && { echo "Файл не найден: ${CATALOG}"; exit 1; }

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log() { echo "[$(date '+%H:%M:%S')] $*"; }

psql_exec() {
  local container
  container=$(docker ps \
    --filter "label=com.docker.compose.service=postgres" \
    --filter "label=com.docker.compose.project=${COMPOSE_PROJECT}" \
    --format '{{.Names}}' | head -1)
  [[ -z "${container}" ]] && \
    container=$(docker ps --filter 'name=postgres' --format '{{.Names}}' | head -1)
  [[ -z "${container}" ]] && { echo "postgres-контейнер не найден" >&2; exit 1; }
  docker exec -i "${container}" psql -U "${DB_USER}" -d "${DB_NAME}" -c "$1"
}

if [[ "${SKIP_DB_INIT}" == false ]]; then
  log "=== Создание категорий ==="
  psql_exec "INSERT INTO categories (name) VALUES
    ('Электроника'),('Книги'),('Настольные игры'),('Косметика'),
    ('Аксессуары'),('Одежда'),('Хобби'),('Украшения для дома'),
    ('Еда и напитки'),('Детские товары')
  ON CONFLICT (name) DO NOTHING;" | grep -v "^$" || true

  log "=== Создание admin ==="
  curl -s -o /dev/null -X POST "${API_URL}/api/v1/users" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\",\"display_name\":\"Admin\"}" || true
  ADMIN_EMAIL_SQL=$(import_escape_sql_literal "${ADMIN_EMAIL}")
  psql_exec "UPDATE users SET role = 'admin', email_verified_at = COALESCE(email_verified_at, NOW()) WHERE email = '${ADMIN_EMAIL_SQL}';" \
    | grep -v "^$" || true
fi

log "=== Конвертация CSV (skip_rows=${SKIP_ROWS}) ==="
IMPORT_CSV="${TMP_DIR}/import.csv"
import_convert_csv "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}" "${SKIP_ROWS}"

TOTAL_ROWS=$(import_count_csv_rows "${IMPORT_CSV}")
log "Строк для импорта: ${TOTAL_ROWS}"
if [[ "${TOTAL_ROWS}" -eq 0 ]]; then
  log "Нечего импортировать. Готово."
  exit 0
fi

import_split_chunks "${IMPORT_CSV}" "${TMP_DIR}" "${CHUNK_SIZE}"

import_upload_chunks "${API_URL}" "${ADMIN_EMAIL}" "${ADMIN_PASSWORD}" "${TMP_DIR}" log || {
  log "Импорт остановлен: ошибка на чанке (см. лог выше)"
  exit 1
}

echo ""
echo "══════════════════════════════════════════════════"
import_print_summary "${API_URL}" log
echo "══════════════════════════════════════════════════"

[[ "${IMPORT_FAILED_CHUNKS:-0}" -gt 0 ]] && exit 1
