"""Загрузка модели и inference. Пока эхо-заглушка."""
from __future__ import annotations

import logging
import os
import pickle
from typing import Any

log = logging.getLogger(__name__)


class Ranker:
    """Singleton-обёртка над LightGBM моделью."""

    _instance: "Ranker | None" = None
    _model: Any = None
    model_version: str = "echo-stub-0.1"

    @classmethod
    def get(cls) -> "Ranker":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def load(self, model_path: str) -> None:
        if not os.path.exists(model_path):
            log.warning("Model file not found: %s — using echo stub", model_path)
            return
        try:
            with open(model_path, "rb") as f:
                self._model = pickle.load(f)
            self.model_version = os.path.basename(model_path).replace(".pkl", "")
            log.info("Model loaded from %s", model_path)
        except Exception as exc:
            log.warning("Failed to load model: %s — using echo stub", exc)

    def predict(self, features: list[dict]) -> list[float]:
        """Возвращает scores для кандидатов.

        Если модель не загружена — возвращает убывающие заглушечные скоры.
        """
        if self._model is None:
            return [1.0 - i * 0.01 for i in range(len(features))]

        try:
            import pandas as pd  # noqa: PLC0415

            df = pd.DataFrame(features)
            numeric_cols = df.select_dtypes(include="number").columns.tolist()
            return self._model.predict(df[numeric_cols]).tolist()
        except Exception as exc:
            log.warning("Predict failed: %s — falling back to echo scores", exc)
            return [1.0 - i * 0.01 for i in range(len(features))]
