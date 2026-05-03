#!/usr/bin/env python3
"""
Сбор товаров из AliExpress тремя способами (в порядке предпочтения):

  1. Affiliate API  — официальный, требует регистрацию на portals.aliexpress.com
                      (ALIEXPRESS_APP_KEY + ALIEXPRESS_APP_SECRET)
  2. Web scraping   — парсинг HTML/JSON через requests + BeautifulSoup
                      (не требует ключей, но нестабилен)
  3. --mock         — тестовые данные (для CI и разработки без подключения к сети)

Использование:
  # Способ 1 — Affiliate API (рекомендуется)
  export ALIEXPRESS_APP_KEY=ваш_ключ
  export ALIEXPRESS_APP_SECRET=ваш_секрет
  python fetch_aliexpress_full.py --output services/ml/dataset/collected/ali_gifts.csv

  # Способ 2 — Web scraping (без ключей)
  python fetch_aliexpress_full.py --mode web --output ali_gifts.csv --pages 5

  # Способ 3 — Mock для тестирования
  python fetch_aliexpress_full.py --mock --output ali_mock.csv
"""
from __future__ import annotations

import argparse
import csv
import hashlib
import hmac
import json
import logging
import os
import re
import time
from datetime import datetime, timezone
from typing import Iterator

import requests

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category", "price", "currency",
    "store_url", "image_url", "age_min", "age_max",
    "brand", "rating", "reviews_count", "discount_pct", "source",
]

# Поисковые запросы для подарочных товаров с AliExpress (на английском — лучше работает)
GIFT_QUERIES = [
    # Электроника
    "wireless earbuds",
    "smart watch gift",
    "bluetooth speaker",
    "portable charger power bank",
    "led desk lamp",
    "mini camera gadget",
    # Красота / уход
    "skincare gift set",
    "perfume gift set",
    "massage device",
    "hair care kit",
    # Дом и уют
    "aromatherapy candle set",
    "tea set ceramic",
    "home decoration gift",
    "wall art decor",
    "plant pot ceramic",
    # Еда и напитки
    "chinese tea gift box",
    "coffee gift set",
    "nuts gift set",
    # Хобби и творчество
    "art supplies set",
    "watercolor paint set",
    "puzzle 1000 piece",
    "diy craft kit",
    # Игры
    "card game gift",
    "chess wooden set",
    # Аксессуары
    "leather wallet gift",
    "silk scarf gift",
    "jewelry gift set",
    "watch luxury",
    # Детские
    "educational toy set",
    "plush toy soft",
    "building blocks kids",
    # Спорт / здоровье
    "yoga mat set",
    "fitness tracker band",
    "neck massager",
]

# Маппинг первой категории AliExpress → внутренняя
AE_CATEGORY_MAP: dict[str, str] = {
    "consumer electronics": "Электроника",
    "phones": "Электроника",
    "computer": "Электроника",
    "audio": "Электроника",
    "smart devices": "Электроника",
    "wearable": "Электроника",
    "camera": "Электроника",
    "beauty": "Косметика",
    "skin care": "Косметика",
    "hair care": "Косметика",
    "perfume": "Косметика",
    "health": "Косметика",
    "massage": "Косметика",
    "home decor": "Украшения для дома",
    "home & garden": "Украшения для дома",
    "furniture": "Украшения для дома",
    "lighting": "Украшения для дома",
    "kitchen": "Украшения для дома",
    "tableware": "Украшения для дома",
    "bedding": "Украшения для дома",
    "food": "Еда и напитки",
    "tea": "Еда и напитки",
    "coffee": "Еда и напитки",
    "snack": "Еда и напитки",
    "toy": "Детские товары",
    "baby": "Детские товары",
    "children": "Детские товары",
    "kids": "Детские товары",
    "sport": "Хобби",
    "fitness": "Хобби",
    "art": "Хобби",
    "craft": "Хобби",
    "office": "Хобби",
    "book": "Книги",
    "bag": "Аксессуары",
    "wallet": "Аксессуары",
    "jewelry": "Аксессуары",
    "watch": "Аксессуары",
    "scarf": "Аксессуары",
    "accessories": "Аксессуары",
    "clothing": "Одежда",
    "apparel": "Одежда",
    "dress": "Одежда",
    "shirt": "Одежда",
    "games": "Настольные игры",
    "puzzle": "Настольные игры",
    "chess": "Настольные игры",
    "board game": "Настольные игры",
}

DEFAULT_CATEGORY = "Хобби"


