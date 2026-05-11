#!/usr/bin/env bash
# Перенос датасетов на сервер и импорт товаров в БД через API.
#
# Что делает:
#   1. rsync датасетов (CSV, parquet, модели) на сервер
#   2. Создаёт категории в БД (docker exec psql)
#   3. Регистрирует admin-пользователя и повышает роль (docker exec psql)
#   4. Конвертирует CSV → формат импорта (локально)
#   5. Загружает чанками через POST /api/v1/admin/import-jobs
#
# Требования:
#   локально : ssh, rsync, curl, python3
#   на сервере: docker, docker compose
#
# Использование:
#   bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154
#   bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --key ~/.ssh/id_rsa
#   bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --max-rows 2000  # тест
#   bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload    # только импорт

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
# shellcheck source=lib_import.sh
source "${SCRIPT_DIR}/lib_import.sh"

# ── Параметры по умолчанию ────────────────────────────────────────────────────
SSH_TARGET=""
SSH_KEY=""
API_URL=""
CATALOG="${REPO_ROOT}/services/ml/dataset/catalog_real.csv"
DATASETS_DIR="${REPO_ROOT}/services/ml/dataset"
ML_MODELS_DIR="${REPO_ROOT}/services/ml/models"
REMOTE_DIR="~/gift-datasets"
COMPOSE_PROJECT="gift-suggestion"
DB_USER="gift"
DB_NAME="gift_suggestion"
ADMIN_EMAIL="${IMPORT_ADMIN_EMAIL:-admin@gift.local}"
ADMIN_PASSWORD="${IMPORT_ADMIN_PASSWORD:-AdminSecret2024!}"
CHUNK_SIZE="${IMPORT_CHUNK_SIZE:-${IMPORT_DEFAULT_CHUNK_SIZE:-1000}}"
MAX_ROWS=0
SKIP_ROWS=0
SKIP_UPLOAD=false
SKIP_DB_INIT=false
UPLOAD_ONLY=false

usage() {
  cat <<EOF
Использование: $(basename "$0") [OPTIONS]

Обязательные:
  --server ilyaytrewq@51.250.117.154       SSH-цель (например, ilyaytrewq@51.250.117.154)

Опциональные:
  --key PATH               Путь к SSH-ключу
  --api URL                URL бэкенда (по умолчанию: http://HOST:8080)
  --catalog PATH           Локальный CSV (default: services/ml/dataset/catalog_real.csv)
  --datasets-dir PATH      Локальная директория датасетов ML для rsync
  --remote-dir PATH        Директория на сервере для датасетов (default: ~/gift-datasets)
  --compose-project NAME   Имя Docker Compose проекта (default: gift-suggestion)
  --db-user NAME           Пользователь PostgreSQL (default: gift)
  --db-name NAME           База данных PostgreSQL (default: gift_suggestion)
  --email EMAIL            Email admin-пользователя
  --password PASS          Пароль admin-пользователя
  --chunk N                Строк на чанк при импорте (default: 5000)
  --max-rows N             Максимум строк для импорта, 0 = все (default: 0)
  --skip-rows N            Пропустить первые N строк CSV (для продолжения упавшего запуска)
  --skip-upload            Пропустить rsync датасетов
  --skip-db-init           Пропустить создание категорий и admin
  --upload-only            Только rsync файлов на сервер, без импорта в БД
  --help                   Показать эту справку

Примеры:
  # Полный запуск
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154

  # Тест на 1000 строках без переноса файлов
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload --max-rows 1000

  # Продолжить упавший запуск начиная со строки 190000 исходного CSV
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload --skip-db-init --skip-rows 190000

  # Только перенос датасетов (без импорта в БД)
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --upload-only
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server)         SSH_TARGET="$2";       shift 2 ;;
    --key)            SSH_KEY="$2";          shift 2 ;;
    --api)            API_URL="$2";          shift 2 ;;
    --catalog)        CATALOG="$2";          shift 2 ;;
    --datasets-dir)   DATASETS_DIR="$2";     shift 2 ;;
    --remote-dir)     REMOTE_DIR="$2";       shift 2 ;;
    --compose-project) COMPOSE_PROJECT="$2"; shift 2 ;;
    --db-user)        DB_USER="$2";          shift 2 ;;
    --db-name)        DB_NAME="$2";          shift 2 ;;
    --email)          ADMIN_EMAIL="$2";      shift 2 ;;
    --password)       ADMIN_PASSWORD="$2";   shift 2 ;;
    --chunk)          CHUNK_SIZE="$2";       shift 2 ;;
    --max-rows)       MAX_ROWS="$2";         shift 2 ;;
    --skip-rows)      SKIP_ROWS="$2";        shift 2 ;;
    --skip-upload)    SKIP_UPLOAD=true;      shift ;;
    --skip-db-init)   SKIP_DB_INIT=true;     shift ;;
    --upload-only)    UPLOAD_ONLY=true;      shift ;;
    --help|-h)        usage; exit 0 ;;
    *) echo "Неизвестный параметр: $1"; usage; exit 1 ;;
  esac
