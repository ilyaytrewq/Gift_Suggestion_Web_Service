#!/usr/bin/env python3
"""
Нормализация сырых CSV/XLSX выгрузок в формат каталога.

Выходной формат:
  gift_id, title, description, category, price, currency, store_url, image_url, age_min, age_max

Поддерживаемые форматы входных файлов:
  - Выгрузки fetch_wb.py / fetch_aliexpress.py (колонки уже в нужном формате)
  - Myntra Products Catalog (Kaggle: shivamb/fashion-clothing-products-catalog)
  - Online Retail II (Kaggle: ramzanzdemir/online-retail-gift-products)
  - Любой CSV с автоопределением колонок

Использование:
  python normalize_to_catalog.py input.csv --output catalog.csv
  python normalize_to_catalog.py file1.csv file2.xlsx  # --output необязателен
"""
from __future__ import annotations

import argparse
import csv
import logging
import os
import sys
from pathlib import Path
from typing import Iterator

try:
    import yaml
except ImportError:
    sys.exit("Установи зависимости: pip install pyyaml")

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MAPPING_PATH = os.path.join(SCRIPT_DIR, "category_mapping.yaml")

REQUIRED_FIELDS = {"gift_id", "title", "price", "currency", "store_url"}

OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category",
    "price", "currency", "store_url", "image_url",
    "age_min", "age_max",
]

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}

# Известные форматы Kaggle-датасетов: (набор обязательных колонок) → preset
KNOWN_PRESETS: list[tuple[set[str], dict]] = [
    (
        {"ProductID", "ProductName", "Price (INR)"},
        {
            "id_col": "ProductID", "id_prefix": "MYN",
            "title_col": "ProductName",
            "description_col": "Description",
            "price_col": "Price (INR)", "currency": "INR",
            "category_col": None,  # нет колонки — категория выводится из title
            "category_text_col": "ProductName",
            "url_template": "https://www.myntra.com/{id}",
            "source": "myntra",
        },
    ),
    (
        {"StockCode", "Description", "Price"},
        {
            "id_col": "StockCode", "id_prefix": "RT",
            "title_col": "Description",
            "description_col": None,
            "price_col": "Price", "currency": "GBP",
            "category_col": None,
            "url_template": "https://retail-catalog.placeholder/{id}",
            "source": "aliexpress",
        },
    ),
    # AliExpress Furniture dataset (Kaggle: kanchana1990/e-commerce-furniture-dataset-2024)
    (
        {"productTitle", "price", "sold", "tagText"},
        {
            "id_col": "productTitle", "id_prefix": "AEF",
            "title_col": "productTitle",
            "description_col": None,
            "price_col": "price", "currency": "USD",
            "category_col": None,
            "category_text_col": "productTitle",
            "url_template": "https://www.aliexpress.com/wholesale?SearchText={id}",
            "source": "aliexpress",
        },
    ),
    # AliExpress Products txt (Kaggle: joseluismartnezgarca/aliexpress-products)
    (
        {"id", "title", "original_price", "price", "seller", "orders"},
        {
            "id_col": "id", "id_prefix": "AEP",
            "title_col": "title",
            "description_col": None,
            "price_col": "price", "currency": "INR",
            "category_col": None,
            "category_text_col": "title",
            "url_template": "https://www.aliexpress.com/item/{id}.html",
            "source": "aliexpress",
        },
    ),
]


def load_mapping(path: str) -> dict:
    with open(path, encoding="utf-8") as f:
        return yaml.safe_load(f)


def detect_source(gift_id: str) -> str:
    if gift_id.startswith("WB"):
        return "wildberries"
    if gift_id.startswith(("AE", "CN")):
        return "aliexpress"
    if gift_id.startswith("MYN"):
        return "myntra"
    return "aliexpress"


def map_category(raw_category: str, source: str, mapping: dict) -> str:
    default = mapping.get("default", "Хобби")
    source_map: dict = mapping.get(source, {})
    lowered = raw_category.lower()
    for keyword, internal in source_map.items():
        if keyword.lower() in lowered:
            return internal
    return default


def validate_row(row: dict) -> bool:
    for field in REQUIRED_FIELDS:
        if not row.get(field, "").strip():
            return False
    try:
        if float(row["price"]) < 0:
            return False
    except (ValueError, TypeError):
        return False
    if not row["store_url"].startswith("http"):
        return False
    return True


def normalize_row(row: dict, mapping: dict) -> dict:
    gift_id = row.get("gift_id", "").strip()
    source = detect_source(gift_id)
    raw_cat = row.get("category", "").strip()
    category = raw_cat if raw_cat in VALID_CATEGORIES else map_category(raw_cat, source, mapping)

    try:
        price_cents = int(float(row.get("price", "0").strip()))
    except ValueError:
        price_cents = 0

    return {
        "gift_id": gift_id,
        "title": row.get("title", "").strip(),
        "description": row.get("description", "").strip(),
        "category": category,
        "price": f"{price_cents}.00",
        "currency": (row.get("currency") or "RUB").strip(),
        "store_url": row.get("store_url", "").strip(),
        "image_url": (row.get("image_url") or "").strip(),
        "age_min": str(row.get("age_min") or "0").strip(),
        "age_max": str(row.get("age_max") or "99").strip(),
    }


