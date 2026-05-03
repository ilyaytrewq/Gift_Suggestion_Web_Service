#!/usr/bin/env bash
# Загрузка каталога подарков в БД через API импорта.
#
# Что делает:
#   1. Вставляет категории в БД напрямую через psql
#   2. Регистрирует admin-пользователя (если не существует) и повышает роль
#   3. Конвертирует catalog_real.csv → формат импорта бэкенда
#   4. Разбивает на чанки и загружает через POST /api/v1/admin/import-jobs
#   5. Ждёт завершения и показывает итог
#
# Использование:
#   bash scripts/data/import_catalog.sh
#   bash scripts/data/import_catalog.sh --api http://localhost:8080 --chunk 5000
#   bash scripts/data/import_catalog.sh --catalog services/ml/dataset/catalog_real.csv
#   bash scripts/data/import_catalog.sh --max-rows 10000   # быстрый тест

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ── Параметры ─────────────────────────────────────────────────────────────────
API_URL="http://localhost:8080"
CATALOG="${REPO_ROOT}/services/ml/dataset/catalog_real.csv"
ADMIN_EMAIL="admin@gift.local"
ADMIN_PASSWORD="AdminSecret2024!"
CHUNK_SIZE=10000
MAX_ROWS=0  # 0 = все строки
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="gift"
DB_NAME="gift_suggestion"
DB_PASS="gift"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --api)       API_URL="$2";       shift 2 ;;
    --catalog)   CATALOG="$2";       shift 2 ;;
    --chunk)     CHUNK_SIZE="$2";    shift 2 ;;
    --max-rows)  MAX_ROWS="$2";      shift 2 ;;
    --email)     ADMIN_EMAIL="$2";   shift 2 ;;
    --password)  ADMIN_PASSWORD="$2"; shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

log() { echo "[$(date '+%H:%M:%S')] $*"; }

# ── Шаг 1: Категории в БД ─────────────────────────────────────────────────────
log "=== [1/5] Создание категорий в БД ==="

PSQL="psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"
export PGPASSWORD="${DB_PASS}"

$PSQL -c "
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
" 2>&1 | grep -v "^$"

CATS=$($PSQL -t -c "SELECT COUNT(*) FROM categories;" | tr -d ' \n')
log "Категорий в БД: ${CATS}"

# ── Шаг 2: Создание / повышение admin-пользователя ───────────────────────────
log "=== [2/5] Подготовка admin-пользователя ==="

# Регистрируем (игнорируем ошибку если уже существует)
REGISTER_RESP=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "${API_URL}/api/v1/users" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\",\"display_name\":\"Admin\"}" \
  2>/dev/null || true)
log "Регистрация → HTTP ${REGISTER_RESP}"

# Повышаем до admin напрямую в БД
$PSQL -c "
UPDATE users SET role = 'admin'
WHERE email = '${ADMIN_EMAIL}';
" 2>&1 | grep -v "^$"

# ── Шаг 3: Получение JWT-токена ───────────────────────────────────────────────
log "=== [3/5] Получение JWT-токена ==="

LOGIN_RESP=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${ADMIN_EMAIL}\",\"password\":\"${ADMIN_PASSWORD}\"}")

TOKEN=$(echo "${LOGIN_RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d['data']['auth']['access_token'])
except Exception as e:
    print('ERROR: ' + str(e), file=sys.stderr)
    sys.exit(1)
" 2>&1)

if [[ "${TOKEN}" == ERROR* ]] || [[ -z "${TOKEN}" ]]; then
  log "Не удалось получить токен"
  log "Ответ логина: ${LOGIN_RESP}"
  exit 1
fi
log "Токен получен (${#TOKEN} символов)"

# ── Шаг 4: Конвертация CSV ────────────────────────────────────────────────────
log "=== [4/5] Конвертация catalog_real.csv → формат импорта ==="

IMPORT_CSV="${TMP_DIR}/import_ready.csv"

python3 - <<'PYEOF' "${CATALOG}" "${IMPORT_CSV}" "${MAX_ROWS}"
import csv, sys, re

src, dst, max_rows_str = sys.argv[1], sys.argv[2], sys.argv[3]
max_rows = int(max_rows_str)

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}

OUT_FIELDS = ["name", "category", "price", "currency", "description",
              "store_link", "image", "age_restriction", "source"]

def clean_price(raw):
    c = re.sub(r"[^\d.]", "", str(raw))
    try:
        v = float(c)
        return str(int(v)) if v > 0 else ""
    except ValueError:
        return ""

