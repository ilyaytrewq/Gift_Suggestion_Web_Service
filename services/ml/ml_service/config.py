import os


class Config:
    grpc_host: str = os.environ.get("ML_GRPC_HOST", "0.0.0.0")
    grpc_port: int = int(os.environ.get("ML_GRPC_PORT", "50051"))
    model_path: str = os.environ.get("ML_MODEL_PATH", "models/lightgbm_v0_1_0.pkl")
    db_dsn: str = os.environ.get("ML_DB_DSN", "")
    log_level: str = os.environ.get("LOG_LEVEL", "INFO")

    @property
    def grpc_addr(self) -> str:
        return f"{self.grpc_host}:{self.grpc_port}"


config = Config()
