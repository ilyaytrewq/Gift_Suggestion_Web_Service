#!/usr/bin/env bash
# Полный сбор каталога подарков: WB (RU) + AliExpress (CN)
# Запускать с российского IP или через VPN RU (WB блокирует зарубежные IP)
#
# Использование:
#   bash scripts/data/collect_all.sh
#   bash scripts/data/collect_all.sh --pages 10 --fast

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COLLECTED="${REPO_ROOT}/services/ml/dataset/collected"
OUTPUT_CSV="${REPO_ROOT}/services/ml/dataset/catalog_real.csv"
OUTPUT_PARQUET="${REPO_ROOT}/services/ml/dataset/catalog_real.parquet"

PAGES=5
DELAY_WB=2.5
DELAY_AE=1.5
WB_QUERIES=35
AE_QUERIES=33
FETCH_DESCRIPTIONS=false

# Разбор аргументов
while [[ $# -gt 0 ]]; do
  case "$1" in
    --pages)     PAGES="$2";      shift 2 ;;
    --fast)      PAGES=2; DELAY_WB=1.5; DELAY_AE=1.0; WB_QUERIES=10; AE_QUERIES=10; shift ;;
    --descriptions) FETCH_DESCRIPTIONS=true; shift ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

mkdir -p "${COLLECTED}"

echo "==============================="
echo " Сбор каталога подарков"
echo "  WB:  ${WB_QUERIES} запросов × ${PAGES} страниц"
echo "  AE:  ${AE_QUERIES} запросов × ${PAGES} страниц"
echo "  Out: ${OUTPUT_CSV}"
echo "==============================="

# ── Шаг 1: Wildberries (RU) ─────────────────────────────────────────────────
echo ""
echo ">>> [1/3] Wildberries (RU)..."

DESCS_FLAG=""
if [ "$FETCH_DESCRIPTIONS" = true ]; then
  DESCS_FLAG="--descriptions"
fi

python3 "${SCRIPT_DIR}/fetch_wb_full.py" \
  --output "${COLLECTED}/wb_gifts.csv" \
  --pages "${PAGES}" \
  --delay "${DELAY_WB}" \
  --queries "${WB_QUERIES}" \
  ${DESCS_FLAG}

echo "WB done → ${COLLECTED}/wb_gifts.csv"

# ── Шаг 2: AliExpress (CN) ──────────────────────────────────────────────────
echo ""
echo ">>> [2/3] AliExpress (CN)..."

if [ -n "${ALIEXPRESS_APP_KEY:-}" ] && [ -n "${ALIEXPRESS_APP_SECRET:-}" ]; then
  echo "  Режим: Affiliate API (ключи найдены)"
  AE_MODE="affiliate"
else
  echo "  Режим: web scraping (без ключей — результат может быть неполным)"
  echo "  Для полного сбора: export ALIEXPRESS_APP_KEY=... ALIEXPRESS_APP_SECRET=..."
  AE_MODE="web"
fi

python3 "${SCRIPT_DIR}/fetch_aliexpress_full.py" \
  --mode "${AE_MODE}" \
  --output "${COLLECTED}/ali_gifts.csv" \
  --pages "${PAGES}" \
  --delay "${DELAY_AE}" \
  --queries "${AE_QUERIES}"

echo "AliExpress done → ${COLLECTED}/ali_gifts.csv"

# ── Шаг 3: Слияние ──────────────────────────────────────────────────────────
echo ""
echo ">>> [3/3] Слияние и валидация..."

python3 "${SCRIPT_DIR}/merge_catalog.py" \
  --inputs "${COLLECTED}/" \
  --output "${OUTPUT_CSV}" \
  --output-parquet "${OUTPUT_PARQUET}"

echo ""
echo "==============================="
echo " Готово!"
echo "  CSV:     ${OUTPUT_CSV}"
echo "  Parquet: ${OUTPUT_PARQUET}"
echo ""
echo " Следующий шаг — загрузить в БД:"
echo "   curl -X POST http://localhost:8080/api/v1/admin/import-jobs \\"
echo "     -H 'Authorization: Bearer <token>' \\"
echo "     -F 'file=@${OUTPUT_CSV}' -F 'format=csv'"
echo ""
echo " Или обучить модель:"
echo "   cd services/ml"
echo "   python3 -m training.prepare_dataset --catalog dataset/catalog_real.csv --queries 2000 --output dataset/training.parquet"
echo "   python3 -m training.train_lightgbm --dataset dataset/training.parquet --out models/lightgbm_v1_0_0.pkl"
echo "==============================="
