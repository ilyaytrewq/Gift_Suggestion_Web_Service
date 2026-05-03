#!/usr/bin/env python3
"""
Выгрузка товаров через AliExpress Affiliate API (affiliate.product.query).

Требует переменных окружения:
  ALIEXPRESS_APP_KEY    — ключ приложения из Aliexpress Open Platform
  ALIEXPRESS_APP_SECRET — секрет приложения

Использование:
  ALIEXPRESS_APP_KEY=xxx ALIEXPRESS_APP_SECRET=yyy \
    python fetch_aliexpress.py --keywords "gift" --pages 5 --output ali_raw.csv

В режиме --mock работает без ключей (для CI и тестов).
"""

import argparse
import csv
import hashlib
import hmac
import os
import sys
import time
import logging
from datetime import datetime, timezone
from typing import Iterator

try:
    import requests
except ImportError:
    sys.exit("Установи зависимости: pip install requests")

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

AE_API_URL = "https://api-sg.aliexpress.com/sync"
AE_METHOD = "aliexpress.affiliate.product.query"

FIELDNAMES = [
    "gift_id", "title", "description", "category",
    "price", "currency", "store_url", "image_url",
    "age_min", "age_max",
]

MOCK_ROWS = [
    {
        "gift_id": "AE_MOCK_1",
        "title": "Wireless Bluetooth Earbuds",
        "description": "High quality wireless earbuds, great gift.",
        "category": "Электроника",
        "price": "1200",
        "currency": "CNY",
        "store_url": "https://www.aliexpress.com/item/123456789.html",
        "image_url": "https://ae01.alicdn.com/kf/mock1.jpg",
        "age_min": "12",
        "age_max": "99",
    },
    {
        "gift_id": "AE_MOCK_2",
        "title": "Chinese Tea Set Ceramic",
        "description": "Beautiful ceramic tea set, perfect gift.",
        "category": "Еда и напитки",
        "price": "890",
        "currency": "CNY",
        "store_url": "https://www.aliexpress.com/item/987654321.html",
        "image_url": "https://ae01.alicdn.com/kf/mock2.jpg",
        "age_min": "0",
        "age_max": "99",
    },
]


def _sign(params: dict, secret: str) -> str:
    sorted_params = sorted(params.items())
    sign_str = secret + "".join(f"{k}{v}" for k, v in sorted_params) + secret
    return hmac.new(secret.encode(), sign_str.encode(), hashlib.md5).hexdigest().upper()


def _build_params(app_key: str, app_secret: str, keywords: str, page: int) -> dict:
    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S")
    params = {
        "app_key": app_key,
        "timestamp": timestamp,
        "sign_method": "md5",
        "method": AE_METHOD,
        "keywords": keywords,
        "page_no": str(page),
        "page_size": "50",
        "sort": "SALE_PRICE_ASC",
        "target_currency": "CNY",
        "target_language": "RU",
        "tracking_id": "gift_suggestion",
    }
    params["sign"] = _sign(params, app_secret)
    return params


def _fetch_page(keywords: str, page: int, app_key: str, app_secret: str,
                session: requests.Session) -> list[dict]:
    params = _build_params(app_key, app_secret, keywords, page)
    try:
        resp = session.get(AE_API_URL, params=params, timeout=15)
        resp.raise_for_status()
        data = resp.json()
    except Exception as exc:
        log.warning("Ошибка запроса страницы %d: %s", page, exc)
        return []

    result = (
        data
        .get("aliexpress_affiliate_product_query_response", {})
        .get("resp_result", {})
    )
    if result.get("resp_code") != 200:
        log.warning("API вернул ошибку: %s", result.get("resp_msg"))
        return []

    return result.get("result", {}).get("products", {}).get("product", []) or []


def _map_product(p: dict) -> dict:
    product_id = str(p.get("product_id", ""))
    price = str(p.get("target_sale_price") or p.get("target_original_price") or "0")
    price = price.split(".")[0]

    return {
        "gift_id": f"AE{product_id}",
        "title": (p.get("product_title") or "").strip(),
        "description": (p.get("product_title") or "").strip(),
        "category": (p.get("first_level_category_name") or "").strip(),
        "price": price,
        "currency": "CNY",
        "store_url": p.get("promotion_link") or p.get("product_detail_url") or "",
        "image_url": p.get("product_main_image_url") or "",
        "age_min": "0",
        "age_max": "99",
    }


def fetch_products(keywords: str, pages: int, delay: float,
                   app_key: str, app_secret: str, mock: bool) -> Iterator[dict]:
    if mock:
        log.info("Режим mock: возвращаем %d тестовых строк", len(MOCK_ROWS))
        yield from MOCK_ROWS
        return

    session = requests.Session()
    for page in range(1, pages + 1):
        log.info("Загружаем страницу %d / %d", page, pages)
        products = _fetch_page(keywords, page, app_key, app_secret, session)
        if not products:
            log.info("Страница %d пустая — останавливаемся", page)
            break

        for p in products:
            row = _map_product(p)
            if row["title"] and row["store_url"]:
                yield row

        if page < pages:
            time.sleep(delay)


def main() -> None:
    parser = argparse.ArgumentParser(description="Выгрузка товаров из AliExpress Affiliate API")
    parser.add_argument("--keywords", required=True, help="Ключевые слова")
    parser.add_argument("--pages", type=int, default=5, help="Кол-во страниц (max 50 товаров/стр)")
    parser.add_argument("--output", required=True, help="Путь к выходному CSV")
    parser.add_argument("--delay", type=float, default=1.0, help="Задержка между запросами (сек)")
    parser.add_argument("--mock", action="store_true", help="Тестовые данные без API ключей (для CI)")
    args = parser.parse_args()

    app_key = os.environ.get("ALIEXPRESS_APP_KEY", "")
    app_secret = os.environ.get("ALIEXPRESS_APP_SECRET", "")

    if not args.mock and (not app_key or not app_secret):
        sys.exit(
            "Нужны ALIEXPRESS_APP_KEY и ALIEXPRESS_APP_SECRET. "
            "Используй --mock для тестового запуска."
        )

    count = 0
    with open(args.output, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=FIELDNAMES)
        writer.writeheader()
        for row in fetch_products(args.keywords, args.pages, args.delay, app_key, app_secret, args.mock):
            writer.writerow(row)
            count += 1

    log.info("Сохранено %d строк → %s", count, args.output)


if __name__ == "__main__":
    main()