def map_ae_category(cat: str) -> str:
    low = (cat or "").lower()
    for key, internal in AE_CATEGORY_MAP.items():
        if key in low:
            return internal
    return DEFAULT_CATEGORY


# ═══════════════════════════════════════════════════════
#  Способ 1: AliExpress Affiliate API
# ═══════════════════════════════════════════════════════

AE_API_URL = "https://api-sg.aliexpress.com/sync"
AE_METHOD = "aliexpress.affiliate.product.query"


def _sign_ae(params: dict, secret: str) -> str:
    sorted_items = sorted(params.items())
    sign_str = secret + "".join(f"{k}{v}" for k, v in sorted_items) + secret
    return hmac.new(secret.encode("utf-8"), sign_str.encode("utf-8"), hashlib.md5).hexdigest().upper()


def affiliate_search(
    session: requests.Session,
    app_key: str,
    app_secret: str,
    keywords: str,
    page: int = 1,
    page_size: int = 50,
) -> list[dict]:
    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S")
    params: dict[str, str] = {
        "app_key": app_key,
        "timestamp": timestamp,
        "sign_method": "md5",
        "method": AE_METHOD,
        "keywords": keywords,
        "page_no": str(page),
        "page_size": str(page_size),
        "sort": "SALE_PRICE_ASC",
        "target_currency": "CNY",
        "target_language": "RU",
        "tracking_id": "gift_suggestion_ml",
        "fields": "product_id,product_title,target_sale_price,target_original_price,"
                  "evaluate_rate,order_count,product_main_image_url,"
                  "product_detail_url,promotion_link,first_level_category_name,"
                  "second_level_category_name",
    }
    params["sign"] = _sign_ae(params, app_secret)

    try:
        resp = session.get(AE_API_URL, params=params, timeout=20)
        resp.raise_for_status()
        data = resp.json()
    except Exception as e:
        log.warning("AE Affiliate API error (query=%r, page=%d): %s", keywords, page, e)
        return []

    result = (
        data.get("aliexpress_affiliate_product_query_response", {})
        .get("resp_result", {})
    )
    if result.get("resp_code") != 200:
        log.warning("AE API error: code=%s msg=%s", result.get("resp_code"), result.get("resp_msg"))
        return []

    return result.get("result", {}).get("products", {}).get("product", []) or []


def map_affiliate_product(p: dict, query_no: int) -> dict | None:
    pid = str(p.get("product_id", "")).strip()
    title = (p.get("product_title") or "").strip()
    if not pid or not title:
        return None

    price_str = str(p.get("target_sale_price") or p.get("target_original_price") or "0")
    try:
        price_cny = int(float(price_str.split(".")[0]))
    except ValueError:
        return None

    orig_str = str(p.get("target_original_price") or price_str)
    try:
        orig_cny = int(float(orig_str.split(".")[0]))
    except ValueError:
        orig_cny = price_cny

    discount = 0
    if orig_cny > price_cny > 0:
        discount = int(100 * (orig_cny - price_cny) / orig_cny)

    cat1 = (p.get("first_level_category_name") or "").strip()
    cat2 = (p.get("second_level_category_name") or "").strip()
    category = map_ae_category(cat1) if cat1 else map_ae_category(cat2)

    store_url = (p.get("promotion_link") or p.get("product_detail_url") or "").strip()
    if not store_url:
        store_url = f"https://www.aliexpress.com/item/{pid}.html"

    image_url = (p.get("product_main_image_url") or "").strip()
    rating_str = str(p.get("evaluate_rate") or "0").replace("%", "")
    try:
        rating = round(float(rating_str) / 20, 1)  # 80% → 4.0
    except ValueError:
        rating = 0.0

    orders = str(p.get("order_count") or 0)

    return {
        "gift_id": f"AE{pid}",
        "title": title,
        "description": title,  # API не даёт описание в базовом запросе
        "category": category,
        "price": f"{price_cny}.00",
        "currency": "CNY",
        "store_url": store_url,
        "image_url": image_url,
        "age_min": "0",
        "age_max": "99",
        "brand": "",
        "rating": str(rating),
        "reviews_count": orders,
        "discount_pct": str(discount),
        "source": "aliexpress_affiliate",
    }


# ═══════════════════════════════════════════════════════
#  Способ 2: Web scraping (без ключей)
# ═══════════════════════════════════════════════════════

def make_ae_session() -> requests.Session:
    session = requests.Session()
    session.headers.update({
        "User-Agent": (
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/124.0.0.0 Safari/537.36"
        ),
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "ru-RU,ru;q=0.9,zh-CN;q=0.8,en;q=0.7",
        "Accept-Encoding": "gzip, deflate, br",
    })
    return session


