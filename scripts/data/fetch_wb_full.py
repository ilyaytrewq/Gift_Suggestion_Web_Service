#!/usr/bin/env python3
"""
Сбор реальных товаров из Wildberries по подарочным категориям.

Запускать с российского IP (или через VPN RU).
WB блокирует запросы с зарубежных адресов.

Использование:
  python fetch_wb_full.py --output services/ml/dataset/collected/wb_gifts.csv
  python fetch_wb_full.py --output wb.csv --pages 10 --delay 2.5

Результат: CSV с полями
  gift_id, title, description, category, price, currency, store_url, image_url,
  age_min, age_max, brand, rating, reviews_count, discount_pct, source
"""
from __future__ import annotations

import argparse
import csv
import json
import logging
import os
import time
from typing import Iterator

import requests

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
log = logging.getLogger(__name__)

OUTPUT_FIELDS = [
    "gift_id", "title", "description", "category", "price", "currency",
    "store_url", "image_url", "age_min", "age_max",
    "brand", "rating", "reviews_count", "discount_pct", "source",
]

# ─── Запросы, которые охватывают разные категории подарков ───────────────────
GIFT_QUERIES = [
    # Универсальные подарки
    ("подарок женщине",            "Для женщин"),
    ("подарок мужчине",            "Для мужчин"),
    ("подарок ребёнку",            "Для детей"),
    ("подарок на день рождения",   "День рождения"),
    ("подарок маме",               "Для мамы"),
    ("подарок папе",               "Для папы"),
    ("подарок коллеге",            "Для коллег"),
    # Электроника
    ("наушники беспроводные",      "Электроника"),
    ("умные часы",                 "Электроника"),
    ("портативная колонка",        "Электроника"),
    ("электронная книга ридер",    "Электроника"),
    # Красота и уход
    ("парфюм подарочный набор",    "Косметика"),
    ("косметика набор подарок",    "Косметика"),
    ("уходовая косметика набор",   "Косметика"),
    # Книги
    ("книга бестселлер",           "Книги"),
    ("книга подарок",              "Книги"),
    # Игры и хобби
    ("настольная игра",            "Настольные игры"),
    ("пазл 1000 деталей",          "Настольные игры"),
    ("шахматы подарочные",         "Настольные игры"),
    ("набор для рисования",        "Хобби"),
    ("скетчбук художника",         "Хобби"),
    # Дом и уют
    ("ароматические свечи набор",  "Украшения для дома"),
    ("декор для дома",             "Украшения для дома"),
    ("ваза для цветов",            "Украшения для дома"),
    ("фоторамки набор",            "Украшения для дома"),
    # Еда и напитки
    ("чайный набор подарок",       "Еда и напитки"),
    ("кофе подарочный набор",      "Еда и напитки"),
    ("мёд набор подарочный",       "Еда и напитки"),
    # Дети
    ("игрушка плюшевая",           "Детские товары"),
    ("конструктор детский",        "Детские товары"),
    ("набор для опытов дети",      "Детские товары"),
    # Аксессуары
    ("кошелёк кожаный подарок",    "Аксессуары"),
    ("ремень кожаный мужской",     "Аксессуары"),
    ("перчатки кожаные",           "Аксессуары"),
    # Спорт
    ("фитнес браслет",             "Электроника"),
    ("термос подарочный",          "Спорт"),
    ("массажёр подарок",           "Косметика"),
]

# Маппинг subjectName WB → внутренняя категория
WB_CATEGORY_MAP: dict[str, str] = {
    "наушники": "Электроника",
    "гарнитура": "Электроника",
    "смарт-часы": "Электроника",
    "умные часы": "Электроника",
    "колонки": "Электроника",
    "электроника": "Электроника",
    "гаджеты": "Электроника",
    "телефоны": "Электроника",
    "ноутбуки": "Электроника",
    "планшеты": "Электроника",
    "электронные книги": "Электроника",
    "ридеры": "Электроника",
    "духи": "Косметика",
    "парфюмерия": "Косметика",
    "туалетная вода": "Косметика",
    "косметика": "Косметика",
    "уход": "Косметика",
    "набор косметики": "Косметика",
    "массажёры": "Косметика",
    "книги": "Книги",
    "литература": "Книги",
    "настольные игры": "Настольные игры",
    "пазлы": "Настольные игры",
    "шахматы": "Настольные игры",
    "шашки": "Настольные игры",
    "лото": "Настольные игры",
    "домино": "Настольные игры",
    "одежда": "Одежда",
    "футболки": "Одежда",
    "рубашки": "Одежда",
    "свитеры": "Одежда",
    "платья": "Одежда",
    "сумки": "Аксессуары",
    "кошельки": "Аксессуары",
    "ремни": "Аксессуары",
    "перчатки": "Аксессуары",
    "часы": "Аксессуары",
    "украшения": "Аксессуары",
    "аксессуары": "Аксессуары",
    "свечи": "Украшения для дома",
    "ароматы для дома": "Украшения для дома",
    "декор": "Украшения для дома",
    "вазы": "Украшения для дома",
    "картины": "Украшения для дома",
    "фоторамки": "Украшения для дома",
    "интерьер": "Украшения для дома",
    "текстиль": "Украшения для дома",
    "посуда": "Украшения для дома",
    "чай": "Еда и напитки",
    "кофе": "Еда и напитки",
    "конфеты": "Еда и напитки",
    "мёд": "Еда и напитки",
    "шоколад": "Еда и напитки",
    "еда": "Еда и напитки",
    "напитки": "Еда и напитки",
    "игрушки": "Детские товары",
    "мягкие игрушки": "Детские товары",
    "конструкторы": "Детские товары",
    "детские игрушки": "Детские товары",
    "настольные игры для детей": "Детские товары",
    "творчество": "Хобби",
    "рисование": "Хобби",
    "скетчбуки": "Хобби",
    "рукоделие": "Хобби",
    "наборы для творчества": "Хобби",
    "термосы": "Хобби",
    "спорт": "Хобби",
    "подарки": "Хобби",
    "сувениры": "Хобби",
}

