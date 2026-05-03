#!/usr/bin/env python3
"""
Загрузка современных датасетов с подарочными товарами (2024-2026).

Источники: Amazon 2023-2024, Etsy, Sephora, Google Books, Flipkart,
           Amazon Electronics/Laptops, Beauty 2024, Wedding Gifts 2024 и др.

Требования:
  kaggle.json в ~/.kaggle/ (https://www.kaggle.com/settings → API → Create New Token)

Использование:
  python3 scripts/data/fetch_datasets_2024.py
  python3 scripts/data/fetch_datasets_2024.py --list
  python3 scripts/data/fetch_datasets_2024.py --only amazon_2023 sephora books_2025
  python3 scripts/data/fetch_datasets_2024.py --max-rows 100000
"""
from __future__ import annotations

import argparse
import csv
import html
import logging
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Iterator

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

SCRIPT_DIR = Path(__file__).parent.resolve()
REPO_ROOT = SCRIPT_DIR.parent.parent
DEFAULT_OUTPUT = str(REPO_ROOT / "services/ml/dataset/collected")

# ── Категории ─────────────────────────────────────────────────────────────────

VALID_CATEGORIES = {
    "Электроника", "Книги", "Настольные игры", "Косметика",
    "Аксессуары", "Одежда", "Хобби", "Украшения для дома",
    "Еда и напитки", "Детские товары",
}

# Ключевые слова, помеченные * используют word-boundary matching (\bkw\b),
# остальные — substring. Multi-word phrases всегда substring.
# Порядок важен: первое совпадение выигрывает.
CATEGORY_KEYWORDS: list[tuple[list[str], str, bool]] = [
    # Кухонная техника → Электроника (раньше, чем "food"/"tea" дадут ложный матч)
    (["steam iron", "electric iron", "garment steamer", "garment iron",
      "electric kettle", "milk frother", "food processor", "food weighing",
      "mixer grinder", "hand blender", "immersion rod", "egg boiler",
      "electric kettle", "rice cooker", "induction cooker", "air cooler",
      "air conditioner", "washing machine", "vacuum cleaner", "water heater",
      "coffee grinder", "coffee maker", "coffee machine", "espresso machine",
      "juicer mixer", "electric grinder", "electric blender", "food chopper",
      "meat grinder", "portable blender"], "Электроника", False),
    (["electronics", "phone", "laptop", "tablet", "camera", "headphone",
      "speaker", "smartwatch", "kindle", "gaming", "computer", "wireless",
      "bluetooth", "charger", "battery", "электроника", "телефон", "ноутбук",
      "наушник", "колонк", "планшет", "смартфон"], "Электроника", False),
    (["book", "novel", "fiction", "nonfiction", "biography", "manga",
      "reading", "textbook", "cookbook", "poetry", "comics", "graphic novel",
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
      "pajama", "pyjama", "underwear", "shorts", "legging", "swimsuit",
      "одежд", "платье", "рубашк", "куртк", "футболк"], "Одежда", False),
    (["home decor", "candle", "vase", "picture frame", "wall art",
      "throw pillow", "blanket", "lamp", "plant pot", "decoration",
      "ornament", "figurine", "tableware", "mug", "kitchenware",
      "furniture", "sofa", "couch", "wardrobe", "bookshelf",
      "patio", "outdoor furniture", "dining table",
      "декор", "свеч", "ваз", "домашн", "интерьер", "посуд",
      "мебел"], "Украшения для дома", False),
    # Еда и напитки: "tea" и "food" используют word-boundary, чтобы не матчить
    # "steamer", "team", "teal", "teana", "stealth", "food grade", "food processor"
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

# Составные фразы-исключения: если title содержит такой фрагмент,
# "tea" или "food" НЕ должны вызывать "Еда и напитки".
_FOOD_EXCLUSIONS = re.compile(
    r"food\s+grade|food\s+processor|food\s+chopper|food\s+weighing|food\s+scale"
    r"|food\s+service|food\s+work\s+shoe|kitchen\s+shoe|chef\s+shoe|chef\s+clog"
    r"|steamer|steam\s+iron|kettle|garment",
    re.IGNORECASE,
)

_COSMETIC_BEFORE_FOOD = re.compile(
    r"\btea\s+tree\b|\bessential\s+oil\b|\baroma\b|\btherapy\s+oil\b",
    re.IGNORECASE,
)

_COLOR_PATTERNS = re.compile(
    r"\bcandy\s*[/\\]|\bwine\s+(red|purple|color|colored)\b"
    r"|\bcoffee\s+(color|colored|brown)\b|\bcandy\s+(blue|pink|red|green|yellow|orange|purple)\b",
    re.IGNORECASE,
)

_CHOCOLATE_COLOR = re.compile(
    r"chocolate\s+(chip|brown|caramel|color|colored|leather|suede|lace|boot|shoe"
    r"|sneaker|upper|finish|tan)\b"
    r"|\bchip\b.*\bboot\b|\bchip\b.*\bshoe\b",
    re.IGNORECASE,
)

_JAM_BRAND = re.compile(
    r"\b(nba|monster|space|def|bang|slam)\s+jam\b|\bjam\s+(shorts|short|pant|boxer)\b",
    re.IGNORECASE,
)


def _matches_keywords(low: str, keywords: list[str], word_boundary: bool) -> bool:
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


def detect_category(text: str) -> str:
    low = text.lower()
    if _COSMETIC_BEFORE_FOOD.search(low):
        return "Косметика"
    for keywords, cat, use_wb in CATEGORY_KEYWORDS:
        if _matches_keywords(low, keywords, use_wb):
            return cat
    return "Хобби"


# ── Вспомогательные функции ───────────────────────────────────────────────────

def find_col(row: dict, candidates: list[str], default: str = "") -> str:
    for col in candidates:
        val = row.get(col)
        if val is not None and str(val).strip() not in ("", "None", "nan", "NaN", "null"):
            return str(val).strip()
    return default


def clean_price(raw: str) -> float:
    cleaned = re.sub(r"[^\d.,]", "", str(raw)).replace(",", ".")
    parts = cleaned.split(".")
    if len(parts) > 2:
        cleaned = parts[0]
    try:
        return max(0.0, float(cleaned))
    except ValueError:
        return 0.0


def strip_html(text: str) -> str:
    text = re.sub(r"<[^>]+>", " ", text)
    return html.unescape(text).strip()


OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category", "price", "currency",
    "store_url", "image_url", "age_min", "age_max",
    "brand", "rating", "reviews_count", "discount_pct", "source",
]


