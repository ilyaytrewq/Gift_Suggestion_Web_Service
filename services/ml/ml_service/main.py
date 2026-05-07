"""Точка входа ML сервиса: gRPC сервер + FastAPI healthz."""
from __future__ import annotations

import logging
import os
import threading

import uvicorn
from fastapi import FastAPI

from ml_service.config import config, root_logging_level_int
from ml_service.grpc_server import create_server
from ml_service.ranker import Ranker

logging.basicConfig(level=root_logging_level_int(), format="%(levelname)s %(name)s %(message)s")
log = logging.getLogger(__name__)

app = FastAPI(title="Gift Suggestion ML Service", docs_url=None, redoc_url=None)


@app.get("/healthz")
def healthz() -> dict:
    return {"status": "ok"}


def start_grpc() -> None:
    ranker = Ranker.get()
    ranker.load(config.model_path)

    server = create_server(ranker, config.grpc_addr)
    server.start()
    log.info("gRPC server started on %s", config.grpc_addr)
    server.wait_for_termination()


def main() -> None:
    grpc_thread = threading.Thread(target=start_grpc, daemon=True)
    grpc_thread.start()

    http_port = int(os.environ.get("ML_HTTP_PORT", "8081"))
    log.info("FastAPI healthz on :%d", http_port)
    uvicorn.run(app, host="0.0.0.0", port=http_port, log_level=config.log_level.lower())


if __name__ == "__main__":
    main()
