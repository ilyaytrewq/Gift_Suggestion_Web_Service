#!/usr/bin/env python3
"""
Финальная сборка каталога из всех источников: WB + AliExpress + Kaggle.

Что делает:
  - Читает все CSV из указанных файлов/папки
  - Объединяет в один датасет
  - Дедуплицирует по store_url (или gift_id если нет URL)
  - Валидирует (цена > 0, непустые обязательные поля, URL начинается с https)
  - Нормализует цену к формату NNNN.00
  - Добавляет age_min/age_max если отсутствуют
  - Сохраняет итоговый CSV + Parquet для ML-обучения
  - Выводит статистику по категориям, источникам, ценам

Использование:
  python merge_catalog.py \
    --inputs services/ml/dataset/collected/ \
    --output services/ml/dataset/catalog_real.csv \
    --output-parquet services/ml/dataset/catalog_real.parquet

  python merge_catalog.py \
    --inputs services/ml/dataset/collected/wb_gifts.csv services/ml/dataset/collected/ali_gifts.csv \
    --output services/ml/dataset/catalog_real.csv
"""
from __future__ import annotations

import argparse
import csv
import glob
import logging
import os
import re
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

REQUIRED_FIELDS = ["gift_id", "title", "price", "currency", "store_url"]

OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category", "price", "currency",
    "store_url", "image_url", "age_min", "age_max",
    "brand", "rating", "reviews_count", "discount_pct", "source",
]

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}


def normalize_price(raw: str) -> str | None:
    clean = re.sub(r"[^\d.,]", "", str(raw)).replace(",", ".")
    parts = clean.split(".")
    if len(parts) > 2 or not parts[0]:
        return None
    try:
        whole = int(parts[0])
        if whole <= 0:
            return None
        return f"{whole}.00"
    except ValueError:
        return None


def validate_row(row: dict) -> str | None:
    """Вернуть причину отказа или None если строка валидна."""
    for field in REQUIRED_FIELDS:
        if not (row.get(field) or "").strip():
            return f"empty_{field}"

    price = normalize_price(row.get("price", ""))
    if price is None:
        return "invalid_price"

    url = (row.get("store_url") or "").strip()
    if url and not url.startswith(("http://", "https://")):
        return "invalid_url"

    return None


def normalize_row(row: dict) -> dict:
    out = {field: (row.get(field) or "").strip() for field in OUTPUT_FIELDS}

    # Нормализация цены
    out["price"] = normalize_price(row.get("price", "")) or "0.00"

    # Дефолтные значения
    if not out["age_min"]:
        out["age_min"] = "0"
    if not out["age_max"]:
        out["age_max"] = "99"
    if not out["currency"]:
        out["currency"] = "RUB"
    if not out["category"] or out["category"] not in VALID_CATEGORIES:
        out["category"] = "Хобби"
    if not out["description"]:
        out["description"] = out["title"]

    return out


def collect_input_files(inputs: list[str]) -> list[Path]:
    files = []
    for inp in inputs:
        p = Path(inp)
        if p.is_dir():
            files.extend(sorted(p.glob("*.csv")))
        elif p.is_file():
            files.append(p)
        else:
            # Glob pattern
            for f in sorted(glob.glob(inp)):
                files.append(Path(f))
    return files


def main() -> None:
    parser = argparse.ArgumentParser(description="Слияние и валидация каталога подарков")
    parser.add_argument("--inputs", nargs="+", required=True,
                        help="CSV файлы или папка с CSV")
    parser.add_argument("--output", default="services/ml/dataset/catalog_real.csv")
    parser.add_argument("--output-parquet", default=None,
                        help="Дополнительно сохранить в Parquet (нужен pandas+pyarrow)")
    parser.add_argument("--min-price", type=float, default=50,
                        help="Минимальная цена в нац. валюте (отфильтровать мусор)")
    parser.add_argument("--max-price", type=float, default=100000)
    args = parser.parse_args()

    input_files = collect_input_files(args.inputs)
    log.info("Input files: %d", len(input_files))
    for f in input_files:
        log.info("  %s", f)

    seen_urls: set[str] = set()
    seen_ids: set[str] = set()
    all_rows: list[dict] = []

    stats_skip: dict[str, int] = {}
    stats_source: dict[str, int] = {}
    stats_cat: dict[str, int] = {}

    for csv_file in input_files:
        file_rows = 0
        file_skipped = 0
        try:
            with open(csv_file, encoding="utf-8", errors="replace") as f:
                reader = csv.DictReader(f)
                for raw_row in reader:
                    reason = validate_row(raw_row)
                    if reason:
                        stats_skip[reason] = stats_skip.get(reason, 0) + 1
                        file_skipped += 1
                        continue

                    row = normalize_row(raw_row)

                    # Проверка диапазона цен
                    try:
                        price_val = float(row["price"])
                    except ValueError:
                        stats_skip["invalid_price"] = stats_skip.get("invalid_price", 0) + 1
                        file_skipped += 1
                        continue

                    if price_val < args.min_price or price_val > args.max_price:
                        stats_skip["price_out_of_range"] = stats_skip.get("price_out_of_range", 0) + 1
                        file_skipped += 1
                        continue

                    # Дедупликация
                    dedup_key = row["store_url"] if row["store_url"] else row["gift_id"]
                    if dedup_key in seen_urls:
                        stats_skip["duplicate"] = stats_skip.get("duplicate", 0) + 1
                        continue

                    seen_urls.add(dedup_key)
                    seen_ids.add(row["gift_id"])
                    all_rows.append(row)
                    file_rows += 1

                    src = row.get("source", "unknown")
                    stats_source[src] = stats_source.get(src, 0) + 1
                    cat = row.get("category", "unknown")
                    stats_cat[cat] = stats_cat.get(cat, 0) + 1

        except Exception as e:
            log.error("Error reading %s: %s", csv_file, e)
            continue

        log.info("  %s: %d rows (skipped: %d)", csv_file.name, file_rows, file_skipped)

    log.info("\n=== ИТОГО: %d валидных товаров ===", len(all_rows))
    log.info("\nПо источникам:")
    for src, cnt in sorted(stats_source.items(), key=lambda x: -x[1]):
        log.info("  %-35s %d", src, cnt)
    log.info("\nПо категориям:")
    for cat, cnt in sorted(stats_cat.items(), key=lambda x: -x[1]):
        log.info("  %-30s %d", cat, cnt)
    if stats_skip:
        log.info("\nПричины отфильтровки:")
        for reason, cnt in sorted(stats_skip.items(), key=lambda x: -x[1]):
            log.info("  %-30s %d", reason, cnt)

    # Сохранение CSV
    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
    with open(args.output, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(all_rows)
    log.info("\nCSV: %s", args.output)

    # Сохранение Parquet
    if args.output_parquet:
        try:
            import pandas as pd
            df = pd.DataFrame(all_rows)
            df.to_parquet(args.output_parquet, index=False)
            log.info("Parquet: %s (%d rows)", args.output_parquet, len(df))
        except ImportError:
            log.warning("pandas/pyarrow не установлены — Parquet не сохранён")


if __name__ == "__main__":
    main()