def read_csv(path: str) -> Iterator[dict]:
    ext = Path(path).suffix.lower()
    # Файлы AliExpress без заголовка — полуточка-разделитель, 7 колонок
    if ext == ".txt":
        COLS = ["id", "title", "original_price", "price", "seller", "orders", "_empty"]
        with open(path, newline="", encoding="utf-8", errors="replace") as f:
            reader = csv.reader(f, delimiter=";")
            for row in reader:
                padded = list(row) + [""] * (len(COLS) - len(row))
                yield {k: v.strip().strip('"') for k, v in zip(COLS, padded)}
        return
    with open(path, newline="", encoding="utf-8-sig") as f:
        yield from csv.DictReader(f)


def read_xlsx(path: str) -> Iterator[dict]:
    try:
        import openpyxl
    except ImportError:
        sys.exit("Установи зависимости: pip install openpyxl")
    wb = openpyxl.load_workbook(path, read_only=True, data_only=True)
    ws = wb.active
    rows = ws.iter_rows(values_only=True)
    headers = [str(h).strip() if h is not None else "" for h in next(rows)]
    for row in rows:
        yield dict(zip(headers, (str(v).strip() if v is not None else "" for v in row)))
    wb.close()


def detect_preset(headers: set[str]) -> dict | None:
    for required_cols, preset in KNOWN_PRESETS:
        if required_cols.issubset(headers):
            return preset
    return None


def remap_row(raw: dict, preset: dict) -> dict:
    id_val = raw.get(preset["id_col"], "").strip()
    gift_id = f"{preset['id_prefix']}{id_val}"
    url = preset["url_template"].format(id=id_val)

    cat_col = preset.get("category_col")
    cat_text_col = preset.get("category_text_col")
    if cat_col:
        raw_cat = raw.get(cat_col, "").strip()
    elif cat_text_col:
        raw_cat = raw.get(cat_text_col, "").strip()
    else:
        raw_cat = ""

    desc_col = preset.get("description_col")
    description = raw.get(desc_col, "").strip() if desc_col else ""

    import re as _re
    raw_price = str(raw.get(preset["price_col"], "0") or "0")
    raw_price = _re.sub(r"[^\d.]", "", raw_price)
    try:
        price = float(raw_price) if raw_price else 0.0
    except ValueError:
        price = 0.0

    return {
        "gift_id": gift_id,
        "title": raw.get(preset["title_col"], "").strip(),
        "description": description,
        "category": raw_cat,
        "price": str(price),
        "currency": preset["currency"],
        "store_url": url,
        "image_url": "",
        "age_min": "0",
        "age_max": "99",
    }


def iter_file(path: str) -> Iterator[dict]:
    ext = Path(path).suffix.lower()
    if ext in {".xlsx", ".xls"}:
        return read_xlsx(path)
    return read_csv(path)


def default_output(inputs: list[str]) -> str:
    stem = Path(inputs[0]).stem
    parent = Path(inputs[0]).parent
    return str(parent / f"{stem}_normalized.csv")


def main() -> None:
    parser = argparse.ArgumentParser(description="Нормализация CSV/XLSX выгрузок в формат каталога")
    parser.add_argument("inputs", nargs="+", help="Входные CSV или XLSX файлы")
    parser.add_argument("--output", help="Путь к выходному CSV (по умолчанию: <input>_normalized.csv)")
    parser.add_argument("--mapping", default=MAPPING_PATH, help="Путь к category_mapping.yaml")
    args = parser.parse_args()

    output_path = args.output or default_output(args.inputs)
    mapping = load_mapping(args.mapping)

    seen_urls: set[str] = set()
    total_in = total_skipped = total_dedup = count = 0

    with open(output_path, "w", newline="", encoding="utf-8") as out_f:
        writer = csv.DictWriter(out_f, fieldnames=OUTPUT_FIELDS)
        writer.writeheader()

        for input_path in args.inputs:
            log.info("Обрабатываем %s", input_path)
            rows = list(iter_file(input_path))
            if not rows:
                log.warning("Файл пустой: %s", input_path)
                continue

            headers = set(rows[0].keys())
            preset = detect_preset(headers)
            if preset:
                log.info("  Формат определён: %s (prefix=%s)", input_path, preset["id_prefix"])
            else:
                log.info("  Формат: стандартный (gift_id/title/price/...)")

            for raw_row in rows:
                total_in += 1
                row = remap_row(raw_row, preset) if preset else raw_row

                if not validate_row(row):
                    total_skipped += 1
                    continue

                store_url = row.get("store_url", "").strip()
                if store_url in seen_urls:
                    total_dedup += 1
                    continue
                seen_urls.add(store_url)

                writer.writerow(normalize_row(row, mapping))
                count += 1

    log.info(
        "Итог: вход=%d, пропущено=%d, дубликаты=%d, записано=%d → %s",
        total_in, total_skipped, total_dedup, count, output_path,
    )


if __name__ == "__main__":
    main()
