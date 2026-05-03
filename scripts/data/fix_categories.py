#!/usr/bin/env python3
"""
Исправление категорий в catalog_real.csv.

Принцип — консервативный:
  1. Source-based жёсткие маппинги (книги = Книги, etc.)
  2. Только для "Еда и напитки": перепроверить через улучшенный detect_category
     (word-boundary для "tea"/"food", kitchen-appliances → Электроника).
  3. Все остальные категории НЕ трогаются — чтобы не ломать правильную разметку.

Использование:
  python3 scripts/data/fix_categories.py --dry-run   # статистика
  python3 scripts/data/fix_categories.py             # обновить CSV
  python3 scripts/data/fix_categories.py --sql fix.sql  # + SQL для БД
"""
from __future__ import annotations

import argparse
import csv
import re
import sys
from collections import Counter
from pathlib import Path

REPO_ROOT = Path(__file__).parent.parent.parent
CATALOG_PATH = REPO_ROOT / "services/ml/dataset/catalog_real.csv"

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}

# Источник → принудительная категория (не пересчитывать).
SOURCE_OVERRIDES: dict[str, str] = {
    "books_google_2025": "Книги",
    "sephora": "Косметика",
    "beauty_2024": "Косметика",
    "amazon_laptops_2024": "Электроника",
    "headphones_2026": "Электроника",
    "marketplace_electronics_2026": "Электроника",
    "amazon_electronics_2025": "Электроника",
    "furniture_aliexpress_2024": "Украшения для дома",
    "myntra": "Одежда",
    "nike_2026": "Одежда",
    "adidas_2026": "Одежда",
    "etsy_wedding_2024": "Украшения для дома",
}

# (keywords, category, use_word_boundary)
# Порядок важен: первое совпадение выигрывает.
CATEGORY_KEYWORDS: list[tuple[list[str], str, bool]] = [
    # Кухонная техника — до "food"/"tea", иначе ложные матчи
    (["steam iron", "electric iron", "garment steamer", "garment iron",
      "electric kettle", "milk frother", "food processor", "food weighing",
      "mixer grinder", "hand blender", "immersion rod", "egg boiler",
      "rice cooker", "induction cooker", "air cooler", "air conditioner",
      "washing machine", "vacuum cleaner", "water heater",
      "coffee grinder", "coffee maker", "coffee machine", "espresso machine",
      "juicer mixer", "electric grinder", "electric blender", "food chopper",
      "meat grinder", "portable blender", "kitchen machine",
      "sandwich maker", "toaster", "oven toaster"], "Электроника", False),
    (["electronics", "phone", "laptop", "tablet", "camera", "headphone",
      "speaker", "smartwatch", "kindle", "gaming", "computer", "wireless",
      "bluetooth", "charger", "battery", "электроника", "телефон", "ноутбук",
      "наушник", "колонк", "планшет", "смартфон", "antenna", "radio",
      "receiver", "transmitter", "amplifier"], "Электроника", False),
    (["book", "novel", "fiction", "nonfiction", "biography", "manga",
      "textbook", "cookbook", "poetry", "comics", "graphic novel",
      "книг", "литератур", "роман"], "Книги", False),
    (["toy", "board game", "card game", "chess", "puzzle", "lego",
      "doll", "action figure", "playset", "игра", "игрушк", "пазл",
      "настольн", "monopoly", "scrabble"], "Настольные игры", False),
    (["beauty", "skincare", "makeup", "cosmetic", "perfume", "fragrance",
      "lipstick", "foundation", "mascara", "serum", "moisturizer", "blush",
      "eyeshadow", "sunscreen", "косметик", "парфюм", "крем", "уход"], "Косметика", False),
    (["bag", "wallet", "purse", "jewelry", "necklace", "bracelet",
      "earring", "ring", "sunglasses", "belt", "luggage", "handbag",
      "backpack", "сумк", "кошел", "украшен", "аксессуар", "ювелир",
      "часы", "watch", "hat", "scarf", "gloves"], "Аксессуары", False),
    (["clothing", "shirt", "dress", "jacket", "sweater", "hoodie",
      "jeans", "pants", "skirt", "coat", "fashion", "apparel", "blouse",
      "pajama", "pyjama", "underwear", "legging", "swimsuit",
      "одежд", "платье", "рубашк", "куртк", "футболк"], "Одежда", False),
    (["home decor", "candle", "vase", "picture frame", "wall art",
      "throw pillow", "blanket", "lamp", "plant pot", "decoration",
      "ornament", "figurine", "tableware", "mug", "kitchenware",
      "furniture", "sofa", "couch", "wardrobe", "bookshelf",
      "patio", "outdoor furniture", "dining table", "side table",
      "декор", "свеч", "ваз", "домашн", "интерьер", "посуд",
      "мебел"], "Украшения для дома", False),
    # "tea" и "food" — word-boundary; исключения через _FOOD_EXCLUSIONS
    (["snack", "chocolate", "candy", "wine", "gourmet", "organic food",
      "nuts", "cookie", "spice", "jam", "honey", "beverage", "alcohol",
      "whiskey", "bourbon", "beer", "cider", "kombucha", "protein bar",
      "еда", "напитк", "чай", "кофе", "конфет", "шоколад", "вино",
      "tea", "food", "coffee"], "Еда и напитки", True),
    (["baby", "infant", "toddler", "kids", "children", "nursery",
      "детск", "малыш", "ребён", "новорождённ", "stroller", "diaper"], "Детские товары", False),
    (["art", "craft", "hobby", "music instrument", "guitar", "paint",
      "drawing", "knitting", "sewing", "fitness", "yoga", "sport",
      "outdoor", "camping", "garden", "pet", "хобби", "творчеств",
      "рукодел", "спорт", "гитар", "рыбалк", "hiking"], "Хобби", False),
]

