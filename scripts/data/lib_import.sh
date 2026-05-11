# Общие константы и функции для скриптов импорта каталога.
# Подключение: source "$(dirname "${BASH_SOURCE[0]}")/lib_import.sh"

# JWT обновляем за 3 мин до истечения (access token по умолчанию 15 мин)
IMPORT_TOKEN_TTL=720
# Импорт синхронный: polling почти не нужен (оставлен короткий на случай pending)
IMPORT_POLL_ATTEMPTS=5
IMPORT_POLL_INTERVAL=2
IMPORT_CURL_MAX_TIME=600
# Без увеличенного HTTP_READ_TIMEOUT на бэкенде держать <=1000
IMPORT_DEFAULT_CHUNK_SIZE=1000

# shellcheck disable=SC2034
IMPORT_CONVERT_PY='
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
            skipped += 1
            continue

        cat = (row.get("category") or "").strip()
        if cat not in VALID_CATEGORIES:
            skipped += 1
            continue

        price = clean_price(row.get("price", ""))
        if not price:
            skipped += 1
            continue

        store_link = (row.get("store_url") or row.get("store_link") or "").strip()
        if not store_link or not store_link.startswith("http"):
            skipped += 1
            continue

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

msg = f"Конвертировано: {written} строк, пропущено: {skipped}"
if skip_rows:
    msg += f" (пропущено по --skip-rows: {skip_rows})"
print(msg)
'

import_convert_csv() {
  local catalog="$1"
  local import_csv="$2"
  local max_rows="$3"
  local skip_rows="$4"
  python3 -c "${IMPORT_CONVERT_PY}" "${catalog}" "${import_csv}" "${max_rows}" "${skip_rows}"
}

import_count_csv_rows() {
  local csv_file="$1"
  python3 -c "
import csv, sys
with open(sys.argv[1], encoding='utf-8') as f:
    print(sum(1 for _ in csv.DictReader(f)))
" "${csv_file}"
}

