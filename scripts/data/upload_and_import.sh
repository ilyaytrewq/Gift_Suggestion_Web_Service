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
ADMIN_EMAIL="admin@gift.local"
ADMIN_PASSWORD="AdminSecret2024!"
CHUNK_SIZE=5000
MAX_ROWS=0
SKIP_ROWS=0
SKIP_UPLOAD=false
SKIP_DB_INIT=false

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
  --help                   Показать эту справку

Примеры:
  # Полный запуск
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154

  # Тест на 1000 строках без переноса файлов
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload --max-rows 1000

  # Продолжить упавший запуск начиная со строки 190000
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload --skip-db-init --skip-rows 190000

  # Только перенос датасетов (без импорта в БД)
  bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --max-rows 0
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

  # Формируем rsync-опции
  RSYNC_OPTS="-avz --progress --partial"
  if [[ -n "${SSH_KEY}" ]]; then
    RSYNC_OPTS="${RSYNC_OPTS} -e 'ssh ${SSH_OPTS}'"
  else
    RSYNC_OPTS="${RSYNC_OPTS} -e 'ssh ${SSH_OPTS}'"
  fi

  # rsync датасетов (CSV + parquet)
  if [[ -d "${DATASETS_DIR}" ]]; then
    log "rsync датасетов: ${DATASETS_DIR}/ → ${SSH_TARGET}:${REMOTE_DIR}/dataset/"
    rsync -avz --progress --partial \
      -e "ssh ${SSH_OPTS}" \
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

# Если MAX_ROWS=0 и skip-db-init, только загружали файлы — выходим
if [[ "${MAX_ROWS}" == "0" ]] && [[ "${SKIP_DB_INIT}" == true ]]; then
  log "MAX_ROWS=0 и --skip-db-init: только перенос файлов. Готово."
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

  remote_psql "UPDATE users SET role = 'admin' WHERE email = '${ADMIN_EMAIL}';" \
    | grep -v "^$" || true
  log "Роль admin установлена для ${ADMIN_EMAIL}"
fi

# ── Шаг 4: JWT-токен ──────────────────────────────────────────────────────────
step "[4/5] Авторизация и конвертация CSV"

TOKEN=""
TOKEN_OBTAINED_AT=0
TOKEN_TTL=720  # обновляем каждые 12 мин (токен живёт 15 мин)

fetch_token() {
  local resp
  resp=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}")

  local tok
  tok=$(echo "${resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d['data']['auth']['access_token'])
except Exception as e:
    print('ERROR: ' + str(e), file=sys.stderr)
    sys.exit(1)
" 2>&1)

  if [[ "${tok}" == ERROR* ]] || [[ -z "${tok}" ]]; then
    log "Ответ логина: ${resp}"
    die "Не удалось получить JWT-токен"
  fi
  TOKEN="${tok}"
  TOKEN_OBTAINED_AT="${SECONDS}"
  log "Токен получен (${#TOKEN} символов)"
}

ensure_token() {
  local elapsed=$(( SECONDS - TOKEN_OBTAINED_AT ))
  if [[ -z "${TOKEN}" ]] || [[ "${elapsed}" -ge "${TOKEN_TTL}" ]]; then
    log "Обновление токена (прошло ${elapsed}с)..."
    fetch_token
  fi
}

fetch_token

# ── Конвертация CSV ────────────────────────────────────────────────────────────
log "Конвертация $(basename "${CATALOG}") → формат импорта..."

IMPORT_CSV="${TMP_DIR}/import_ready.csv"

python3 - <<'PYEOF' "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}" "${SKIP_ROWS}"
import csv, sys, re

src, dst, max_rows_str, skip_rows_str = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
max_rows = int(max_rows_str)
skip_rows = int(skip_rows_str)

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}

OUT_FIELDS = ["name", "category", "price", "currency",
              "description", "store_link", "image", "age_restriction", "source"]

def clean_price(raw):
    c = re.sub(r"[^\d.]", "", str(raw))
    try:
        v = float(c)
        return str(int(round(v))) if v > 0 else ""
    except ValueError:
        return ""

