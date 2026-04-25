FROM node:25-alpine AS build

WORKDIR /app

COPY services/frontend/package*.json services/frontend/
COPY services/backend/docs/openapi/backend.yaml services/backend/docs/openapi/backend.yaml

WORKDIR /app/services/frontend
RUN npm ci

COPY services/frontend/ ./

ARG VITE_API_BASE_URL=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build

FROM nginx:1.27-alpine AS runtime

ENV FRONTEND_PUBLIC_URL=http://localhost:5173

COPY services/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY services/frontend/docker/99-print-frontend-url.sh /docker-entrypoint.d/99-print-frontend-url.sh
RUN chmod +x /docker-entrypoint.d/99-print-frontend-url.sh
COPY --from=build /app/services/frontend/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --retries=10 \
  CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