done

if [[ -z "${SSH_TARGET}" ]]; then
  echo "Ошибка: --server обязателен"
  usage
  exit 1
fi

# Автоопределение API URL из SSH-цели
if [[ -z "${API_URL}" ]]; then
  SERVER_HOST="${SSH_TARGET##*@}"
  API_URL="http://${SERVER_HOST}:8080"
fi

# SSH-опции
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes"
if [[ -n "${SSH_KEY}" ]]; then
  SSH_OPTS="${SSH_OPTS} -i ${SSH_KEY}"
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
die()  { echo "[ERROR] $*" >&2; exit 1; }
step() { echo ""; echo "══════════════════════════════════════════════════════"; log "$*"; echo "══════════════════════════════════════════════════════"; }

# ── Проверка зависимостей ──────────────────────────────────────────────────────
check_deps() {
  local missing=()
  for cmd in ssh rsync curl python3; do
    command -v "$cmd" &>/dev/null || missing+=("$cmd")
  done
  if [[ ${#missing[@]} -gt 0 ]]; then
    die "Отсутствуют зависимости: ${missing[*]}"
  fi
}

check_deps
log "Сервер   : ${SSH_TARGET}"
log "API      : ${API_URL}"
log "Каталог  : ${CATALOG}"

# ── Шаг 1: rsync датасетов на сервер ─────────────────────────────────────────
if [[ "${SKIP_UPLOAD}" == true ]]; then
  log "=== [1/5] Пропускаем rsync (--skip-upload) ==="
else
  step "[1/5] Перенос датасетов на сервер (rsync)"

  # Проверяем соединение
  if ! ssh ${SSH_OPTS} "${SSH_TARGET}" "mkdir -p ${REMOTE_DIR}" 2>/dev/null; then
    die "Не удалось подключиться к ${SSH_TARGET}"
  fi
  log "Директория на сервере: ${REMOTE_DIR}"

  # rsync датасетов (CSV + parquet), включая подкаталоги (collected/)
  if [[ -d "${DATASETS_DIR}" ]]; then
    log "rsync датасетов: ${DATASETS_DIR}/ → ${SSH_TARGET}:${REMOTE_DIR}/dataset/"
    rsync -avz --progress --partial \
      -e "ssh ${SSH_OPTS}" \
      --include="*/" \
      --include="*.csv" \
      --include="*.parquet" \
      --exclude="*" \
      "${DATASETS_DIR}/" \
      "${SSH_TARGET}:${REMOTE_DIR}/dataset/" \
      || log "Предупреждение: rsync датасетов завершился с ошибкой"
  else
    log "Директория датасетов не найдена: ${DATASETS_DIR}"
  fi

  # rsync моделей ML (если есть)
  if [[ -d "${ML_MODELS_DIR}" ]] && [[ -n "$(ls -A "${ML_MODELS_DIR}" 2>/dev/null)" ]]; then
    log "rsync моделей ML: ${ML_MODELS_DIR}/ → ${SSH_TARGET}:${REMOTE_DIR}/models/"
    ssh ${SSH_OPTS} "${SSH_TARGET}" "mkdir -p ${REMOTE_DIR}/models" 2>/dev/null
    rsync -avz --progress --partial \
      -e "ssh ${SSH_OPTS}" \
      "${ML_MODELS_DIR}/" \
      "${SSH_TARGET}:${REMOTE_DIR}/models/" \
      || log "Предупреждение: rsync моделей завершился с ошибкой"
  fi

  log "rsync завершён"
fi

if [[ "${UPLOAD_ONLY}" == true ]]; then
  log "Только перенос файлов (--upload-only). Готово."
  exit 0
fi

# Проверяем каталог
if [[ ! -f "${CATALOG}" ]]; then
  die "Файл каталога не найден: ${CATALOG}"
fi

# ── Вспомогательная функция: psql через docker exec ───────────────────────────
remote_psql() {
  local query="$1"
  ssh ${SSH_OPTS} "${SSH_TARGET}" "
    CONTAINER=\$(docker ps \
      --filter 'label=com.docker.compose.service=postgres' \
      --filter 'label=com.docker.compose.project=${COMPOSE_PROJECT}' \
      --format '{{.Names}}' | head -1)

    if [[ -z \"\${CONTAINER}\" ]]; then
      # Fallback: ищем по имени
      CONTAINER=\$(docker ps --filter 'name=postgres' --format '{{.Names}}' | head -1)
    fi

    if [[ -z \"\${CONTAINER}\" ]]; then
      echo 'Ошибка: postgres-контейнер не найден' >&2
      exit 1
    fi

    echo \"[psql] контейнер: \${CONTAINER}\" >&2
    docker exec -i \"\${CONTAINER}\" psql -U ${DB_USER} -d ${DB_NAME} -c \"${query}\"
  " 2>&1
}

# ── Шаг 2: Категории в БД ─────────────────────────────────────────────────────
if [[ "${SKIP_DB_INIT}" == true ]]; then
  log "=== [2/5] Пропускаем инициализацию БД (--skip-db-init) ==="
else
  step "[2/5] Создание категорий в БД"

  remote_psql "
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
" | grep -v "^$" || true

  CATS_COUNT=$(remote_psql "SELECT COUNT(*) FROM categories;" 2>/dev/null | grep -E '^\s+[0-9]+' | tr -d ' ' || echo "?")
  log "Категорий в БД: ${CATS_COUNT}"
fi

# ── Шаг 3: Admin-пользователь ─────────────────────────────────────────────────
if [[ "${SKIP_DB_INIT}" == true ]]; then
  log "=== [3/5] Пропускаем создание admin (--skip-db-init) ==="
else
  step "[3/5] Создание admin-пользователя"

  REGISTER_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_URL}/api/v1/users" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\",\"display_name\":\"Admin\"}" \
    2>/dev/null || echo "000")
  log "Регистрация → HTTP ${REGISTER_CODE}"

  ADMIN_EMAIL_SQL=$(import_escape_sql_literal "${ADMIN_EMAIL}")
  remote_psql "UPDATE users SET role = 'admin', email_verified_at = COALESCE(email_verified_at, NOW()) WHERE email = '${ADMIN_EMAIL_SQL}';" \
    | grep -v "^$" || true
  log "Роль admin и подтверждение email для ${ADMIN_EMAIL}"
fi

# ── Шаг 4–5: конвертация и загрузка ───────────────────────────────────────────
step "[4/5] Конвертация CSV"
log "Конвертация $(basename "${CATALOG}") → формат импорта..."

IMPORT_CSV="${TMP_DIR}/import_ready.csv"
import_convert_csv "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}" "${SKIP_ROWS}"

TOTAL_ROWS=$(import_count_csv_rows "${IMPORT_CSV}")
log "Готово к импорту: ${TOTAL_ROWS} строк"
[[ "${TOTAL_ROWS}" -eq 0 ]] && die "После конвертации не осталось строк для импорта"

step "[5/5] Загрузка в БД (чанки по ${CHUNK_SIZE} строк → ${API_URL})"
import_split_chunks "${IMPORT_CSV}" "${TMP_DIR}" "${CHUNK_SIZE}"

import_upload_chunks "${API_URL}" "${ADMIN_EMAIL}" "${ADMIN_PASSWORD}" "${TMP_DIR}" log || \
  die "Импорт остановлен: ошибка на чанке (см. лог выше)"

echo ""
echo "══════════════════════════════════════════════════════"
import_print_summary "${API_URL}" log
echo "══════════════════════════════════════════════════════"

[[ "${IMPORT_FAILED_CHUNKS:-0}" -gt 0 ]] && exit 1