DEFAULT_CATEGORY = "Хобби"


def map_wb_category(wb_subject: str) -> str:
    low = wb_subject.lower().strip()
    for key, cat in WB_CATEGORY_MAP.items():
        if key in low:
            return cat
    return DEFAULT_CATEGORY


def _wb_basket_host(product_id: int) -> str:
    """Определяет CDN-хост Wildberries по диапазону vol."""
    vol = product_id // 100000
    ranges = [
        (143, 1), (287, 2), (431, 3), (719, 4), (1007, 5),
        (1061, 6), (1115, 7), (1169, 8), (1313, 9), (1601, 10),
        (1655, 11), (1919, 12), (2045, 13), (2189, 14), (2405, 15),
        (2621, 16), (2837, 17),
    ]
    for threshold, n in ranges:
        if vol <= threshold:
            return f"basket-{n:02d}.wbbasket.ru"
    return "basket-18.wbbasket.ru"


def wb_image_url(product_id: int) -> str:
    host = _wb_basket_host(product_id)
    vol = product_id // 100000
    part = product_id // 1000
    return f"https://{host}/vol{vol}/part{part}/{product_id}/images/big/1.jpg"


def wb_store_url(product_id: int) -> str:
    return f"https://www.wildberries.ru/catalog/{product_id}/detail.aspx"


def make_session() -> requests.Session:
    session = requests.Session()
    session.headers.update({
        "User-Agent": (
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
            "AppleWebKit/537.36 (KHTML, like Gecko) "
            "Chrome/124.0.0.0 Safari/537.36"
        ),
        "Accept": "application/json, text/plain, */*",
        "Accept-Language": "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
        "Accept-Encoding": "gzip, deflate, br",
        "Origin": "https://www.wildberries.ru",
        "Referer": "https://www.wildberries.ru/",
        "sec-ch-ua": '"Google Chrome";v="124", "Chromium";v="124", "Not-A.Brand";v="99"',
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": '"Windows"',
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "cross-site",
        "Connection": "keep-alive",
    })
    return session


def warm_up_session(session: requests.Session) -> None:
    """Получить куки WB через главную страницу."""
    try:
        session.get("https://www.wildberries.ru/", timeout=12)
        time.sleep(1.5)
    except Exception as e:
        log.warning("Warm-up failed: %s", e)


def fetch_wb_page(
    session: requests.Session,
    query: str,
    page: int,
) -> list[dict]:
    import urllib.parse
    encoded_query = urllib.parse.quote(query)
    url = (
        f"https://search.wb.ru/exactmatch/ru/common/v4/search"
        f"?query={encoded_query}"
        f"&resultset=catalog"
        f"&sort=popular"
        f"&page={page}"
        f"&curr=rub"
        f"&dest=-1257786"
        f"&appType=1"
    )
    try:
        resp = session.get(url, timeout=15)
        if resp.status_code == 429:
            log.warning("Rate limited (429), sleeping 10s...")
            time.sleep(10)
            resp = session.get(url, timeout=15)
        resp.raise_for_status()
        data = resp.json()
        return data.get("data", {}).get("products", [])
    except requests.HTTPError as e:
        log.warning("HTTP %s for query=%r page=%d: %s", e.response.status_code, query, page, e)
        return []
    except Exception as e:
        log.warning("Error fetching query=%r page=%d: %s", query, page, e)
        return []