written = skipped = 0
with open(src, encoding="utf-8", errors="replace") as f_in, \
     open(dst, "w", newline="", encoding="utf-8") as f_out:

    reader = csv.DictReader(f_in)
    writer = csv.DictWriter(f_out, fieldnames=OUT_FIELDS)
    writer.writeheader()

    for row in reader:
        if max_rows > 0 and written >= max_rows:
            break

        name = (row.get("title") or "").strip()
        if not name:
            skipped += 1; continue

        cat = (row.get("category") or "").strip()
        if cat not in VALID_CATEGORIES:
            skipped += 1; continue

        price = clean_price(row.get("price", ""))
        if not price:
            skipped += 1; continue

        store_link = (row.get("store_url") or "").strip()
        if not store_link or not store_link.startswith("http"):
            skipped += 1; continue

        desc = (row.get("description") or name).strip()
        if not desc:
            desc = name

        image = (row.get("image_url") or "").strip()
        # Пропускаем невалидные image URL
        if image and not image.startswith("http"):
            image = ""

        currency = (row.get("currency") or "RUB").strip()
        source = (row.get("source") or "catalog").strip()

        # age_restriction: берём age_min, маппим в 0/12/16/18
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
            "name": name[:250],
            "category": cat,
            "price": price,
            "currency": currency,
            "description": desc[:600],
            "store_link": store_link[:400],
            "image": image[:400],
            "age_restriction": age_r,
            "source": source,
        })
        written += 1

print(f"Конвертировано: {written} строк, пропущено: {skipped}")
PYEOF

TOTAL_ROWS=$(wc -l < "${IMPORT_CSV}")
TOTAL_ROWS=$((TOTAL_ROWS - 1))  # минус заголовок
log "Готово к импорту: ${TOTAL_ROWS} строк → ${IMPORT_CSV}"

# ── Шаг 5: Загрузка чанками ──────────────────────────────────────────────────
log "=== [5/5] Загрузка в БД (чанки по ${CHUNK_SIZE} строк) ==="

# Разбиваем на чанки с помощью Python
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
log "Чанков: ${N_CHUNKS}"

TOTAL_IMPORTED=0
TOTAL_SKIPPED=0
FAILED_CHUNKS=0

for i in "${!CHUNK_FILES[@]}"; do
  chunk_file="${CHUNK_FILES[$i]}"
  chunk_num=$((i + 1))
  chunk_rows=$(wc -l < "${chunk_file}")
  chunk_rows=$((chunk_rows - 1))

  log "  Чанк ${chunk_num}/${N_CHUNKS}: ${chunk_rows} строк..."

  RESP=$(curl -s -X POST "${API_URL}/api/v1/admin/import-jobs" \
    -H "Authorization: Bearer ${TOKEN}" \
    -F "file=@${chunk_file};type=text/csv" \
    -F "source=catalog_import" \
    2>/dev/null)

  JOB_ID=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    # job может быть уже завершён синхронно (completed или completed_with_errors)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('id', ''))
except:
    print('')
" 2>/dev/null)

  if [[ -z "${JOB_ID}" ]]; then
    log "    ✗ Ошибка создания job: ${RESP}"
    FAILED_CHUNKS=$((FAILED_CHUNKS + 1))
    continue
  fi

  # Если бэкенд вернул завершённый job синхронно — используем его ответ сразу
  STATUS_RESP="${RESP}"
  STATUS=$(echo "${RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('status', 'pending'))
except:
    print('pending')
" 2>/dev/null)

  # Если ещё не завершён — polling
  if [[ "${STATUS}" != "completed" && "${STATUS}" != "completed_with_errors" && "${STATUS}" != "failed" ]]; then
    for attempt in $(seq 1 60); do
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
      if [[ "${STATUS}" == "completed" || "${STATUS}" == "completed_with_errors" || "${STATUS}" == "failed" ]]; then
        break
      fi
      sleep 2
    done
  fi

  # Итог чанка
  SUMMARY=$(echo "${STATUS_RESP}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    s = j.get('summary', {}) or {}
    imp = s.get('imported_rows', 0)
    upd = s.get('updated_rows', 0)
    skip = s.get('skipped_rows', 0) + s.get('duplicate_in_catalog_rows', 0)
    err = s.get('error_rows', 0)
    print(f'imported={imp} updated={upd} skipped={skip} errors={err} status={j.get(\"status\",\"?\")}')
except Exception as e:
    print(f'parse_error={e}')
" 2>/dev/null)

  log "    ✓ Job ${JOB_ID}: ${SUMMARY}"

  IMP=$(echo "${SUMMARY}" | grep -o 'imported=[0-9]*' | cut -d= -f2 || echo 0)
  SKIP=$(echo "${SUMMARY}" | grep -o 'skipped=[0-9]*' | cut -d= -f2 || echo 0)
  TOTAL_IMPORTED=$((TOTAL_IMPORTED + IMP))
  TOTAL_SKIPPED=$((TOTAL_SKIPPED + SKIP))
done

# ── Итог ─────────────────────────────────────────────────────────────────────
echo ""
echo "=============================================="
log "ГОТОВО!"
echo "  Загружено товаров : ${TOTAL_IMPORTED}"
echo "  Пропущено         : ${TOTAL_SKIPPED}"
if [[ "${FAILED_CHUNKS}" -gt 0 ]]; then
  echo "  ✗ Упавших чанков  : ${FAILED_CHUNKS}"
fi

# Проверка через API
GIFTS_COUNT=$(curl -s "${API_URL}/api/v1/catalog/gifts?limit=1" 2>/dev/null | \
  python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('page',{}).get('total',0))" 2>/dev/null || echo "?")
echo ""
echo "  Товаров в каталоге: ${GIFTS_COUNT}"
echo "  Фронтенд: http://localhost:5173/catalog"
echo "=============================================="
