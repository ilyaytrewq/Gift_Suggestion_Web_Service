"""Оценка качества модели на отложенной выборке."""
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
from training.train_lightgbm import evaluate_ndcg, FEATURE_COLS

log = logging.getLogger(__name__)


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")

    parser = argparse.ArgumentParser(description="Оценка NDCG модели")
    parser.add_argument("--model", required=True)
    parser.add_argument("--dataset", default="dataset/training.parquet")
    parser.add_argument("--out", default=None)
    args = parser.parse_args()

    with open(args.model, "rb") as f:
        model = pickle.load(f)

    df = pd.read_parquet(args.dataset)
    metrics = evaluate_ndcg(model, df, k_list=[5, 10])

    log.info("NDCG@5  = %.4f", metrics["ndcg@5"])
    log.info("NDCG@10 = %.4f", metrics["ndcg@10"])

    if args.out:
        with open(args.out, "w") as f:
            json.dump(metrics, f, indent=2)
        log.info("Метрики сохранены → %s", args.out)


if __name__ == "__main__":
    main()