def fetch_wb_descriptions(session: requests.Session, ids: list[int]) -> dict[int, str]:
    """Загрузить описания товаров через card API. Пакет по 10 id."""
    descriptions = {}
    for i in range(0, len(ids), 10):
        batch = ids[i:i + 10]
        nm = ";".join(str(x) for x in batch)
        url = f"https://card.wb.ru/cards/v2/detail?appType=1&curr=rub&dest=-1257786&nm={nm}"
        try:
            resp = session.get(url, timeout=12)
            resp.raise_for_status()
            data = resp.json()
            for card in data.get("data", {}).get("products", []):
                pid = card.get("id")
                desc = card.get("description", "")
                if pid and desc:
                    descriptions[pid] = desc.strip()
        except Exception as e:
            log.warning("Description fetch error for ids %s: %s", batch[:3], e)
        time.sleep(0.5)
    return descriptions


def map_product(
    p: dict,
    query_hint: str,
    descriptions: dict[int, str],
) -> dict | None:
    pid = p.get("id")
    name = (p.get("name") or "").strip()
    if not pid or not name:
        return None

    price_u = p.get("salePriceU") or p.get("priceU") or 0
    orig_u = p.get("priceU") or price_u
    price_rub = price_u // 100

    if price_rub <= 0:
        return None

    discount = 0
    if orig_u > price_u and orig_u > 0:
        discount = int(100 * (orig_u - price_u) / orig_u)

    wb_cat = (p.get("subjectName") or "").strip()
    category = map_wb_category(wb_cat) if wb_cat else DEFAULT_CATEGORY

    description = descriptions.get(pid, "").strip()
    if not description:
        description = name  # fallback: использовать название как описание

    rating = p.get("reviewRating") or 0.0
    feedbacks = p.get("feedbacks") or 0
    brand = (p.get("brand") or "").strip()

    return {
        "gift_id": f"WB{pid}",
        "title": name,
        "description": description,
        "category": category,
        "price": f"{price_rub}.00",
        "currency": "RUB",
        "store_url": wb_store_url(pid),
        "image_url": wb_image_url(pid),
        "age_min": "0",
        "age_max": "99",
        "brand": brand,
        "rating": str(rating),
        "reviews_count": str(feedbacks),
        "discount_pct": str(discount),
        "source": "wildberries",
    }


def collect_wb(
    queries: list[tuple[str, str]],
    pages_per_query: int,
    delay: float,
    fetch_descriptions: bool,
) -> Iterator[dict]:
    session = make_session()
    log.info("Warming up WB session...")
    warm_up_session(session)

    seen_ids: set[str] = set()
    total = 0

    for query, hint in queries:
        log.info("Query: %r (hint: %s)", query, hint)
        query_products = []

        for page in range(1, pages_per_query + 1):
            products = fetch_wb_page(session, query, page)
            if not products:
                break

            query_products.extend(products)
            log.info("  page %d: %d products", page, len(products))

            if len(products) < 5:
                break
            time.sleep(delay)

        ids_to_describe = [p["id"] for p in query_products if p.get("id")]
        descriptions: dict[int, str] = {}
        if fetch_descriptions and ids_to_describe:
            log.info("  Fetching %d descriptions...", len(ids_to_describe))
            descriptions = fetch_wb_descriptions(session, ids_to_describe)

        for p in query_products:
            row = map_product(p, hint, descriptions)
            if row is None:
                continue
            if row["gift_id"] in seen_ids:
                continue
            seen_ids.add(row["gift_id"])
            total += 1
            yield row

    log.info("Total WB products collected: %d", total)


def main() -> None:
    parser = argparse.ArgumentParser(description="Сбор подарков из Wildberries")
    parser.add_argument("--output", default="services/ml/dataset/collected/wb_gifts.csv")
    parser.add_argument("--pages", type=int, default=5, help="Страниц на запрос (100 товаров/стр)")
    parser.add_argument("--delay", type=float, default=2.0, help="Задержка между запросами (сек)")
    parser.add_argument("--descriptions", action="store_true", help="Загружать полные описания (медленнее)")
    parser.add_argument("--queries", type=int, default=len(GIFT_QUERIES), help="Кол-во запросов из списка")
    args = parser.parse_args()

    os.makedirs(os.path.dirname(args.output) or ".", exist_ok=True)

    queries = GIFT_QUERIES[: args.queries]
    log.info("Сбор из %d запросов × %d страниц", len(queries), args.pages)

    count = 0
    with open(args.output, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=OUTPUT_FIELDS)
        writer.writeheader()

        for row in collect_wb(queries, args.pages, args.delay, args.descriptions):
            writer.writerow(row)
            count += 1
            if count % 100 == 0:
                f.flush()
                log.info("Сохранено %d строк", count)

    log.info("Итог: %d товаров → %s", count, args.output)


if __name__ == "__main__":
    main()
