import logging
import os


def _read_log_level_name() -> str:
    ml = os.environ.get("ML_LOG_LEVEL", "").strip()
    raw = ml if ml else (os.environ.get("LOG_LEVEL") or "INFO")
    name = str(raw).strip().upper()
    return name if name else "INFO"


class Config:
    grpc_host: str = os.environ.get("ML_GRPC_HOST", "0.0.0.0")
    grpc_port: int = int(os.environ.get("ML_GRPC_PORT", "50051"))
    model_path: str = os.environ.get("ML_MODEL_PATH", "models/lightgbm_v0_1_0.pkl")
    db_dsn: str = os.environ.get("ML_DB_DSN", "")
    log_level: str = _read_log_level_name()

    @property
    def grpc_addr(self) -> str:
        return f"{self.grpc_host}:{self.grpc_port}"


config = Config()


def root_logging_level_int() -> int:
    return logging.getLevelNamesMapping().get(config.log_level, logging.INFO)
