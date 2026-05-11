#!/usr/bin/env bash
# Загрузка каталога подарков в БД через API импорта (локально или через Docker Postgres).
#
# Использование:
#   bash scripts/data/import_catalog.sh
#   bash scripts/data/import_catalog.sh --docker
#   bash scripts/data/import_catalog.sh --api http://localhost:8080 --chunk 5000
#   bash scripts/data/import_catalog.sh --catalog services/ml/dataset/catalog_real.csv
#   bash scripts/data/import_catalog.sh --max-rows 10000 --skip-db-init

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib_import.sh
source "${SCRIPT_DIR}/lib_import.sh"

API_URL="http://localhost:8080"
CATALOG="${REPO_ROOT}/services/ml/dataset/catalog_real.csv"
ADMIN_EMAIL="${IMPORT_ADMIN_EMAIL:-admin@gift.local}"
ADMIN_PASSWORD="${IMPORT_ADMIN_PASSWORD:-AdminSecret2024!}"
CHUNK_SIZE="${IMPORT_CHUNK_SIZE:-${IMPORT_DEFAULT_CHUNK_SIZE:-1000}}"
MAX_ROWS=0
SKIP_ROWS=0
SKIP_DB_INIT=false
USE_DOCKER=false
COMPOSE_PROJECT="gift-suggestion"
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="gift"
DB_NAME="gift_suggestion"
DB_PASS="${PGPASSWORD:-gift}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api)             API_URL="$2";          shift 2 ;;
    --catalog)         CATALOG="$2";          shift 2 ;;
    --chunk)           CHUNK_SIZE="$2";       shift 2 ;;
    --max-rows)        MAX_ROWS="$2";         shift 2 ;;
    --skip-rows)       SKIP_ROWS="$2";        shift 2 ;;
    --email)           ADMIN_EMAIL="$2";      shift 2 ;;
    --password)        ADMIN_PASSWORD="$2";   shift 2 ;;
    --skip-db-init)    SKIP_DB_INIT=true;     shift ;;
    --docker)          USE_DOCKER=true;       shift ;;
    --compose-project) COMPOSE_PROJECT="$2";  shift 2 ;;
    --db-host)         DB_HOST="$2";          shift 2 ;;
    --db-port)         DB_PORT="$2";          shift 2 ;;
    --db-user)         DB_USER="$2";          shift 2 ;;
    --db-name)         DB_NAME="$2";          shift 2 ;;
    --db-pass)         DB_PASS="$2";          shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

[[ ! -f "${CATALOG}" ]] && { echo "Файл не найден: ${CATALOG}"; exit 1; }

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log() { echo "[$(date '+%H:%M:%S')] $*"; }

psql_exec() {
  local query="$1"
  if [[ "${USE_DOCKER}" == true ]]; then
    local container
    container=$(docker ps \
      --filter "label=com.docker.compose.service=postgres" \
      --filter "label=com.docker.compose.project=${COMPOSE_PROJECT}" \
      --format '{{.Names}}' | head -1)
    [[ -z "${container}" ]] && \
      container=$(docker ps --filter 'name=postgres' --format '{{.Names}}' | head -1)
    [[ -z "${container}" ]] && { echo "postgres-контейнер не найден" >&2; return 1; }
    docker exec -i "${container}" psql -U "${DB_USER}" -d "${DB_NAME}" -c "${query}"
  else
    PGPASSWORD="${DB_PASS}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c "${query}"
  fi
}

if [[ "${SKIP_DB_INIT}" == false ]]; then
  log "=== [1/5] Создание категорий в БД ==="
  psql_exec "
INSERT INTO categories (name) VALUES
  ('Электроника'),
  ('Книги'),
  ('Настольные игры'),
  ('Косметика'),
  ('Аксессуары'),
  ('Одежда'),
  ('Хобби'),
  ('Украшения для дома'),
  ('Еда и напитки'),
  ('Детские товары')
ON CONFLICT (name) DO NOTHING;
" 2>&1 | grep -v "^$" || true

  CATS=$(psql_exec "SELECT COUNT(*) FROM categories;" 2>/dev/null | grep -E '[0-9]+' | tr -d ' ' | head -1)
  log "Категорий в БД: ${CATS:-?}"

  log "=== [2/5] Подготовка admin-пользователя ==="
  REGISTER_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_URL}/api/v1/users" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\",\"display_name\":\"Admin\"}" \
    2>/dev/null || echo "000")
  log "Регистрация → HTTP ${REGISTER_CODE}"

  ADMIN_EMAIL_SQL=$(import_escape_sql_literal "${ADMIN_EMAIL}")
  psql_exec "UPDATE users SET role = 'admin', email_verified_at = COALESCE(email_verified_at, NOW()) WHERE email = '${ADMIN_EMAIL_SQL}';" \
    2>&1 | grep -v "^$" || true
else
  log "=== [1-2/5] Пропуск инициализации БД (--skip-db-init) ==="
fi

log "=== [3/5] Конвертация CSV ==="
IMPORT_CSV="${TMP_DIR}/import_ready.csv"
import_convert_csv "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}" "${SKIP_ROWS}"

TOTAL_ROWS=$(import_count_csv_rows "${IMPORT_CSV}")
log "Готово к импорту: ${TOTAL_ROWS} строк"
[[ "${TOTAL_ROWS}" -eq 0 ]] && { log "Нечего импортировать."; exit 0; }

log "=== [4-5/5] Загрузка в БД (чанки по ${CHUNK_SIZE}) ==="
import_split_chunks "${IMPORT_CSV}" "${TMP_DIR}" "${CHUNK_SIZE}"

import_upload_chunks "${API_URL}" "${ADMIN_EMAIL}" "${ADMIN_PASSWORD}" "${TMP_DIR}" log || {
  log "Импорт остановлен: ошибка на чанке (см. лог выше)"
  exit 1
}

echo ""
echo "=============================================="
import_print_summary "${API_URL}" log
echo "=============================================="

[[ "${IMPORT_FAILED_CHUNKS:-0}" -gt 0 ]] && exit 1
