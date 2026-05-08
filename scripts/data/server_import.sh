#!/usr/bin/env bash
# Импорт каталога прямо на сервере (запускать там, где лежат датасеты).
#
# Использование:
#   bash server_import.sh
#   bash server_import.sh --skip-rows 190000   # продолжить с позиции
#   bash server_import.sh --max-rows 1000       # тест

set -uo pipefail

CATALOG="${HOME}/gift-datasets/dataset/catalog_real.csv"
API_URL="http://localhost:8080"
ADMIN_EMAIL="admin@gift.local"
ADMIN_PASSWORD="AdminSecret2024!"
COMPOSE_PROJECT="gift-suggestion"
DB_USER="gift"
DB_NAME="gift_suggestion"
CHUNK_SIZE=2000
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

# ── Категории и admin ─────────────────────────────────────────────────────────
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
  psql_exec "UPDATE users SET role = 'admin' WHERE email = '${ADMIN_EMAIL}';" \
    | grep -v "^$" || true
fi

# ── JWT-токен ─────────────────────────────────────────────────────────────────
TOKEN=""
TOKEN_OBTAINED_AT=0
TOKEN_TTL=720

fetch_token() {
  local resp tok
  resp=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}")
  tok=$(echo "${resp}" | python3 -c "
import sys, json
try: print(json.load(sys.stdin)['data']['auth']['access_token'])
except: sys.exit(1)
" 2>/dev/null) || { log "Не удалось получить токен. Ответ: ${resp}"; exit 1; }
  TOKEN="${tok}"
  TOKEN_OBTAINED_AT="${SECONDS}"
  log "Токен получен (${#TOKEN} символов)"
}

ensure_token() {
  (( SECONDS - TOKEN_OBTAINED_AT >= TOKEN_TTL )) && { log "Обновление токена..."; fetch_token; } || true
}

fetch_token

# ── Конвертация CSV ───────────────────────────────────────────────────────────
log "=== Конвертация CSV (skip_rows=${SKIP_ROWS}) ==="
IMPORT_CSV="${TMP_DIR}/import.csv"

python3 - <<'PYEOF' "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}" "${SKIP_ROWS}"
import csv, sys, re

src, dst, max_rows_str, skip_rows_str = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
max_rows, skip_rows = int(max_rows_str), int(skip_rows_str)

VALID_CATEGORIES = {
    "Электроника","Книги","Настольные игры","Косметика",
    "Аксессуары","Одежда","Хобби","Украшения для дома",
    "Еда и напитки","Детские товары",
}
OUT_FIELDS = ["name","category","price","currency","description",
              "store_link","image","age_restriction","source"]

def clean_price(raw):
    c = re.sub(r"[^\d.]", "", str(raw))
    try:
        v = float(c); return str(int(round(v))) if v > 0 else ""
    except ValueError:
        return ""

written = skipped = total_read = 0
with open(src, encoding="utf-8", errors="replace") as fi, \
     open(dst, "w", newline="", encoding="utf-8") as fo:
    reader = csv.DictReader(fi)
    writer = csv.DictWriter(fo, fieldnames=OUT_FIELDS)
    writer.writeheader()
    for row in reader:
        total_read += 1
        if total_read <= skip_rows:
            continue
        if max_rows > 0 and written >= max_rows:
            break
        name  = (row.get("title") or row.get("name") or "").strip()
        cat   = (row.get("category") or "").strip()
        price = clean_price(row.get("price", ""))
        link  = (row.get("store_url") or row.get("store_link") or "").strip()
        if not name or cat not in VALID_CATEGORIES or not price or not link.startswith("http"):
            skipped += 1; continue
        desc  = (row.get("description") or name).strip() or name
        image = (row.get("image_url") or row.get("image") or "").strip()
        if image and not image.startswith("http"): image = ""
        try:   age_val = int(float((row.get("age_min") or "0").strip()))
        except ValueError: age_val = 0
        age_r = "18" if age_val>=18 else "16" if age_val>=16 else "12" if age_val>=12 else "0"
        writer.writerow({
            "name":            name[:250],
            "category":        cat,
            "price":           price,
            "currency":        (row.get("currency") or "RUB").strip(),
            "description":     desc[:600],
            "store_link":      link[:400],
            "image":           image[:400],
            "age_restriction": age_r,
            "source":          (row.get("source") or "catalog").strip(),
        })
        written += 1

msg = f"Конвертировано: {written}, пропущено: {skipped}"
if skip_rows: msg += f" (пропущено по --skip-rows: {skip_rows})"
print(msg)
PYEOF