written = skipped = total_read = 0
with open(src, encoding="utf-8", errors="replace") as f_in, \
     open(dst, "w", newline="", encoding="utf-8") as f_out:

    reader = csv.DictReader(f_in)
    writer = csv.DictWriter(f_out, fieldnames=OUT_FIELDS)
    writer.writeheader()

    for row in reader:
        total_read += 1
        if total_read <= skip_rows:
            continue
        if max_rows > 0 and written >= max_rows:
            break

        name = (row.get("title") or row.get("name") or "").strip()
        if not name:
            skipped += 1; continue

        cat = (row.get("category") or "").strip()
        if cat not in VALID_CATEGORIES:
            skipped += 1; continue

        price = clean_price(row.get("price", ""))
        if not price:
            skipped += 1; continue

        store_link = (row.get("store_url") or row.get("store_link") or "").strip()
        if not store_link or not store_link.startswith("http"):
            skipped += 1; continue

        desc = (row.get("description") or name).strip() or name

        image = (row.get("image_url") or row.get("image") or "").strip()
        if image and not image.startswith("http"):
            image = ""

        currency = (row.get("currency") or "RUB").strip()
        source = (row.get("source") or "catalog").strip()

        age_raw = (row.get("age_min") or "0").strip()
        try:
            age_val = int(float(age_raw))
        except ValueError:
            age_val = 0
        if age_val >= 18:
            age_r = "18"
        elif age_val >= 16:
            age_r = "16"
        elif age_val >= 12:
            age_r = "12"
        else:
            age_r = "0"

        writer.writerow({
            "name":           name[:250],
            "category":       cat,
            "price":          price,
            "currency":       currency,
            "description":    desc[:600],
            "store_link":     store_link[:400],
            "image":          image[:400],
            "age_restriction": age_r,
            "source":         source,
        })
        written += 1

print(f"Конвертировано: {written} строк, пропущено: {skipped}" + (f" (пропущено по --skip-rows: {skip_rows})" if skip_rows else ""))
PYEOF

