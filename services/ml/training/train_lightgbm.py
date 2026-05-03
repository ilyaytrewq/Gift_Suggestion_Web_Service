"""Обучение LightGBM LambdaRank модели."""
from __future__ import annotations

import argparse
import json
import logging
import os
import pickle
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import numpy as np
import pandas as pd

log = logging.getLogger(__name__)

FEATURE_COLS = [
    "price_cents",
    "price_bucket",
    "budget_cents",
    "budget_bucket",
    "price_within_budget",
    "budget_gap",
    "budget_ratio",
    "age_fits",
    "interest_overlap",
    "tfidf_score",
    "category_index",
    "already_in_wishlist",
    "gender_encoded",
]


def ndcg_at_k(labels: list[int], k: int = 5) -> float:
    labels = labels[:k]
    dcg = sum((2**r - 1) / np.log2(i + 2) for i, r in enumerate(labels))
    ideal = sorted(labels, reverse=True)
    idcg = sum((2**r - 1) / np.log2(i + 2) for i, r in enumerate(ideal))
    return dcg / idcg if idcg > 0 else 0.0


def evaluate_ndcg(model: object, df_val: pd.DataFrame, k_list: list[int] = [5, 10]) -> dict:
    import lightgbm as lgb

    X_val = df_val[FEATURE_COLS]
    y_val = df_val["label"].values
    preds = model.predict(X_val)

    scores = {f"ndcg@{k}": [] for k in k_list}
    for qid in df_val["query_id"].unique():
        mask = df_val["query_id"] == qid
        q_labels = y_val[mask]
        q_preds = preds[mask]
        order = np.argsort(q_preds)[::-1]
        ranked_labels = q_labels[order].tolist()
        for k in k_list:
            scores[f"ndcg@{k}"].append(ndcg_at_k(ranked_labels, k))

    return {key: float(np.mean(vals)) for key, vals in scores.items()}


def train(dataset_path: str, model_out: str, metrics_out: str, seed: int = 42) -> None:
    import lightgbm as lgb

    log.info("Загрузка датасета: %s", dataset_path)
    df = pd.read_parquet(dataset_path)
    log.info("Датасет: %d строк, %d запросов", len(df), df["query_id"].nunique())

    # train/val split by query_id
    query_ids = df["query_id"].unique()
    np.random.seed(seed)
    np.random.shuffle(query_ids)
    split = int(len(query_ids) * 0.85)
    train_qids = set(query_ids[:split])
    val_qids = set(query_ids[split:])

    df_train = df[df["query_id"].isin(train_qids)].reset_index(drop=True)
    df_val = df[df["query_id"].isin(val_qids)].reset_index(drop=True)

    X_train = df_train[FEATURE_COLS]
    y_train = df_train["label"].values
    group_train = df_train.groupby("query_id").size().values

    X_val = df_val[FEATURE_COLS]
    y_val = df_val["label"].values
    group_val = df_val.groupby("query_id").size().values

    train_data = lgb.Dataset(X_train, label=y_train, group=group_train)
    val_data = lgb.Dataset(X_val, label=y_val, group=group_val, reference=train_data)

    params = {
        "objective": "lambdarank",
        "metric": "ndcg",
        "ndcg_eval_at": [5, 10],
        "learning_rate": 0.05,
        "num_leaves": 15,
        "min_data_in_leaf": 20,
        "lambda_l1": 0.1,
        "lambda_l2": 0.1,
        "feature_fraction": 0.8,
        "bagging_fraction": 0.8,
        "bagging_freq": 5,
        "verbose": -1,
        "seed": seed,
    }

    log.info("Обучение LightGBM LambdaRank...")
    model = lgb.train(
        params,
        train_data,
        num_boost_round=200,
        valid_sets=[val_data],
        callbacks=[lgb.early_stopping(20, verbose=False), lgb.log_evaluation(50)],
    )

    metrics = evaluate_ndcg(model, df_val)
    log.info("Метрики на валидации: %s", metrics)

    os.makedirs(os.path.dirname(model_out), exist_ok=True)
    with open(model_out, "wb") as f:
        pickle.dump(model, f)
    log.info("Модель сохранена: %s", model_out)

    with open(metrics_out, "w") as f:
        json.dump(metrics, f, indent=2)
    log.info("Метрики сохранены: %s", metrics_out)


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")

    parser = argparse.ArgumentParser(description="Обучение LightGBM ранжировщика")
    parser.add_argument("--dataset", default="dataset/training.parquet")
    parser.add_argument("--out", default="models/lightgbm_v0_1_0.pkl")
    parser.add_argument("--metrics", default="models/metrics.json")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    train(args.dataset, args.out, args.metrics, args.seed)


if __name__ == "__main__":
    main()