def _extract_json_from_script(html: str) -> list[dict]:
    """Извлечь данные товаров из JSON в теге <script> страницы AliExpress."""
    products = []
    # AliExpress embedding JSON: window.runParams = {...}
    patterns = [
        r'window\.runParams\s*=\s*({.*?});\s*(?:window|var)',
        r'"items"\s*:\s*(\[.*?\])',
        r'"itemList"\s*:\s*(\[.*?\])',
    ]
    for pattern in patterns:
        m = re.search(pattern, html, re.DOTALL)
        if not m:
            continue
        try:
            data = json.loads(m.group(1))
            if isinstance(data, list):
                products.extend(data)
            elif isinstance(data, dict):
                for key in ("items", "itemList", "result", "data"):
                    if key in data and isinstance(data[key], list):
                        products.extend(data[key])
                        break
        except json.JSONDecodeError:
            continue
    return products


def web_scrape_page(session: requests.Session, query: str, page: int) -> list[dict]:
    import urllib.parse
    enc = urllib.parse.quote(query)
    url = (
        f"https://www.aliexpress.com/w/wholesale-{enc.replace('%20', '-')}.html"
        f"?SearchText={enc}&SortType=total_tranpro_desc&page={page}"
    )
    try:
        resp = session.get(url, timeout=20)
        resp.raise_for_status()
        return _extract_json_from_script(resp.text)
    except Exception as e:
        log.warning("AE web scrape error query=%r page=%d: %s", query, page, e)
        return []


def map_web_product(p: dict, idx: int) -> dict | None:
    """Нормализация продукта из web-scraping JSON (структура может отличаться)."""
    if not isinstance(p, dict):
        return None
    pid = (
        str(p.get("productId") or p.get("id") or p.get("itemId") or "").strip()
    )
    title = (
        p.get("title") or p.get("name") or p.get("productTitle") or ""
    ).strip()
    if not pid or not title:
        return None

    # Цена — несколько возможных полей
    price_field = (
        p.get("price") or p.get("salePrice") or p.get("minSalePrice") or {}
    )
    if isinstance(price_field, dict):
        price_str = str(price_field.get("minAmount") or price_field.get("value") or "0")
    else:
        price_str = str(price_field)
    try:
        price_cny = int(float(re.sub(r"[^\d.]", "", price_str)))
    except ValueError:
        return None
    if price_cny <= 0:
        return None

    cat = (p.get("categoryName") or p.get("firstLevelCategoryName") or "").strip()
    category = map_ae_category(cat)

    store_url = (p.get("productUrl") or p.get("url") or "").strip()
    if not store_url:
        store_url = f"https://www.aliexpress.com/item/{pid}.html"
    elif store_url.startswith("//"):
        store_url = "https:" + store_url

    image_url = (
        p.get("imageUrl") or p.get("image") or p.get("mainImage") or ""
    ).strip()
    if image_url.startswith("//"):
        image_url = "https:" + image_url

    rating = 0.0
    rating_raw = p.get("starRating") or p.get("avg_star") or p.get("averageStar") or 0
    try:
        rating = round(float(rating_raw), 1)
    except ValueError:
        pass

    orders = str(p.get("trade") or p.get("sold") or p.get("tradeCount") or 0)
    orders = re.sub(r"[^\d]", "", str(orders))

    return {
        "gift_id": f"AE{pid}",
        "title": title,
        "description": title,
        "category": category,
        "price": f"{price_cny}.00",
        "currency": "CNY",
        "store_url": store_url,
        "image_url": image_url,
        "age_min": "0",
        "age_max": "99",
        "brand": "",
        "rating": str(rating),
        "reviews_count": orders,
        "discount_pct": "0",
        "source": "aliexpress_web",
    }


# ═══════════════════════════════════════════════════════
#  Способ 3: Mock
# ═══════════════════════════════════════════════════════