TOTAL_ROWS=$(python3 -c "
import csv
with open('${IMPORT_CSV}') as f:
    print(sum(1 for _ in csv.DictReader(f)))
")
log "Готово к импорту: ${TOTAL_ROWS} строк"

if [[ "${TOTAL_ROWS}" -eq 0 ]]; then
  die "После конвертации не осталось строк для импорта"
fi

# ── Шаг 5: Разбивка на чанки и загрузка ──────────────────────────────────────
step "[5/5] Загрузка в БД (чанки по ${CHUNK_SIZE} строк → ${API_URL})"

python3 - <<PYEOF "${IMPORT_CSV}" "${TMP_DIR}" "${CHUNK_SIZE}"
import csv, sys, os, math

src, out_dir, chunk_str = sys.argv[1], sys.argv[2], sys.argv[3]
chunk = int(chunk_str)

with open(src, encoding="utf-8") as f:
    reader = csv.DictReader(f)
    fieldnames = reader.fieldnames
    rows = list(reader)

total = len(rows)
n_chunks = math.ceil(total / chunk)
print(f"Разбиваем {total} строк на {n_chunks} чанков по ~{chunk}")

for i in range(n_chunks):
    chunk_rows = rows[i*chunk:(i+1)*chunk]
    chunk_file = os.path.join(out_dir, f"chunk_{i+1:04d}.csv")
    with open(chunk_file, "w", newline="", encoding="utf-8") as out:
        writer = csv.DictWriter(out, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(chunk_rows)
PYEOF

CHUNK_FILES=("${TMP_DIR}"/chunk_*.csv)
N_CHUNKS=${#CHUNK_FILES[@]}
log "Чанков для загрузки: ${N_CHUNKS}"

TOTAL_IMPORTED=0
TOTAL_SKIPPED=0
TOTAL_ERRORS=0
FAILED_CHUNKS=0

for i in "${!CHUNK_FILES[@]}"; do
  chunk_file="${CHUNK_FILES[$i]}"
  chunk_num=$((i + 1))
  chunk_rows=$(python3 -c "
import csv
with open('${chunk_file}') as f:
    print(sum(1 for _ in csv.DictReader(f)))
")

  ensure_token
  log "  Чанк ${chunk_num}/${N_CHUNKS}: ${chunk_rows} строк..."

  # Загружаем с retry (до 3 попыток, backoff 10/30/60с)
  RESP=""
  JOB_ID=""
  for retry in 1 2 3; do
    ensure_token
    RESP=$(curl -s --max-time 180 \
      -X POST "${API_URL}/api/v1/admin/import-jobs" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "file=@${chunk_file};type=text/csv" \
      -F "source=catalog_import" \
      2>/dev/null)

    JOB_ID=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('id', ''))
except:
    print('')
" 2>/dev/null)

    [[ -n "${JOB_ID}" ]] && break

    RETRY_WAIT=$(( retry * 20 ))
    log "    Попытка ${retry}/3 неудачна: '${RESP}' — ждём ${RETRY_WAIT}с..."
    sleep "${RETRY_WAIT}"
  done

  if [[ -z "${JOB_ID}" ]]; then
    log "    ✗ Чанк ${chunk_num} не загружен после 3 попыток: ${RESP}"
    FAILED_CHUNKS=$((FAILED_CHUNKS + 1))
    continue
  fi

  # Проверяем статус — бэкенд может вернуть завершённый job синхронно
  STATUS=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('status', 'pending'))
except:
    print('pending')
" 2>/dev/null)

  STATUS_RESP="${RESP}"

  # Polling до завершения (макс 2 минуты)
  if [[ "${STATUS}" != "completed" && "${STATUS}" != "completed_with_errors" && "${STATUS}" != "failed" ]]; then
    for attempt in $(seq 1 60); do
      sleep 2
      STATUS_RESP=$(curl -s "${API_URL}/api/v1/admin/import-jobs/${JOB_ID}" \
        -H "Authorization: Bearer ${TOKEN}" 2>/dev/null)
      STATUS=$(echo "${STATUS_RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d['data']['job']
    print(j['status'])
except:
    print('unknown')
" 2>/dev/null)
      [[ "${STATUS}" == "completed" || "${STATUS}" == "completed_with_errors" || "${STATUS}" == "failed" ]] && break
    done
  fi

  SUMMARY=$(echo "${STATUS_RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    s = j.get('summary', {}) or {}
    imp  = s.get('imported_rows', 0)
    upd  = s.get('updated_rows', 0)
    skip = s.get('skipped_rows', 0) + s.get('duplicate_in_catalog_rows', 0)
    err  = s.get('error_rows', 0)
    st   = j.get('status', '?')
    print(f'imported={imp} updated={upd} skipped={skip} errors={err} status={st}')
except Exception as e:
    print(f'parse_error={e}')
" 2>/dev/null)

  log "    ✓ Job ${JOB_ID}: ${SUMMARY}"

  IMP=$(echo  "${SUMMARY}" | grep -oE 'imported=[0-9]+'  | cut -d= -f2 || echo 0)
  SKIP=$(echo "${SUMMARY}" | grep -oE 'skipped=[0-9]+'   | cut -d= -f2 || echo 0)
  ERR=$(echo  "${SUMMARY}" | grep -oE 'errors=[0-9]+'    | cut -d= -f2 || echo 0)
  TOTAL_IMPORTED=$((TOTAL_IMPORTED + ${IMP:-0}))
  TOTAL_SKIPPED=$((TOTAL_SKIPPED   + ${SKIP:-0}))
  TOTAL_ERRORS=$((TOTAL_ERRORS     + ${ERR:-0}))

  # Показываем примеры ошибок для первого чанка с нулевым импортом
  if [[ "${IMP:-0}" -eq 0 && "${ERR:-0}" -gt 0 && "${chunk_num}" -eq 1 ]]; then
    log "    Примеры ошибок (первые 3):"
    curl -s "${API_URL}/api/v1/admin/import-jobs/${JOB_ID}/errors?limit=3" \
      -H "Authorization: Bearer ${TOKEN}" 2>/dev/null | python3 -c "
import sys, json
try:
    errs = json.load(sys.stdin).get('data', {}).get('errors', [])
    for e in errs:
        row = e.get('row_number', '?')
        field = e.get('field_name') or '-'
        code = e.get('error_code', '?')
        msg = e.get('message', '?')
        raw = json.dumps(e.get('raw_record') or {}, ensure_ascii=False)[:120]
        print(f'      row={row} field={field} [{code}] {msg}')
        print(f'      raw: {raw}')
except Exception as ex:
    print(f'      (не удалось распарсить: {ex})')
" 2>/dev/null || true
  fi
done

# ── Итог ──────────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════"
log "ГОТОВО!"
echo "  Загружено товаров : ${TOTAL_IMPORTED}"
echo "  Пропущено         : ${TOTAL_SKIPPED}"
echo "  Ошибок в строках  : ${TOTAL_ERRORS}"
if [[ "${FAILED_CHUNKS}" -gt 0 ]]; then
  echo "  ✗ Упавших чанков  : ${FAILED_CHUNKS}"
fi

GIFTS_TOTAL=$(curl -s "${API_URL}/api/v1/catalog/gifts?limit=1" 2>/dev/null | \
  python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('page', {}).get('total', 0))
except:
    print('?')
" 2>/dev/null || echo "?")
echo ""
echo "  Товаров в каталоге: ${GIFTS_TOTAL}"
echo "  API              : ${API_URL}"
echo "══════════════════════════════════════════════════════"