def make_row(
    gift_id: str, title: str, description: str, category: str,
    price: float, currency: str, store_url: str,
    image_url: str = "", brand: str = "",
    rating: float = 0.0, reviews_count: int = 0,
    age_min: int = 0, age_max: int = 99, source: str = "",
) -> dict:
    cat = category if category in VALID_CATEGORIES else detect_category(f"{title} {category}")
    return {
        "gift_id": gift_id[:80],
        "title": title[:250],
        "description": (description or title)[:600],
        "category": cat,
        "price": f"{int(price)}.00",
        "currency": currency,
        "store_url": store_url[:400],
        "image_url": image_url[:400],
        "age_min": str(age_min),
        "age_max": str(age_max),
        "brand": brand[:100],
        "rating": f"{rating:.1f}",
        "reviews_count": str(reviews_count),
        "discount_pct": "0",
        "source": source,
    }


# ── Процессоры ────────────────────────────────────────────────────────────────

def process_amazon_2023(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    asaniczka/amazon-products-dataset-2023-1-4m-products
    Cols: asin, title, imgUrl, productURL, stars, reviews, price, listPrice, category_id, isBestSeller
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["title"])
            asin = find_col(row, ["asin"])
            if not title or not asin:
                continue
            price = clean_price(find_col(row, ["price"])) or clean_price(find_col(row, ["listPrice"]))
            if price <= 0 or price > 2000:
                continue
            stars = 0.0
            try:
                stars = float(find_col(row, ["stars"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["reviews"], "0")) or "0")
            url = find_col(row, ["productURL"], f"https://www.amazon.com/dp/{asin}")
            cat_raw = find_col(row, ["category_id", "category"])
            yield make_row(
                gift_id=f"AMZ_{asin}", title=title, description=title,
                category=detect_category(f"{title} {cat_raw}"),
                price=price, currency="USD", store_url=url,
                image_url=find_col(row, ["imgUrl"]),
                rating=stars, reviews_count=reviews, source="amazon_2023",
            )
            count += 1