_FOOD_EXCLUSIONS = re.compile(
    r"food\s+grade|food\s+processor|food\s+chopper|food\s+weighing|food\s+scale"
    r"|food\s+service|food\s+work\s+shoe|food\s+storage|kitchen\s+shoe|chef\s+shoe|chef\s+clog"
    r"|steamer|steam\s+iron|kettle|garment"
    r"|coffee\s+table|tea\s+table|teapoy|tea\s+cart|coffee\s+cart",
    re.IGNORECASE,
)

# Цвета/характеристики материалов где "chocolate" = цвет, а не продукт
_CHOCOLATE_COLOR = re.compile(
    r"chocolate\s+(chip|brown|caramel|color|colored|leather|suede|lace|boot|shoe"
    r"|sneaker|upper|finish|tan)\b"
    r"|\bchip\b.*\bboot\b|\bchip\b.*\bshoe\b",
    re.IGNORECASE,
)

# "Jam" как название игры/бренда (не варенье)
_JAM_BRAND = re.compile(
    r"\b(nba|monster|space|def|bang|slam)\s+jam\b|\bjam\s+(shorts|short|pant|boxer)\b",
    re.IGNORECASE,
)

# Ключевые слова для косметики — матчат раньше "tea" (tea tree oil → Косметика)
_COSMETIC_BEFORE_FOOD = re.compile(
    r"\btea\s+tree\b|\bessential\s+oil\b|\baroma\b|\btherapy\s+oil\b",
    re.IGNORECASE,
)

# Цвета, содержащие слова food-категории (candy, wine, coffee — как цвет)
_COLOR_PATTERNS = re.compile(
    r"\bcandy\s*[/\\]|\bwine\s+(red|purple|color|colored)\b"
    r"|\bcoffee\s+(color|colored|brown)\b|\bcandy\s+(blue|pink|red|green|yellow|orange|purple)\b",
    re.IGNORECASE,
)


def _matches(low: str, keywords: list[str], word_boundary: bool) -> bool:
    for kw in keywords:
        if " " in kw:
            if kw in low:
                return True
        elif word_boundary:
            if re.search(r"\b" + re.escape(kw) + r"\b", low):
                if kw in ("tea", "food", "coffee", "candy", "wine", "chocolate", "jam"):
                    if _FOOD_EXCLUSIONS.search(low):
                        continue
                    if kw in ("candy", "wine") and _COLOR_PATTERNS.search(low):
                        continue
                    if kw == "chocolate" and _CHOCOLATE_COLOR.search(low):
                        continue
                    if kw == "jam" and _JAM_BRAND.search(low):
                        continue
                return True
        else:
            if kw in low:
                return True
    return False


def detect_category(title: str, description: str = "") -> str:
    """Определить категорию по заголовку (и опционально описанию)."""
    low = (title + " " + description).lower()
    # "tea tree" → Косметика раньше, чем "tea" → Еда и напитки
    if _COSMETIC_BEFORE_FOOD.search(low):
        return "Косметика"
    for keywords, cat, use_wb in CATEGORY_KEYWORDS:
        if _matches(low, keywords, use_wb):
            return cat
    return "Хобби"


