FROM python:3.13-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends gcc libpq-dev && rm -rf /var/lib/apt/lists/*

COPY services/ml/pyproject.toml .
RUN pip install --no-cache-dir --upgrade setuptools && \
    pip install --no-cache-dir ".[dev]" 2>/dev/null || pip install --no-cache-dir .

COPY services/ml/ml_service ./ml_service
COPY services/ml/models ./models

ENV ML_GRPC_PORT=50051 \
    ML_HTTP_PORT=8081 \
    LOG_LEVEL=INFO

EXPOSE 50051 8081

CMD ["python", "-m", "ml_service.main"]
