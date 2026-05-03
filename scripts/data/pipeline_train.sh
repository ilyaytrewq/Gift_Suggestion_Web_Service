#!/usr/bin/env bash
# Полный пайплайн: сбор датасетов → merge → prepare → train
#
# Использование:
#   bash scripts/data/pipeline_train.sh                 # всё
#   bash scripts/data/pipeline_train.sh --skip-fetch    # только merge+train (данные уже есть)
#   bash scripts/data/pipeline_train.sh --only amazon_2023 etsy sephora
#   bash scripts/data/pipeline_train.sh --max-rows 100000
#   bash scripts/data/pipeline_train.sh --queries 3000  # больше синтетических запросов

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

COLLECTED="${REPO_ROOT}/services/ml/dataset/collected"
CATALOG="${REPO_ROOT}/services/ml/dataset/catalog_real.csv"
TRAINING_PARQUET="${REPO_ROOT}/services/ml/dataset/training.parquet"
MODELS_DIR="${REPO_ROOT}/services/ml/models"

SKIP_FETCH=false
ONLY_ARGS=""
MAX_ROWS=50000
QUERIES=3000
VERSION="v2_0_0"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-fetch)   SKIP_FETCH=true; shift ;;
    --only)         shift; ONLY_ARGS="$*"; break ;;
    --max-rows)     MAX_ROWS="$2"; shift 2 ;;
    --queries)      QUERIES="$2"; shift 2 ;;
    --version)      VERSION="$2"; shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

mkdir -p "${COLLECTED}" "${MODELS_DIR}"

echo "=========================================="
echo " Gift ML Training Pipeline"
echo "  Collected dir : ${COLLECTED}"
echo "  Catalog       : ${CATALOG}"
echo "  Training data : ${TRAINING_PARQUET}"
echo "  Model version : ${VERSION}"
echo "=========================================="

# ── Шаг 1: Загрузка датасетов ──────────────────────────────────────────────
if [ "$SKIP_FETCH" = false ]; then
  echo ""
  echo ">>> [1/4] Загрузка датасетов (Kaggle)..."

  ONLY_FLAG=""
  if [ -n "${ONLY_ARGS}" ]; then
    ONLY_FLAG="--only ${ONLY_ARGS}"
  fi

  python3 "${SCRIPT_DIR}/fetch_datasets_2024.py" \
    --output "${COLLECTED}" \
    --max-rows "${MAX_ROWS}" \
    ${ONLY_FLAG}

  echo "Датасеты загружены → ${COLLECTED}"
else
  echo ""
  echo ">>> [1/4] Пропускаем загрузку (--skip-fetch)"
fi

# ── Шаг 2: Слияние в единый каталог ────────────────────────────────────────
echo ""
echo ">>> [2/4] Слияние и валидация каталога..."

python3 "${SCRIPT_DIR}/merge_catalog.py" \
  --inputs "${COLLECTED}/" \
  --output "${CATALOG}" \
  --min-price 100 \
  --max-price 500000

ROWS=$(wc -l < "${CATALOG}")
echo "Каталог: $((ROWS - 1)) товаров → ${CATALOG}"

# ── Шаг 3: Подготовка обучающего датасета ──────────────────────────────────
echo ""
echo ">>> [3/4] Генерация синтетических запросов (${QUERIES} шт.)..."

cd "${REPO_ROOT}/services/ml"
python3 -m training.prepare_dataset \
  --catalog ../../"${CATALOG#${REPO_ROOT}/}" \
  --queries "${QUERIES}" \
  --output dataset/training.parquet

echo "Training parquet: dataset/training.parquet"

# ── Шаг 4: Обучение модели ─────────────────────────────────────────────────
echo ""
echo ">>> [4/4] Обучение LightGBM LambdaRank..."

MODEL_PKL="models/lightgbm_${VERSION}.pkl"
METRICS_JSON="models/metrics_${VERSION}.json"

python3 -m training.train_lightgbm \
  --dataset dataset/training.parquet \
  --out "${MODEL_PKL}" \
  --metrics "${METRICS_JSON}"

echo ""
echo "=========================================="
echo " Готово!"
echo "  Модель:  services/ml/${MODEL_PKL}"
echo "  Метрики: services/ml/${METRICS_JSON}"
echo ""
echo " Запустить сервис с новой моделью:"
echo "   ML_MODEL_PATH=${MODEL_PKL} python3 -m ml_service.main"
echo ""
echo " Или в docker-compose.yml:"
echo "   ML_MODEL_PATH: ${MODEL_PKL}"
echo "=========================================="