def correct_category(row: dict) -> str:
    """
    Вернуть исправленную категорию для строки.

    Логика (консервативная — не трогаем то, что уже верно):
    1. Если текущая категория НЕ "Еда и напитки" → не меняем ничего.
    2. Если "Еда и напитки":
       a. Source override имеет приоритет (books → Книги, furniture → Украшения и т.д.)
       b. Иначе — detect_category по заголовку (с word-boundary для tea/food).
    """
    current = row.get("category", "Хобби")

    # 1. Не трогаем правильно категоризированные товары
    if current != "Еда и напитки":
        return current

    # 2. Пересчитываем только "Еда и напитки"
    source = row.get("source", "")

    # 2a. Source override для надёжных источников
    if source in SOURCE_OVERRIDES:
        return SOURCE_OVERRIDES[source]

    # 2b. Keyword-based с title (не description — там может быть мусор)
    title = row.get("title", "")
    return detect_category(title)


def main() -> None:
    parser = argparse.ArgumentParser(description="Исправление категорий в catalog_real.csv")
    parser.add_argument("--catalog", default=str(CATALOG_PATH))
    parser.add_argument("--dry-run", action="store_true", help="Только статистика, без записи")
    parser.add_argument("--sql", metavar="PATH", help="Записать SQL UPDATE в файл")
    args = parser.parse_args()

    catalog_path = Path(args.catalog)
    if not catalog_path.exists():
        sys.exit(f"Файл не найден: {catalog_path}")

    rows: list[dict] = []
    with open(catalog_path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        fieldnames = list(reader.fieldnames or [])
        rows = list(reader)

    changed: list[tuple[str, str, str]] = []  # (store_url, old_cat, new_cat)
    change_counter: Counter[str] = Counter()

    for row in rows:
        old_cat = row["category"]
        new_cat = correct_category(row)
        if old_cat != new_cat:
            changed.append((row["store_url"], old_cat, new_cat))
            change_counter[f"{old_cat} → {new_cat}"] += 1
            row["category"] = new_cat

    print(f"Всего строк: {len(rows)}")
    print(f"Изменено категорий: {len(changed)}")
    print()
    print("Изменения по типу:")
    for transition, cnt in change_counter.most_common(30):
        print(f"  {cnt:>5}  {transition}")

    # Итоговое распределение
    final_counter: Counter[str] = Counter(r["category"] for r in rows)
    print()
    print("Итоговые категории:")
    for cat, cnt in sorted(final_counter.items(), key=lambda x: -x[1]):
        print(f"  {cnt:>6}  {cat}")

    if args.dry_run:
        print("\n[dry-run] Файл не изменён.")
        return

    with open(catalog_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)
    print(f"\nЗаписано: {catalog_path}")

    if args.sql:
        _write_sql(changed, Path(args.sql))
        print(f"SQL записан: {args.sql}")


def _write_sql(changed: list[tuple[str, str, str]], out_path: Path) -> None:
    """Генерирует SQL UPDATE для обновления категорий в живой БД."""
    by_new_cat: dict[str, list[str]] = {}
    for store_url, _old, new_cat in changed:
        by_new_cat.setdefault(new_cat, []).append(store_url)

    lines = [
        "-- Автоматически сгенерированный SQL для исправления категорий.",
        "-- Запуск: psql <conn_string> -f fix.sql",
        "",
        "DO $$",
        "DECLARE",
    ]
    for cat in by_new_cat:
        lines.append(f"  {_cat_var(cat)} UUID;")
    lines += ["BEGIN", ""]

    for cat in by_new_cat:
        var = _cat_var(cat)
        escaped = cat.replace("'", "''")
        lines.append(f"  SELECT id INTO {var} FROM categories WHERE name = '{escaped}';")
    lines.append("")

    for new_cat, urls in by_new_cat.items():
        var = _cat_var(new_cat)
        for i in range(0, len(urls), 500):
            chunk = urls[i:i + 500]
            quoted = ", ".join(f"'{u.replace(chr(39), chr(39)+chr(39))}'" for u in chunk)
            lines += [
                f"  -- → {new_cat} (пакет {i // 500 + 1}, {len(chunk)} шт)",
                f"  UPDATE gifts SET category_id = {var}",
                f"  WHERE store_link IN ({quoted});",
                "",
            ]

    lines += ["END $$;", ""]
    out_path.write_text("\n".join(lines), encoding="utf-8")


def _cat_var(cat: str) -> str:
    return "v_" + re.sub(r"[^a-z0-9]", "_", cat.lower())


if __name__ == "__main__":
    main()
