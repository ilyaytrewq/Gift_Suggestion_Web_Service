#!/usr/bin/env python3
"""
Выгрузка товаров из поискового API Wildberries.

Использование:
  python fetch_wb.py --query "подарок" --pages 5 --output wb_raw.csv

ВНИМАНИЕ: Использует неофициальный публичный поиск WB без авторизации.
Не имеет гарантий стабильности. Ставь задержку (--delay), чтобы не получить бан.
"""

import argparse
import csv
import sys
import time
import logging
from typing import Iterator

try:
    import requests
except ImportError:
    sys.exit("Установи зависимости: pip install requests")

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

WB_SEARCH_URL = "https://search.wb.ru/exactmatch/ru/common/v4/search"
WB_PRODUCT_URL = "https://www.wildberries.ru/catalog/{id}/detail.aspx"
WB_IMAGE_TPL = "https://images.wbstatic.net/c246x328/new/{bucket}0000/{id}-1.jpg"

FIELDNAMES = [
    "gift_id", "title", "description", "category",
    "price", "currency", "store_url", "image_url",
    "age_min", "age_max",
]

MOCK_ROWS = [
    {
        "gift_id": "WB_MOCK_1",
        "title": "Плед флисовый мягкий",
        "description": "Тёплый плед из флиса, подойдёт в подарок.",
        "category": "Дом и уют",
        "price": "1490",
        "currency": "RUB",
        "store_url": "https://www.wildberries.ru/catalog/123456/detail.aspx",
        "image_url": "https://images.wbstatic.net/c246x328/new/1230000/123456-1.jpg",
        "age_min": "0",
        "age_max": "99",
    },
    {
        "gift_id": "WB_MOCK_2",
        "title": "Набор чаев ассорти",
        "description": "Подарочный набор из 10 видов чая.",
        "category": "Еда и напитки",
        "price": "890",
        "currency": "RUB",
        "store_url": "https://www.wildberries.ru/catalog/654321/detail.aspx",
        "image_url": "https://images.wbstatic.net/c246x328/new/6540000/654321-1.jpg",
        "age_min": "0",
        "age_max": "99",
    },
]


def _wb_image_url(product_id: int) -> str:
    bucket = product_id // 100000
    return WB_IMAGE_TPL.format(bucket=bucket, id=product_id)


def _fetch_page(query: str, page: int, session: requests.Session) -> list[dict]:
    params = {
        "TestGroup": "no_test",
        "TestID": "no_test",
        "appType": "1",
        "curr": "rub",
        "dest": "-1257786",
        "page": page,
        "query": query,
        "resultset": "catalog",
        "sort": "popular",
        "suppressSpellcheck": "false",
    }
    try:
        resp = session.get(WB_SEARCH_URL, params=params, timeout=10)
        resp.raise_for_status()
        data = resp.json()
    except Exception as exc:
        log.warning("Ошибка запроса страницы %d: %s", page, exc)
        return []

    products = (data.get("data") or {}).get("products") or []
    return products


def _map_product(p: dict) -> dict:
    product_id = p.get("id", 0)
    price_raw = p.get("salePriceU") or p.get("priceU") or 0
    price = str(price_raw // 100)

    return {
        "gift_id": f"WB{product_id}",
        "title": (p.get("name") or "").strip(),
        "description": (p.get("description") or "").strip(),
        "category": (p.get("subjectName") or p.get("category") or "").strip(),
        "price": price,
        "currency": "RUB",
        "store_url": WB_PRODUCT_URL.format(id=product_id),
        "image_url": _wb_image_url(product_id),
        "age_min": "0",
        "age_max": "99",
    }


def fetch_products(query: str, pages: int, delay: float, mock: bool) -> Iterator[dict]:
    if mock:
        log.info("Режим mock: возвращаем %d тестовых строк", len(MOCK_ROWS))
        yield from MOCK_ROWS
        return

    session = requests.Session()
    session.headers["User-Agent"] = (
        "Mozilla/5.0 (compatible; GiftSuggestionBot/1.0)"
    )

    for page in range(1, pages + 1):
        log.info("Загружаем страницу %d / %d", page, pages)
        products = _fetch_page(query, page, session)
        if not products:
            log.info("Страница %d пустая — останавливаемся", page)
            break

        for p in products:
            row = _map_product(p)
            if row["title"] and row["price"] != "0":
                yield row

        if page < pages:
            time.sleep(delay)


def main() -> None:
    parser = argparse.ArgumentParser(description="Выгрузка товаров из WB")
    parser.add_argument("--query", required=True, help="Поисковый запрос")
    parser.add_argument("--pages", type=int, default=5, help="Кол-во страниц")
    parser.add_argument("--output", required=True, help="Путь к выходному CSV")
    parser.add_argument("--delay", type=float, default=1.0, help="Задержка между запросами (сек)")
    parser.add_argument("--mock", action="store_true", help="Использовать тестовые данные (для CI)")
    args = parser.parse_args()

    count = 0
    with open(args.output, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=FIELDNAMES)
        writer.writeheader()
        for row in fetch_products(args.query, args.pages, args.delay, args.mock):
            writer.writerow(row)
            count += 1

    log.info("Сохранено %d строк → %s", count, args.output)


if __name__ == "__main__":
    main()