def process_sephora(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    nadyinky/sephora-products-and-skincare-reviews → product_info.csv
    Cols: product_id, product_name, brand_name, price_usd, sale_price_usd, rating, reviews, ...
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["product_name", "name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["sale_price_usd", "price_usd", "price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["product_id", "id"])
            brand = find_col(row, ["brand_name", "brand"])
            rating = 0.0
            try:
                rating = float(find_col(row, ["rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["reviews", "reviews_count"], "0")) or "0")
            yield make_row(
                gift_id=f"SEP_{pid or count}", title=title,
                description=f"{title} by {brand}".strip(" by "),
                category="Косметика", price=price, currency="USD",
                store_url=f"https://www.sephora.com/search?keyword={title[:40].replace(' ', '+')}",
                brand=brand, rating=rating, reviews_count=reviews, source="sephora",
            )
            count += 1


def process_beauty_2024(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    waqi786/most-used-beauty-cosmetics-products-in-the-world
    Cols: Product_Name, Brand, Category, Price_USD, Rating, Number_of_Reviews, Skin_Type, Gender_Target
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["Product_Name", "product_name", "name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["Price_USD", "price_usd", "price"]))
            if price <= 0:
                continue
            brand = find_col(row, ["Brand", "brand"])
            skin = find_col(row, ["Skin_Type"])
            gender = find_col(row, ["Gender_Target"])
            desc = title
            if skin:
                desc += f". Для типа кожи: {skin}"
            if gender:
                desc += f". Аудитория: {gender}"
            rating = 0.0
            try:
                rating = float(find_col(row, ["Rating", "rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["Number_of_Reviews", "reviews"], "0")) or "0")
            yield make_row(
                gift_id=f"BTY_{count}", title=title, description=desc,
                category="Косметика", price=price, currency="USD",
                store_url=f"https://www.amazon.com/s?k={title[:40].replace(' ', '+')}",
                brand=brand, rating=rating, reviews_count=reviews, source="beauty_2024",
            )
            count += 1


def process_wedding_gifts_etsy(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    kanchana1990/wedding-gift-items-dataset → etsy_wedding_gift_dataset.csv
    Cols: name, descriptionHTML, Price, listedOn, favorites
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["name", "title", "Name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["Price", "price", "PRICE"]))
            if price <= 0:
                continue
            desc_html = find_col(row, ["descriptionHTML", "description"])
            desc = strip_html(desc_html)[:400] if desc_html else title
            favorites = int(re.sub(r"[^\d]", "", find_col(row, ["favorites"], "0")) or "0")
            yield make_row(
                gift_id=f"ETSY_{count}", title=title, description=desc,
                category=detect_category(f"{title} wedding gift handmade"),
                price=price, currency="USD",
                store_url=f"https://www.etsy.com/search?q={title[:40].replace(' ', '+')}",
                reviews_count=favorites, source="etsy_wedding_2024",
            )
            count += 1


def process_books_google(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    mihikaajayjadhav/books-dataset-15k-books-across-100-categories → google_books_dataset.csv
    Cols: book_id, title, subtitle, authors, publisher, description, categories,
          average_rating, ratings_count, list_price, currency, thumbnail
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["title", "Title"])
            if not title:
                continue
            price = clean_price(find_col(row, ["list_price", "price"]))
            if price <= 0:
                price = 10.0 + (count % 40)
            bid = find_col(row, ["book_id", "isbn_13", "isbn_10", "id"])
            authors = find_col(row, ["authors", "author"])
            desc = find_col(row, ["description"], "")
            if not desc:
                desc = f"{title}"
                if authors:
                    desc += f" by {authors}"
            subtitle = find_col(row, ["subtitle"])
            if subtitle:
                title = f"{title}: {subtitle}"[:250]
            currency = find_col(row, ["currency"], "USD")
            rating = 0.0
            try:
                rating = float(find_col(row, ["average_rating", "rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["ratings_count", "reviews_count"], "0")) or "0")
            img = find_col(row, ["thumbnail", "image_url"])
            cat_raw = find_col(row, ["categories", "search_category", "category"])
            yield make_row(
                gift_id=f"BOOK_{bid or count}", title=title, description=desc[:500],
                category=detect_category(f"book {cat_raw} {title}"),
                price=price, currency=currency or "USD",
                store_url=f"https://books.google.com/books?id={bid}" if bid else f"https://www.amazon.com/s?k={title[:40].replace(' ', '+')}",
                image_url=img, brand=authors,
                rating=rating, reviews_count=reviews, source="books_google_2025",
            )
            count += 1


def process_amazon_laptops_2024(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    kanchana1990/amazons-500-bestsellers-in-laptop-gear-2024 → amazon_top500.csv
    Cols: title, brand, description, price/currency, price/value, stars, reviewsCount
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["title", "Title", "product_title"])
            if not title:
                continue
            price = clean_price(find_col(row, ["price/value", "price", "Price"]))
            if price <= 0:
                continue
            brand = find_col(row, ["brand", "Brand"])
            desc = find_col(row, ["description", "about_product"], title)
            currency = find_col(row, ["price/currency", "currency"], "USD")
            rating = 0.0
            try:
                rating = float(find_col(row, ["stars", "rating", "Rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["reviewsCount", "reviews_count", "reviews"], "0")) or "0")
            yield make_row(
                gift_id=f"AMZLP_{count}", title=title, description=desc,
                category="Электроника",
                price=price, currency=currency,
                store_url=f"https://www.amazon.com/s?k={title[:40].replace(' ', '+')}",
                brand=brand, rating=rating, reviews_count=reviews, source="amazon_laptops_2024",
            )
            count += 1


def process_amazon_sales_india(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    karkavelrajaj/amazon-sales-dataset → amazon.csv
    Cols: product_id, product_name, category, discounted_price, actual_price,
          discount_percentage, rating, rating_count, about_product, img_link, product_link
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["product_name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["discounted_price", "actual_price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["product_id"])
            desc = find_col(row, ["about_product"], title)
            cat_raw = find_col(row, ["category"])
            img = find_col(row, ["img_link"])
            store_url = find_col(row, ["product_link"], "")
            if store_url and not store_url.startswith("http"):
                store_url = ""
            rating = 0.0
            try:
                rating = float(find_col(row, ["rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["rating_count"], "0")) or "0")
            discount_str = find_col(row, ["discount_percentage"], "0")
            yield make_row(
                gift_id=f"AMZIN_{pid or count}", title=title, description=desc[:400],
                category=detect_category(f"{title} {cat_raw}"),
                price=price, currency="INR",
                store_url=store_url or f"https://www.amazon.in/s?k={title[:40].replace(' ', '+')}",
                image_url=img, rating=rating, reviews_count=reviews, source="amazon_india_2024",
            )
            count += 1


def process_flipkart(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    PromptCloudHQ/flipkart-products → flipkart_com-ecommerce_sample.csv
    Cols: pid, product_name, description, retail_price, discounted_price, product_rating,
          overall_rating, brand, image, product_url, product_category_tree
    """
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["product_name", "name", "title"])
            if not title:
                continue
            price = clean_price(find_col(row, ["discounted_price", "retail_price", "actual_price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["pid", "product_id", "uniq_id"])
            desc = strip_html(find_col(row, ["description", "product_specifications"], ""))[:400] or title
            img = find_col(row, ["image", "image_url"])
            brand = find_col(row, ["brand", "Brand"])
            cat_raw = find_col(row, ["product_category_tree", "category"])
            store_url = find_col(row, ["product_url", "url"], "")
            if store_url and not store_url.startswith("http"):
                store_url = ""
            rating = 0.0
            try:
                rating = float(find_col(row, ["product_rating", "overall_rating", "rating"]))
            except ValueError:
                pass
            yield make_row(
                gift_id=f"FLP_{pid or count}", title=title, description=desc,
                category=detect_category(f"{title} {cat_raw}"),
                price=price, currency="INR",
                store_url=store_url or f"https://www.flipkart.com/search?q={title[:40].replace(' ', '+')}",
                image_url=img, brand=brand, rating=rating, source="flipkart",
            )
            count += 1


def process_aliexpress(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """Универсальный процессор для AliExpress датасетов."""
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["product_title", "productTitle", "title", "name", "product_name"])
            if not title:
                continue
            price = clean_price(find_col(row, [
                "price", "discounted_price", "sale_price", "original_price", "originalPrice",
            ]))
            if price <= 0:
                continue
            pid = find_col(row, ["product_id", "id", "asin"])
            img = find_col(row, ["image_url", "main_image", "imgUrl"])
            cat_raw = find_col(row, ["category", "first_level_category", "tagText"])
            rating = 0.0
            try:
                rating = float(find_col(row, ["rating", "evaluate_rate", "ratings", "stars"]))
            except ValueError:
                pass
            sold = int(re.sub(r"[^\d]", "", find_col(row, ["sold", "orders", "total_sold"], "0")) or "0")
            yield make_row(
                gift_id=f"AE_{pid or count}", title=title, description=title,
                category=detect_category(f"{title} {cat_raw}"),
                price=price, currency="USD",
                store_url=f"https://www.aliexpress.com/wholesale?SearchText={title[:40].replace(' ', '+')}",
                image_url=img, rating=rating, reviews_count=sold, source="aliexpress",
            )
            count += 1


def process_furniture_2024(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    kanchana1990/e-commerce-furniture-dataset-2024
    Cols: productTitle, originalPrice, price, sold, tagText
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["productTitle", "product_title", "title", "name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["price", "Price", "originalPrice"]))
            if price <= 0:
                continue
            sold = int(re.sub(r"[^\d]", "", find_col(row, ["sold", "orders"], "0")) or "0")
            cat_raw = find_col(row, ["tagText", "tag", "category"])
            yield make_row(
                gift_id=f"FURNI_{count}", title=title, description=title,
                category=detect_category(f"{title} {cat_raw} home furniture"),
                price=price, currency="USD",
                store_url=f"https://www.aliexpress.com/wholesale?SearchText={title[:40].replace(' ', '+')}",
                reviews_count=sold, source="furniture_aliexpress_2024",
            )
            count += 1


def process_myntra(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    shivamb/fashion-clothing-products-catalog
    Cols: ProductID, ProductName, ProductBrand, Gender, Price (INR), NumImages, Description, PrimaryColor
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["ProductName", "product_name", "Name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["Price (INR)", "Price", "price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["ProductID", "product_id", "id"])
            brand = find_col(row, ["ProductBrand", "brand"])
            desc = find_col(row, ["Description", "description"], title)
            gender = find_col(row, ["Gender", "gender"])
            color = find_col(row, ["PrimaryColor", "color"])
            if gender or color:
                desc = f"{desc}. {gender} {color}".strip(". ")
            yield make_row(
                gift_id=f"MYN_{pid or count}", title=title, description=desc[:400],
                category=detect_category(f"{title} clothing fashion apparel"),
                price=price, currency="INR",
                store_url=f"https://www.myntra.com/{pid}" if pid else f"https://www.myntra.com/search?q={title[:30].replace(' ', '+')}",
                brand=brand, source="myntra",
            )
            count += 1


def process_amazon_electronics_2025(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    prothomeshmistry/amazon-electronics-and-accessories-2025
    Cols: Product_Name, Price, Rating, Review_Count, ASIN, Product_URL, Availability
    BOM в первой колонке.
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["Product_Name", "product_name", "title"])
            if not title:
                continue
            price = clean_price(find_col(row, ["Price", "price", "discounted_price"]))
            if price <= 0:
                continue
            asin = find_col(row, ["ASIN", "asin", "product_id"])
            url = find_col(row, ["Product_URL", "product_url", "url"], "")
            if url and not url.startswith("http"):
                url = ""
            rating = 0.0
            try:
                rating = float(find_col(row, ["Rating", "rating", "stars"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["Review_Count", "reviews", "rating_count"], "0")) or "0")
            yield make_row(
                gift_id=f"AMZE25_{asin or count}", title=title, description=title,
                category=detect_category(f"{title} electronics gadget"),
                price=price, currency="INR",
                store_url=url or f"https://www.amazon.in/dp/{asin}" if asin else f"https://www.amazon.in/s?k={title[:40].replace(' ', '+')}",
                rating=rating, reviews_count=reviews, source="amazon_electronics_2025",
            )
            count += 1


def process_nike_adidas_2026(csv_path: Path, max_rows: int, brand_name: str = "") -> Iterator[dict]:
    """
    bsthere/nike-global-catalogue-2026 и bsthere/adidas-global-catalogue-2026
    Cols: snapshot_date, country_code, product_name, currency, price_local,
          sale_price_local, gender_segment, category, subcategory, product_id, brand_name
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        seen: set[str] = set()
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["product_name", "name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["sale_price_local", "price_local", "price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["product_id", "sku", "model_number"])
            brand = find_col(row, ["brand_name", "brand"]) or brand_name
            currency = find_col(row, ["currency"], "USD")
            cat_raw = find_col(row, ["category", "subcategory"])
            gender = find_col(row, ["gender_segment", "gender"])
            # Дедупликация по product_id (один товар во многих размерах)
            dedup = f"{pid}_{currency}"
            if dedup in seen:
                continue
            seen.add(dedup)
            desc = title
            if gender:
                desc = f"{title}. {gender}"
            cat = detect_category(f"{title} {cat_raw} clothing shoes sport")
            yield make_row(
                gift_id=f"{brand_name.upper() or 'SPORT'}_{pid or count}_{currency}",
                title=f"{brand} {title}".strip(),
                description=desc,
                category=cat,
                price=price, currency=currency,
                store_url=f"https://www.{brand_name.lower()}.com/search?q={title[:40].replace(' ', '+')}",
                brand=brand, source=f"{brand_name.lower()}_2026",
            )
            count += 1


def process_marketplace_electronics_2026(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    asadullahcreative/e-commerce-microphone-marketplace-dataset (2026)
    Cols: title, price, sold_count, rating, review_count, location, seller_name,
          category, original_price, discount_percent
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, ["title", "product_name", "name"])
            if not title:
                continue
            price = clean_price(find_col(row, ["price", "Price", "sale_price", "original_price"]))
            if price <= 0:
                continue
            sold = int(re.sub(r"[^\d]", "", find_col(row, ["sold_count", "sold", "orders"], "0")) or "0")
            rating = 0.0
            try:
                rating = float(find_col(row, ["rating", "Rating", "stars"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["review_count", "reviews"], "0")) or "0")
            cat_raw = find_col(row, ["category", "Category"])
            yield make_row(
                gift_id=f"MKTELEC_{count}", title=title, description=title,
                category=detect_category(f"{title} {cat_raw} electronics"),
                price=price, currency="USD",
                store_url=f"https://www.aliexpress.com/wholesale?SearchText={title[:40].replace(' ', '+')}",
                rating=rating, reviews_count=max(sold, reviews), source="marketplace_electronics_2026",
            )
            count += 1


def process_headphones_2026(csv_path: Path, max_rows: int) -> Iterator[dict]:
    """
    maulikgajera/global-headphones-specifications-and-market-dataset (2026)
    Cols: ID, Brand, Model, Type, Connectivity, Price (USD), Release Year, Avg Rating, Review Count, ...
    """
    with open(csv_path, encoding="utf-8-sig", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            brand = find_col(row, ["Brand", "brand"])
            model = find_col(row, ["Model", "model"])
            if not model:
                continue
            title = f"{brand} {model}".strip() if brand else model
            price = clean_price(find_col(row, ["Price (USD)", "price", "Price"]))
            if price <= 0:
                continue
            pid = find_col(row, ["ID", "id"])
            connectivity = find_col(row, ["Connectivity"])
            hp_type = find_col(row, ["Type"])
            desc = f"{title}. {hp_type} {connectivity}".strip(". ")
            rating = 0.0
            try:
                rating = float(find_col(row, ["Avg Rating", "rating"]))
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, ["Review Count", "reviews"], "0")) or "0")
            yield make_row(
                gift_id=f"HP26_{pid or count}", title=title, description=desc,
                category="Электроника", price=price, currency="USD",
                store_url=f"https://www.amazon.com/s?k={title[:40].replace(' ', '+')}",
                brand=brand, rating=rating, reviews_count=reviews, source="headphones_2026",
            )
            count += 1


def process_generic(csv_path: Path, max_rows: int, prefix: str, source: str,
                    currency: str = "USD", default_category: str = "") -> Iterator[dict]:
    """Универсальный процессор для любого e-commerce датасета."""
    with open(csv_path, encoding="utf-8", errors="replace") as f:
        reader = csv.DictReader(f)
        count = 0
        for row in reader:
            if count >= max_rows:
                break
            title = find_col(row, [
                "product_name", "product_title", "title", "name", "Name", "Title",
                "TITLE", "item_name", "product",
            ])
            if not title:
                continue
            price = clean_price(find_col(row, [
                "selling_price", "discounted_price", "actual_price", "price",
                "Price", "mrp", "sale_price", "list_price",
            ]))
            if price <= 0:
                continue
            pid = find_col(row, ["product_id", "asin", "ASIN", "id", "isbn", "uniq_id"])
            img = find_col(row, ["image", "img_link", "image_url", "imgUrl", "thumbnail"])
            brand = find_col(row, ["brand", "Brand", "brand_name", "author", "authors"])
            desc = strip_html(find_col(row, [
                "description", "about_product", "product_description", "details",
            ], ""))[:400] or title
            cat_raw = find_col(row, ["category", "Category", "main_category", "categories"])
            rating_raw = find_col(row, ["rating", "ratings", "stars", "average_rating"])
            rating = 0.0
            try:
                rating = float(rating_raw)
            except ValueError:
                pass
            reviews = int(re.sub(r"[^\d]", "", find_col(row, [
                "reviews", "rating_count", "no_of_ratings", "reviews_count", "num_ratings",
            ], "0")) or "0")
            cat = default_category or detect_category(f"{title} {cat_raw}")
            yield make_row(
                gift_id=f"{prefix}_{pid or count}", title=title, description=desc,
                category=cat, price=price, currency=currency,
                store_url=f"https://www.amazon.com/s?k={title[:40].replace(' ', '+')}",
                image_url=img, brand=brand, rating=rating,
                reviews_count=reviews, source=source,
            )
            count += 1


# ── Датасеты ──────────────────────────────────────────────────────────────────

DATASETS: list[dict] = [
    # ── Главный датасет Amazon 2023 (1.4M товаров) ───────────────────────────
    {
        "key": "amazon_2023",
        "kaggle_id": "asaniczka/amazon-products-dataset-2023-1-4m-products",
        "desc": "Amazon 2023 — 1.4M товаров всех категорий (США) ★★★★★",
        "year": 2023,
        "processor": process_amazon_2023,
        "prefer_file": "amazon_products.csv",
    },

    # ── Beauty / Косметика ────────────────────────────────────────────────────
    {
        "key": "sephora",
        "kaggle_id": "nadyinky/sephora-products-and-skincare-reviews",
        "desc": "Sephora — 8000+ косметических продуктов с рейтингами 2023",
        "year": 2023,
        "processor": process_sephora,
        "prefer_file": "product_info.csv",
    },
    {
        "key": "beauty_2024",
        "kaggle_id": "waqi786/most-used-beauty-cosmetics-products-in-the-world",
        "desc": "Beauty & Cosmetics 2024 — топ мировых брендов ★★★★★",
        "year": 2024,
        "processor": process_beauty_2024,
    },

    # ── Книги ─────────────────────────────────────────────────────────────────
    {
        "key": "books_2025",
        "kaggle_id": "mihikaajayjadhav/books-dataset-15k-books-across-100-categories",
        "desc": "Google Books 2025 — 15K книг в 100+ категориях с ценами ★★★★★",
        "year": 2025,
        "processor": process_books_google,
    },

    # ── Amazon Electronics & Laptops ─────────────────────────────────────────
    {
        "key": "amazon_india_2024",
        "kaggle_id": "karkavelrajaj/amazon-sales-dataset",
        "desc": "Amazon India 2024 — электроника, гаджеты с описаниями (INR)",
        "year": 2024,
        "processor": process_amazon_sales_india,
    },
    {
        "key": "amazon_laptops_2024",
        "kaggle_id": "kanchana1990/amazons-500-bestsellers-in-laptop-gear-2024",
        "desc": "Amazon Bestsellers 2024 — ноутбуки и аксессуары ★★★★★",
        "year": 2024,
        "processor": process_amazon_laptops_2024,
    },
    {
        "key": "amazon_electronics2",
        "kaggle_id": "lokeshparab/amazon-products-dataset",
        "desc": "Amazon — широкий ассортимент по категориям",
        "year": 2023,
        "processor": lambda p, m: process_generic(p, m, "AMZG", "amazon_general", "USD"),
    },

    # ── Etsy / Подарки ────────────────────────────────────────────────────────
    {
        "key": "wedding_gifts_2024",
        "kaggle_id": "kanchana1990/wedding-gift-items-dataset",
        "desc": "Etsy — свадебные подарки 2024, хендмейд (USD) ★★★★★",
        "year": 2024,
        "processor": process_wedding_gifts_etsy,
    },

    # ── Flipkart (Индия) ──────────────────────────────────────────────────────
    {
        "key": "flipkart",
        "kaggle_id": "PromptCloudHQ/flipkart-products",
        "desc": "Flipkart — широкий ассортимент товаров (INR)",
        "year": 2023,
        "processor": process_flipkart,
    },

    # ── AliExpress ────────────────────────────────────────────────────────────
    {
        "key": "furniture_2024",
        "kaggle_id": "kanchana1990/e-commerce-furniture-dataset-2024",
        "desc": "AliExpress — мебель и декор для дома 2024 ★★★★",
        "year": 2024,
        "processor": process_furniture_2024,
    },
    {
        "key": "aliexpress_detail",
        "kaggle_id": "promptcloud/product-details-on-aliexpress",
        "desc": "AliExpress — детальные данные с описаниями",
        "year": 2023,
        "processor": process_aliexpress,
    },

    # ── Myntra (Индия, мода) ──────────────────────────────────────────────────
    {
        "key": "myntra",
        "kaggle_id": "shivamb/fashion-clothing-products-catalog",
        "desc": "Myntra — одежда и аксессуары (INR)",
        "year": 2022,
        "processor": process_myntra,
    },

    # ── Amazon Oct 2024 ───────────────────────────────────────────────────────
    {
        "key": "amazon_oct_2024",
        "kaggle_id": "promptcloud/amazon-product-listing-1st-oct-31st-oct-2024",
        "desc": "Amazon Product Listing — октябрь 2024 ★★★★★",
        "year": 2024,
        "processor": lambda p, m: process_generic(p, m, "AMZOCT", "amazon_oct_2024", "USD"),
    },

    # ── Новые датасеты 2025-2026 ──────────────────────────────────────────────
    {
        "key": "amazon_electronics_2025",
        "kaggle_id": "prothomeshmistry/amazon-electronics-and-accessories-2025",
        "desc": "Amazon Electronics & Accessories 2025 — 9K реальных товаров (INR) ★★★★★",
        "year": 2025,
        "processor": process_amazon_electronics_2025,
    },
    {
        "key": "adidas_2026",
        "kaggle_id": "bsthere/adidas-global-catalogue-2026",
        "desc": "Adidas Global Catalogue 2026 — 45K актуальных товаров ★★★★★",
        "year": 2026,
        "processor": lambda p, m: process_nike_adidas_2026(p, m, "Adidas"),
        "prefer_file": "Adidas_Global.csv",
    },
    {
        "key": "nike_2026",
        "kaggle_id": "bsthere/nike-global-catalogue-2026",
        "desc": "Nike Global Catalogue 2026 — 40K актуальных товаров ★★★★★",
        "year": 2026,
        "processor": lambda p, m: process_nike_adidas_2026(p, m, "Nike"),
        "prefer_file": "Nike_AT.csv",
    },
    {
        "key": "electronics_marketplace_2026",
        "kaggle_id": "asadullahcreative/e-commerce-microphone-marketplace-dataset",
        "desc": "Marketplace Electronics 2026 — микрофоны и аксессуары ★★★★",
        "year": 2026,
        "processor": process_marketplace_electronics_2026,
    },
    {
        "key": "headphones_2026",
        "kaggle_id": "maulikgajera/global-headphones-specifications-and-market-dataset",
        "desc": "Global Headphones Market 2026 — наушники и гарнитуры ★★★",
        "year": 2026,
        "processor": process_headphones_2026,
    },
]


# ── Загрузка ──────────────────────────────────────────────────────────────────

def kaggle_download(dataset_id: str, tmp_dir: Path) -> list[Path]:
    tmp_dir.mkdir(parents=True, exist_ok=True)
    cmd = ["kaggle", "datasets", "download", "-d", dataset_id, "-p", str(tmp_dir), "--unzip"]
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip()[:400])
    return list(tmp_dir.glob("**/*.csv"))


def pick_csv(csvs: list[Path], prefer: str = "") -> Path | None:
    if not csvs:
        return None
    if prefer:
        for p in csvs:
            if p.name == prefer:
                return p
    return max(csvs, key=lambda p: p.stat().st_size)


def save_rows(rows: list[dict], out_path: Path) -> int:
    seen: set[str] = set()
    unique = [r for r in rows if r["gift_id"] not in seen and not seen.add(r["gift_id"])]  # type: ignore[func-returns-value]
    with open(out_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(unique)
    return len(unique)


# ── main ─────────────────────────────────────────────────────────────────────

def main() -> None:
    parser = argparse.ArgumentParser(description="Загрузка датасетов подарков 2024-2026")
    parser.add_argument("--output", default=DEFAULT_OUTPUT)
    parser.add_argument("--list", action="store_true")
    parser.add_argument("--only", nargs="+", metavar="KEY")
    parser.add_argument("--max-rows", type=int, default=60000)
    parser.add_argument("--keep-tmp", action="store_true")
    args = parser.parse_args()

    if args.list:
        print(f"\n{'KEY':<22} {'YEAR':<6} {'KAGGLE ID':<55} ОПИСАНИЕ")
        print("-" * 130)
        for d in DATASETS:
            print(f"{d['key']:<22} {d['year']:<6} {d['kaggle_id']:<55} {d['desc']}")
        print(f"\nВсего: {len(DATASETS)} датасетов")
        return

    try:
        r = subprocess.run(["kaggle", "--version"], capture_output=True, text=True)
        log.info("kaggle CLI: %s", r.stdout.strip())
    except FileNotFoundError:
        log.error("Установи kaggle: pip install kaggle")
        sys.exit(1)

    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    selected = DATASETS
    if args.only:
        keys = set(args.only)
        selected = [d for d in DATASETS if d["key"] in keys]
        if not selected:
            log.error("Нет датасетов с ключами: %s", args.only)
            sys.exit(1)

    total_saved = 0
    succeeded: list[str] = []
    failed: list[str] = []

    for ds in selected:
        key = ds["key"]
        out_csv = output_dir / f"ds_{key}.csv"
        if out_csv.exists() and out_csv.stat().st_size > 4096:
            log.info("[%s] уже есть (%d KB) — пропуск", key, out_csv.stat().st_size // 1024)
            succeeded.append(key)
            continue

        log.info("\n=== [%s] %s ===", key, ds["desc"])
        tmp_dir = output_dir / f"_tmp_{key}"
        try:
            csvs = kaggle_download(ds["kaggle_id"], tmp_dir)
            main_csv = pick_csv(csvs, ds.get("prefer_file", ""))
            if not main_csv:
                log.warning("[%s] нет CSV файлов", key)
                failed.append(key)
                continue
            log.info("  файл: %s (%.1f MB)", main_csv.name, main_csv.stat().st_size / 1e6)

            rows = list(ds["processor"](main_csv, args.max_rows))
            if not rows:
                log.warning("[%s] 0 строк после обработки (проверь колонки)", key)
                failed.append(key)
                continue

            saved = save_rows(rows, out_csv)
            log.info("[%s] сохранено %d строк → %s", key, saved, out_csv.name)
            total_saved += saved
            succeeded.append(key)
        except Exception as e:
            log.warning("[%s] ошибка: %s", key, e)
            failed.append(key)
        finally:
            if not args.keep_tmp and tmp_dir.exists():
                shutil.rmtree(tmp_dir, ignore_errors=True)

    print("\n" + "=" * 60)
    print(f"ИТОГО: {total_saved:,} строк в {len(succeeded)} датасетах")
    if succeeded:
        print(f"  ✓ {', '.join(succeeded)}")
    if failed:
        print(f"  ✗ {', '.join(failed)}")
    print(f"\nСледующий шаг:")
    print(f"  python3 scripts/data/merge_catalog.py \\")
    print(f"    --inputs {args.output}/ \\")
    print(f"    --output services/ml/dataset/catalog_real.csv")
    print("=" * 60)


if __name__ == "__main__":
    main()
