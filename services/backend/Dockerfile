FROM golang:1.26-alpine AS builder

WORKDIR /src/services/backend

COPY services/backend/go.mod services/backend/go.sum ./
RUN go mod download

COPY services/backend ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/backend ./cmd/api

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /out/backend /app/backend

EXPOSE 8080

CMD ["/app/backend"]