import_split_chunks() {
  local import_csv="$1"
  local tmp_dir="$2"
  local chunk_size="$3"
  python3 - <<PYEOF "${import_csv}" "${tmp_dir}" "${chunk_size}"
import csv, sys, os, math

src, out_dir, chunk_str = sys.argv[1], sys.argv[2], sys.argv[3]
chunk = int(chunk_str)

with open(src, encoding="utf-8") as f:
    reader = csv.DictReader(f)
    fieldnames = reader.fieldnames
    rows = list(reader)

total = len(rows)
n_chunks = math.ceil(total / chunk) if total else 0
print(f"Разбиваем {total} строк на {n_chunks} чанков по ~{chunk}")

for i in range(n_chunks):
    chunk_rows = rows[i * chunk:(i + 1) * chunk]
    chunk_file = os.path.join(out_dir, f"chunk_{i + 1:04d}.csv")
    with open(chunk_file, "w", newline="", encoding="utf-8") as out:
        writer = csv.DictWriter(out, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(chunk_rows)
PYEOF
}

import_escape_sql_literal() {
  printf "%s" "$1" | sed "s/'/''/g"
}

import_fetch_token() {
  local api_url="$1"
  local email="$2"
  local password="$3"
  local resp tok
  resp=$(curl -s -X POST "${api_url}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${email}\",\"password\":\"${password}\"}")
  tok=$(echo "${resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    if d.get('status') == 'error':
        err = d.get('error', {})
        print(f\"{err.get('code', 'login_failed')}: {err.get('message', d)}\", file=sys.stderr)
        sys.exit(1)
    print(d['data']['auth']['access_token'])
except Exception as e:
    print(f'login_parse_error: {e}', file=sys.stderr)
    sys.exit(1)
" 2>&1) || {
    echo "[ERROR] Не удалось войти как ${email}: ${tok}" >&2
    return 1
  }
  IMPORT_TOKEN="${tok}"
  IMPORT_TOKEN_OBTAINED_AT="${SECONDS}"
}

import_ensure_token() {
  local api_url="$1"
  local email="$2"
  local password="$3"
  local elapsed=$((SECONDS - IMPORT_TOKEN_OBTAINED_AT))
  if [[ -z "${IMPORT_TOKEN:-}" ]] || [[ "${elapsed}" -ge "${IMPORT_TOKEN_TTL}" ]]; then
    import_fetch_token "${api_url}" "${email}" "${password}" || return 1
  fi
}

import_upload_chunks() {
  local api_url="$1"
  local email="$2"
  local password="$3"
  local tmp_dir="$4"
  local log_fn="${5:-echo}"

  IMPORT_TOKEN=""
  IMPORT_TOKEN_OBTAINED_AT=0
  import_fetch_token "${api_url}" "${email}" "${password}" || return 1
  "${log_fn}" "Токен получен (${#IMPORT_TOKEN} символов)"

  local chunk_files=("${tmp_dir}"/chunk_*.csv)
  local n_chunks=${#chunk_files[@]}
  "${log_fn}" "Чанков для загрузки: ${n_chunks}"

  IMPORT_TOTAL_IMPORTED=0
  IMPORT_TOTAL_SKIPPED=0
  IMPORT_TOTAL_ERRORS=0
  IMPORT_FAILED_CHUNKS=0

  local i chunk_file chunk_num resp job_id status status_resp summary
  local curl_err http_code raw_resp err_hint

  for i in "${!chunk_files[@]}"; do
    chunk_file="${chunk_files[$i]}"
    chunk_num=$((i + 1))

    import_ensure_token "${api_url}" "${email}" "${password}" || return 1
    "${log_fn}" "  Чанк ${chunk_num}/${n_chunks}..."

    resp=""
    job_id=""
    curl_err=$(mktemp)
    raw_resp=$(curl -s --max-time "${IMPORT_CURL_MAX_TIME}" \
      -X POST "${api_url}/api/v1/admin/import-jobs" \
      -H "Authorization: Bearer ${IMPORT_TOKEN}" \
      -F "file=@${chunk_file};type=text/csv" \
      -F "source=catalog_import" \
      -w $'\n__HTTP_CODE__:%{http_code}' \
      2>"${curl_err}") || true
    http_code=$(echo "${raw_resp}" | sed -n 's/^__HTTP_CODE__://p' | tail -1)
    resp=$(echo "${raw_resp}" | sed '/^__HTTP_CODE__:/d')

    job_id=$(echo "${resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('id', ''))
except Exception:
    print('')
" 2>/dev/null)

    if [[ -z "${job_id}" ]]; then
      if [[ -n "${resp}" ]]; then
        err_hint=$(echo "${resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    e = d.get('error', {})
    print(f\"{e.get('code','?')}: {e.get('message','?')}\")
except Exception:
    print('')
" 2>/dev/null)
      else
        err_hint=$(tr '\n' ' ' < "${curl_err}" | head -c 200)
      fi
      rm -f "${curl_err}"
      "${log_fn}" "    ✗ Чанк ${chunk_num} не загружен (HTTP ${http_code:-?}): ${err_hint:-${resp:0:300}}"
      IMPORT_FAILED_CHUNKS=1
      return 1
    fi
    rm -f "${curl_err}"

    status=$(echo "${resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    print(j.get('status', 'pending'))
except Exception:
    print('pending')
" 2>/dev/null)
    status_resp="${resp}"

    if [[ "${status}" != "completed" && "${status}" != "completed_with_errors" && "${status}" != "failed" ]]; then
      local attempt
      for attempt in $(seq 1 "${IMPORT_POLL_ATTEMPTS}"); do
        sleep "${IMPORT_POLL_INTERVAL}"
        import_ensure_token "${api_url}" "${email}" "${password}" || return 1
        status_resp=$(curl -s "${api_url}/api/v1/admin/import-jobs/${job_id}" \
          -H "Authorization: Bearer ${IMPORT_TOKEN}" 2>/dev/null)
        status=$(echo "${status_resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d['data']['job']['status'])
except Exception:
    print('unknown')
" 2>/dev/null)
        [[ "${status}" == "completed" || "${status}" == "completed_with_errors" || "${status}" == "failed" ]] && break
      done
    fi

    if [[ "${status}" == "failed" ]]; then
      summary=$(echo "${status_resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    s = j.get('summary', {}) or {}
    print(f\"status={j.get('status')} errors={s.get('error_rows', 0)}\")
except Exception as e:
    print(f'parse_error={e}')
" 2>/dev/null)
      "${log_fn}" "    ✗ Чанк ${chunk_num}: job failed — ${summary}"
      IMPORT_FAILED_CHUNKS=1
      return 1
    fi

    if [[ "${status}" != "completed" && "${status}" != "completed_with_errors" ]]; then
      "${log_fn}" "    ✗ Чанк ${chunk_num}: неожиданный статус «${status}» (job ${job_id})"
      IMPORT_FAILED_CHUNKS=1
      return 1
    fi

    summary=$(echo "${status_resp}" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    j = d.get('data', {}).get('job', d.get('data', {}))
    s = j.get('summary', {}) or {}
    imp  = s.get('imported_rows', 0)
    upd  = s.get('updated_rows', 0)
    skip = s.get('skipped_rows', 0)
    dup  = s.get('duplicate_in_catalog_rows', 0)
    err  = s.get('error_rows', 0)
    st   = j.get('status', '?')
    print(f'imported={imp} updated={upd} skipped={skip} duplicates={dup} errors={err} status={st}')
except Exception as e:
    print(f'parse_error={e}')
" 2>/dev/null)

    "${log_fn}" "    ✓ Job ${job_id}: ${summary}"

    local imp skip err
  imp=$(echo "${summary}" | grep -oE 'imported=[0-9]+' | cut -d= -f2 || echo 0)
  skip=$(echo "${summary}" | grep -oE 'skipped=[0-9]+' | cut -d= -f2 || echo 0)
  dup=$(echo "${summary}" | grep -oE 'duplicates=[0-9]+' | cut -d= -f2 || echo 0)
  err=$(echo "${summary}" | grep -oE 'errors=[0-9]+' | cut -d= -f2 || echo 0)
    IMPORT_TOTAL_IMPORTED=$((IMPORT_TOTAL_IMPORTED + ${imp:-0}))
    IMPORT_TOTAL_SKIPPED=$((IMPORT_TOTAL_SKIPPED + ${skip:-0}))
    IMPORT_TOTAL_ERRORS=$((IMPORT_TOTAL_ERRORS + ${err:-0}))

    if [[ "${imp:-0}" -eq 0 && "${err:-0}" -gt 0 && "${chunk_num}" -eq 1 ]]; then
      "${log_fn}" "    Примеры ошибок (первые 3):"
      curl -s "${api_url}/api/v1/admin/import-jobs/${job_id}/errors?limit=3" \
        -H "Authorization: Bearer ${IMPORT_TOKEN}" 2>/dev/null | python3 -c "
import sys, json
try:
    errs = json.load(sys.stdin).get('data', {}).get('errors', [])
    for e in errs:
        row = e.get('row_number', '?')
        field = e.get('field_name') or '-'
        code = e.get('error_code', '?')
        msg = e.get('message', '?')
        print(f'      row={row} field={field} [{code}] {msg}')
except Exception:
    pass
" 2>/dev/null || true
    fi
  done
}

import_print_summary() {
  local api_url="$1"
  local log_fn="${2:-echo}"
  "${log_fn}" "ГОТОВО!"
  echo "  Загружено товаров : ${IMPORT_TOTAL_IMPORTED:-0}"
  echo "  Пропущено         : ${IMPORT_TOTAL_SKIPPED:-0}"
  echo "  Ошибок в строках  : ${IMPORT_TOTAL_ERRORS:-0}"
  if [[ "${IMPORT_FAILED_CHUNKS:-0}" -gt 0 ]]; then
    echo "  ✗ Упавших чанков  : ${IMPORT_FAILED_CHUNKS}"
  fi
  local gifts_total
  gifts_total=$(curl -s "${api_url}/api/v1/catalog/gifts?limit=1" 2>/dev/null | \
    python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('data', {}).get('page', {}).get('total', '?'))
except Exception:
    print('?')
" 2>/dev/null || echo "?")
  echo ""
  echo "  Товаров в каталоге: ${gifts_total}"
  echo "  API              : ${api_url}"
}
