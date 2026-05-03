#!/usr/bin/env python3
"""
Загрузка готовых датасетов с Kaggle — резервный источник данных.

Kaggle содержит несколько готовых датасетов с товарами WB и AliExpress,
собранных сторонними исследователями.

Требования:
  pip install kaggle
  kaggle.json с API-токеном в ~/.kaggle/kaggle.json
  (Получить: https://www.kaggle.com/settings → API → Create New Token)

Использование:
  python fetch_kaggle_datasets.py --output services/ml/dataset/collected/
  python fetch_kaggle_datasets.py --list    # показать доступные датасеты
  python fetch_kaggle_datasets.py --dataset eliseygusev/wildberries-data --output /tmp/wb_kaggle.csv
"""
from __future__ import annotations

import argparse
import csv
import logging
import os
import re
import shutil
import subprocess
import sys
import zipfile
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

# Известные датасеты с подарочными товарами RU/CN
KNOWN_DATASETS = [
    {
        "id": "eliseygusev/wildberries-data",
        "source": "wildberries",
        "currency": "RUB",
        "description": "Товары WB с ценами, категориями, рейтингами (актуальный датасет)",
        "cols": {
            "id": ["id", "product_id", "ID"],
            "title": ["name", "Наименование", "product_name", "title"],
            "price": ["priceU", "sale_price", "price", "Цена"],
            "category": ["subjectName", "category", "Категория"],
            "brand": ["brand", "Бренд", "Brand"],
            "rating": ["reviewRating", "rating", "Рейтинг"],
            "reviews": ["feedbacks", "reviews_count", "Отзывы"],
            "image": ["image", "img_url", "photo_url"],
        },
    },
    {
        "id": "latifahhukma/aliexpress-bestsellers-products-dataset",
        "source": "aliexpress",
        "currency": "USD",
        "description": "Бестселлеры AliExpress с категориями и рейтингами",
        "cols": {
            "id": ["product_id", "id", "asin"],
            "title": ["product_title", "title", "name"],
            "price": ["price", "sale_price"],
            "category": ["category", "first_level_category"],
            "brand": ["brand", "seller"],
            "rating": ["rating", "evaluate_rate"],
            "reviews": ["sold", "orders", "reviews_count"],
            "image": ["image_url", "main_image"],
        },
    },
    {
        "id": "promptcloud/product-details-on-aliexpress",
        "source": "aliexpress",
        "currency": "USD",
        "description": "Детальные данные товаров AliExpress включая описания",
        "cols": {
            "id": ["product_id", "id"],
            "title": ["product_name", "title"],
            "price": ["discounted_price", "original_price", "price"],
            "category": ["breadcrumbs", "category"],
            "brand": ["brand"],
            "rating": ["average_rating", "rating"],
            "reviews": ["total_reviews", "reviews"],
            "image": ["image_url"],
            "description": ["description", "product_details"],
        },
    },
]

# Маппинг категорий для данных из Kaggle
KAGGLE_CATEGORY_MAP: dict[str, str] = {
    "электроника": "Электроника",
    "phones": "Электроника",
    "electronics": "Электроника",
    "audio": "Электроника",
    "beauty": "Косметика",
    "косметика": "Косметика",
    "книги": "Книги",
    "books": "Книги",
    "игры": "Настольные игры",
    "games": "Настольные игры",
    "puzzle": "Настольные игры",
    "одежда": "Одежда",
    "clothing": "Одежда",
    "fashion": "Одежда",
    "аксессуары": "Аксессуары",
    "accessories": "Аксессуары",
    "jewelry": "Аксессуары",
    "дом": "Украшения для дома",
    "home": "Украшения для дома",
    "decor": "Украшения для дома",
    "еда": "Еда и напитки",
    "food": "Еда и напитки",
    "tea": "Еда и напитки",
    "детск": "Детские товары",
    "toy": "Детские товары",
    "kids": "Детские товары",
    "sport": "Хобби",
    "craft": "Хобби",
    "хобби": "Хобби",
}


def map_category(raw: str) -> str:
    low = (raw or "").lower()
    for key, cat in KAGGLE_CATEGORY_MAP.items():
        if key in low:
            return cat
    return "Хобби"


def find_col(row: dict, candidates: list[str]) -> str:
    for col in candidates:
        val = row.get(col)
        if val and str(val).strip():
            return str(val).strip()
    return ""


def convert_price(price_str: str, currency: str) -> tuple[str, str]:
    clean = re.sub(r"[^\d.,]", "", price_str).replace(",", ".")
    try:
        val = float(clean)
    except ValueError:
        return "0.00", currency

    if currency == "USD":
        # Конвертируем USD → CNY (примерный курс 7.2)
        val = round(val * 7.2, 0)
        return f"{int(val)}.00", "CNY"
    elif currency == "RUB":
        return f"{int(val)}.00", "RUB"
    return f"{int(val)}.00", currency


def download_dataset(dataset_id: str, output_dir: str) -> list[Path]:
    """Скачать датасет через kaggle CLI."""
    tmp = Path(output_dir) / "_kaggle_tmp"
    tmp.mkdir(parents=True, exist_ok=True)

    cmd = ["kaggle", "datasets", "download", "-d", dataset_id, "-p", str(tmp), "--unzip"]
    log.info("Downloading: %s", " ".join(cmd))
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode != 0:
        log.error("kaggle CLI failed:\n%s", result.stderr)
        raise RuntimeError(f"kaggle download failed: {result.stderr}")

    csvs = list(tmp.glob("**/*.csv"))
    log.info("Downloaded: %d CSV files", len(csvs))
    return csvs