MOCK_ROWS = [
    {
        "gift_id": "AE_MOCK_EARBUDS",
        "title": "Wireless Bluetooth Earbuds TWS",
        "description": "High quality wireless earbuds, noise canceling, 24h battery life. Great gift.",
        "category": "Электроника",
        "price": "599.00",
        "currency": "CNY",
        "store_url": "https://www.aliexpress.com/item/1005001234567.html",
        "image_url": "https://ae01.alicdn.com/kf/earbuds.jpg",
        "age_min": "12",
        "age_max": "99",
        "brand": "QCY",
        "rating": "4.5",
        "reviews_count": "12500",
        "discount_pct": "30",
        "source": "mock",
    },
    {
        "gift_id": "AE_MOCK_TEABOX",
        "title": "Chinese Premium Pu-erh Tea Gift Box",
        "description": "Premium aged pu-erh tea, 8 varieties, beautiful gift packaging.",
        "category": "Еда и напитки",
        "price": "189.00",
        "currency": "CNY",
        "store_url": "https://www.aliexpress.com/item/1005002345678.html",
        "image_url": "https://ae01.alicdn.com/kf/teabox.jpg",
        "age_min": "0",
        "age_max": "99",
        "brand": "TenFu",
        "rating": "4.8",
        "reviews_count": "3200",
        "discount_pct": "15",
        "source": "mock",
    },
    {
        "gift_id": "AE_MOCK_PUZZLEMAP",
        "title": "3D Wooden Puzzle World Map",
        "description": "Handcrafted 3D wooden world map puzzle, 150 pieces, perfect desk decoration and gift.",
        "category": "Настольные игры",
        "price": "245.00",
        "currency": "CNY",
        "store_url": "https://www.aliexpress.com/item/1005003456789.html",
        "image_url": "https://ae01.alicdn.com/kf/puzzle.jpg",
        "age_min": "12",
        "age_max": "99",
        "brand": "Wooden City",
        "rating": "4.7",
        "reviews_count": "856",
        "discount_pct": "0",
        "source": "mock",
    },
]


# ═══════════════════════════════════════════════════════
#  Основной сборщик
# ═══════════════════════════════════════════════════════

def collect(
    mode: str,
    app_key: str,
    app_secret: str,
    queries: list[str],
    pages: int,
    delay: float,
    mock: bool,
) -> Iterator[dict]:
    if mock:
        log.info("Mock mode: %d rows", len(MOCK_ROWS))
        yield from MOCK_ROWS
        return

    session = requests.Session()
    session.headers["User-Agent"] = (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0.0.0 Safari/537.36"
    )

    seen: set[str] = set()
    total = 0

    for query in queries:
        log.info("Query: %r", query)

        for page in range(1, pages + 1):
            if mode == "affiliate":
                raw = affiliate_search(session, app_key, app_secret, query, page)
                mapper = map_affiliate_product
            else:
                raw = web_scrape_page(session, query, page)
                mapper = map_web_product

            if not raw:
                break

            log.info("  page %d: %d items", page, len(raw))

            for i, item in enumerate(raw):
                row = mapper(item, i)
                if row is None or row["gift_id"] in seen:
                    continue
                seen.add(row["gift_id"])
                total += 1
                yield row

            time.sleep(delay)

    log.info("Total AliExpress products: %d", total)


def main() -> None:
    parser = argparse.ArgumentParser(description="Сбор товаров из AliExpress")
    parser.add_argument("--output", default="services/ml/dataset/collected/ali_gifts.csv")
    parser.add_argument("--mode", choices=["affiliate", "web"], default="affiliate",
                        help="affiliate (с ключами) или web (без ключей, нестабильно)")
    parser.add_argument("--pages", type=int, default=5, help="Страниц на запрос")
    parser.add_argument("--delay", type=float, default=1.5, help="Задержка между запросами (сек)")
    parser.add_argument("--queries", type=int, default=len(GIFT_QUERIES), help="Кол-во запросов")
    parser.add_argument("--mock", action="store_true", help="Тестовые данные без сети")
    args = parser.parse_args()

    app_key = os.environ.get("ALIEXPRESS_APP_KEY", "")
    app_secret = os.environ.get("ALIEXPRESS_APP_SECRET", "")

    if args.mode == "affiliate" and not args.mock and (not app_key or not app_secret):
        print(
            "Для режима 'affiliate' нужны переменные окружения:\n"
            "  export ALIEXPRESS_APP_KEY=ваш_ключ\n"
            "  export ALIEXPRESS_APP_SECRET=ваш_секрет\n\n"
            "Регистрация (бесплатно): https://portals.aliexpress.com\n\n"
            "Или запусти с --mode web (без ключей) или --mock (тест)."
        )
        raise SystemExit(1)

    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)
    queries = GIFT_QUERIES[: args.queries]

    count = 0
    with open(args.output, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS)
        writer.writeheader()

        for row in collect(args.mode, app_key, app_secret, queries, args.pages, args.delay, args.mock):
            writer.writerow(row)
            count += 1
            if count % 100 == 0:
                f.flush()
                log.info("Сохранено %d строк", count)

    log.info("Итог: %d товаров → %s", count, args.output)


if __name__ == "__main__":
    main()