TOTAL_ROWS=$(python3 -c "
import csv
with open('${IMPORT_CSV}') as f:
    print(sum(1 for _ in csv.DictReader(f)))
")
log "Строк для импорта: ${TOTAL_ROWS}"
[[ "${TOTAL_ROWS}" -eq 0 ]] && { log "Нечего импортировать. Готово."; exit 0; }

# ── Разбивка на чанки ─────────────────────────────────────────────────────────
python3 - <<PYEOF "${IMPORT_CSV}" "${TMP_DIR}" "${CHUNK_SIZE}"
import csv, sys, os, math
src, out_dir, chunk_str = sys.argv[1], sys.argv[2], sys.argv[3]
chunk = int(chunk_str)
with open(src) as f:
    reader = csv.DictReader(f)
    fields = reader.fieldnames
    rows = list(reader)
n = math.ceil(len(rows) / chunk)
print(f"Разбиваем {len(rows)} строк на {n} чанков по ~{chunk}")
for i in range(n):
    path = os.path.join(out_dir, f"chunk_{i+1:04d}.csv")
    with open(path, "w", newline="") as out:
        w = csv.DictWriter(out, fieldnames=fields)
        w.writeheader()
        w.writerows(rows[i*chunk:(i+1)*chunk])
PYEOF

CHUNK_FILES=("${TMP_DIR}"/chunk_*.csv)
N_CHUNKS=${#CHUNK_FILES[@]}
log "Чанков: ${N_CHUNKS}"

TOTAL_IMPORTED=0
TOTAL_SKIPPED=0
TOTAL_ERRORS=0
FAILED_CHUNKS=0

for i in "${!CHUNK_FILES[@]}"; do
  chunk_file="${CHUNK_FILES[$i]}"
  chunk_num=$((i + 1))

  ensure_token
  log "  Чанк ${chunk_num}/${N_CHUNKS}..."

  RESP=""
  JOB_ID=""
  for retry in 1 2 3; do
    ensure_token
    RESP=$(curl -s --max-time 180 \
      -X POST "${API_URL}/api/v1/admin/import-jobs" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "file=@${chunk_file};type=text/csv" \
      -F "source=catalog_import" 2>/dev/null)
    JOB_ID=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('id', ''))
except: print('')
" 2>/dev/null)
    [[ -n "${JOB_ID}" ]] && break
    log "    Попытка ${retry}/3: '${RESP}' — ждём $((retry*20))с..."
    sleep $((retry * 20))
  done

  if [[ -z "${JOB_ID}" ]]; then
    log "    ✗ Чанк ${chunk_num} не загружен после 3 попыток: ${RESP}"
    FAILED_CHUNKS=$((FAILED_CHUNKS + 1))
    continue
  fi

  STATUS=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('status', 'pending'))
except: print('pending')
" 2>/dev/null)
  STATUS_RESP="${RESP}"

  if [[ "${STATUS}" != "completed" && "${STATUS}" != "completed_with_errors" && "${STATUS}" != "failed" ]]; then
    for _ in $(seq 1 60); do
      sleep 2
      STATUS_RESP=$(curl -s "${API_URL}/api/v1/admin/import-jobs/${JOB_ID}" \
        -H "Authorization: Bearer ${TOKEN}" 2>/dev/null)
      STATUS=$(echo "${STATUS_RESP}" | python3 -c "
import sys, json
try: print(json.load(sys.stdin)['data']['job']['status'])
except: print('unknown')
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
    print(f'imported={imp} updated={upd} skipped={skip} errors={err} status={j.get(\"status\",\"?\")}')
except Exception as e: print(f'parse_error={e}')
" 2>/dev/null)

  log "    ✓ Job ${JOB_ID}: ${SUMMARY}"

  IMP=$(echo  "${SUMMARY}" | grep -oE 'imported=[0-9]+' | cut -d= -f2 || echo 0)
  SKIP=$(echo "${SUMMARY}" | grep -oE 'skipped=[0-9]+'  | cut -d= -f2 || echo 0)
  ERR=$(echo  "${SUMMARY}" | grep -oE 'errors=[0-9]+'   | cut -d= -f2 || echo 0)
  TOTAL_IMPORTED=$((TOTAL_IMPORTED + ${IMP:-0}))
  TOTAL_SKIPPED=$((TOTAL_SKIPPED   + ${SKIP:-0}))
  TOTAL_ERRORS=$((TOTAL_ERRORS     + ${ERR:-0}))

  if [[ "${IMP:-0}" -eq 0 && "${ERR:-0}" -gt 0 && "${chunk_num}" -eq 1 ]]; then
    log "    Примеры ошибок:"
    curl -s "${API_URL}/api/v1/admin/import-jobs/${JOB_ID}/errors?limit=3" \
      -H "Authorization: Bearer ${TOKEN}" 2>/dev/null | python3 -c "
import sys, json
try:
    for e in json.load(sys.stdin).get('data', {}).get('errors', []):
        print(f'      row={e.get(\"row_number\",\"?\")} [{e.get(\"error_code\",\"?\")}] {e.get(\"message\",\"?\")}')
except: pass
" 2>/dev/null || true
  fi
done

echo ""
echo "══════════════════════════════════════════════════"
log "ГОТОВО!"
echo "  Загружено : ${TOTAL_IMPORTED}"
echo "  Пропущено : ${TOTAL_SKIPPED}"
echo "  Ошибок    : ${TOTAL_ERRORS}"
[[ "${FAILED_CHUNKS}" -gt 0 ]] && echo "  ✗ Упавших чанков: ${FAILED_CHUNKS}"
GIFTS_TOTAL=$(curl -s "${API_URL}/api/v1/catalog/gifts?limit=1" 2>/dev/null | \
  python3 -c "
import sys, json
try: print(json.load(sys.stdin).get('data',{}).get('page',{}).get('total',0))
except: print('?')
" 2>/dev/null || echo "?")
echo "  Товаров в каталоге: ${GIFTS_TOTAL}"
echo "══════════════════════════════════════════════════"