def process_csv(
    csv_path: Path,
    dataset_info: dict,
    max_rows: int = 10000,
) -> list[dict]:
    cols = dataset_info["cols"]
    source = dataset_info["source"]
    currency = dataset_info["currency"]

    rows = []
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        for i, row in enumerate(reader):
            if i >= max_rows:
                break

            pid = find_col(row, cols.get("id", []))
            title = find_col(row, cols.get("title", []))
            if not pid or not title:
                continue

            price_raw = find_col(row, cols.get("price", []))
            price, cur = convert_price(price_raw, currency)
            if price == "0.00":
                continue

            cat_raw = find_col(row, cols.get("category", []))
            category = map_category(cat_raw)

            description = find_col(row, cols.get("description", []))
            if not description:
                description = title

            image = find_col(row, cols.get("image", []))
            brand = find_col(row, cols.get("brand", []))

            rating_raw = find_col(row, cols.get("rating", []))
            try:
                rating = round(float(re.sub(r"[^\d.]", "", rating_raw) or "0"), 1)
            except ValueError:
                rating = 0.0

            reviews_raw = find_col(row, cols.get("reviews", []))
            reviews = re.sub(r"[^\d]", "", reviews_raw) or "0"

            prefix = "WB" if source == "wildberries" else "AE"
            cur_final = "RUB" if source == "wildberries" else "CNY"

            rows.append({
                "gift_id": f"{prefix}_KGL_{pid}",
                "title": title[:200],
                "description": description[:500],
                "category": category,
                "price": price,
                "currency": cur_final,
                "store_url": "",
                "image_url": image,
                "age_min": "0",
                "age_max": "99",
                "brand": brand,
                "rating": str(rating),
                "reviews_count": reviews,
                "discount_pct": "0",
                "source": f"kaggle_{source}",
            })

    return rows


OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category", "price", "currency",
    "store_url", "image_url", "age_min", "age_max",
    "brand", "rating", "reviews_count", "discount_pct", "source",
]


def main() -> None:
    parser = argparse.ArgumentParser(description="Загрузка датасетов с Kaggle")
    parser.add_argument("--output", default="services/ml/dataset/collected/",
                        help="Папка для сохранения CSV файлов")
    parser.add_argument("--dataset", default=None,
                        help="Конкретный dataset ID (slug/name). По умолчанию — все из списка")
    parser.add_argument("--max-rows", type=int, default=10000, help="Максимум строк на датасет")
    parser.add_argument("--list", action="store_true", help="Показать список датасетов и выйти")
    args = parser.parse_args()

    if args.list:
        print("\nДоступные датасеты:\n")
        for d in KNOWN_DATASETS:
            print(f"  {d['id']}")
            print(f"    Source: {d['source']}, Currency: {d['currency']}")
            print(f"    Description: {d['description']}\n")
        return

    try:
        result = subprocess.run(["kaggle", "--version"], capture_output=True, text=True)
        log.info("kaggle CLI: %s", result.stdout.strip())
    except FileNotFoundError:
        print("kaggle CLI не найден. Установите: pip install kaggle")
        print("Затем создайте ~/.kaggle/kaggle.json с API токеном:")
        print("  https://www.kaggle.com/settings → API → Create New Token")
        sys.exit(1)

    os.makedirs(args.output, exist_ok=True)

    if args.dataset:
        datasets = [d for d in KNOWN_DATASETS if d["id"] == args.dataset]
        if not datasets:
            # Создаём entry для произвольного датасета
            datasets = [{"id": args.dataset, "source": "kaggle", "currency": "RUB",
                         "cols": {"id": ["id"], "title": ["name", "title"], "price": ["price"],
                                  "category": ["category"]}}]
    else:
        datasets = KNOWN_DATASETS

    for dataset_info in datasets:
        did = dataset_info["id"]
        slug = did.replace("/", "_")
        out_path = os.path.join(args.output, f"kaggle_{slug}.csv")

        log.info("=== %s ===", did)
        try:
            csv_files = download_dataset(did, args.output)
        except Exception as e:
            log.error("Skipping %s: %s", did, e)
            continue

        all_rows: list[dict] = []
        for csv_file in csv_files:
            try:
                rows = process_csv(csv_file, dataset_info, args.max_rows)
                all_rows.extend(rows)
                log.info("  %s: %d rows", csv_file.name, len(rows))
            except Exception as e:
                log.warning("  Error reading %s: %s", csv_file.name, e)

        if not all_rows:
            log.warning("No rows extracted from %s", did)
            continue

        # Дедупликация по gift_id
        seen: set[str] = set()
        unique_rows = []
        for row in all_rows:
            if row["gift_id"] not in seen:
                seen.add(row["gift_id"])
                unique_rows.append(row)

        with open(out_path, "w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS, extrasaction="ignore")
            writer.writeheader()
            writer.writerows(unique_rows)

        log.info("Saved: %d rows → %s", len(unique_rows), out_path)


if __name__ == "__main__":
    main()
