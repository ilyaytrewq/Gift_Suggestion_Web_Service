#!/usr/bin/env bash
# Перевод пользователя в роль admin.
#
# Использование (локально):
#   bash scripts/data/make_admin.sh --email user@example.com
#
# Использование (сервер через SSH):
#   bash scripts/data/make_admin.sh --email user@example.com --server ilyaytrewq@51.250.117.154
#   bash scripts/data/make_admin.sh --email user@example.com --server ilyaytrewq@51.250.117.154 --key ~/.ssh/id_rsa

set -euo pipefail

EMAIL=""
SSH_TARGET=""
SSH_KEY=""
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="gift"
DB_NAME="gift_suggestion"
DB_PASS="gift"
COMPOSE_PROJECT="gift-suggestion"

usage() {
  cat <<EOF
Использование: $(basename "$0") --email EMAIL [OPTIONS]

Обязательные:
  --email EMAIL            Email пользователя

Для локального запуска:
  --db-host HOST           (default: localhost)
  --db-port PORT           (default: 5432)
  --db-user USER           (default: gift)
  --db-name NAME           (default: gift_suggestion)
  --db-pass PASS           (default: gift)

Для удалённого сервера:
  --server ilyaytrewq@51.250.117.154       SSH-цель; если указан — psql через docker exec
  --key PATH               Путь к SSH-ключу
  --compose-project NAME   Имя Docker Compose проекта (default: gift-suggestion)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --email)           EMAIL="$2";           shift 2 ;;
    --server)          SSH_TARGET="$2";      shift 2 ;;
    --key)             SSH_KEY="$2";         shift 2 ;;
    --db-host)         DB_HOST="$2";         shift 2 ;;
    --db-port)         DB_PORT="$2";         shift 2 ;;
    --db-user)         DB_USER="$2";         shift 2 ;;
    --db-name)         DB_NAME="$2";         shift 2 ;;
    --db-pass)         DB_PASS="$2";         shift 2 ;;
    --compose-project) COMPOSE_PROJECT="$2"; shift 2 ;;
    --help|-h)         usage; exit 0 ;;
    *) echo "Неизвестный параметр: $1"; usage; exit 1 ;;
  esac
done

if [[ -z "${EMAIL}" ]]; then
  echo "Ошибка: --email обязателен"
  usage
  exit 1
fi

log() { echo "[$(date '+%H:%M:%S')] $*"; }

run_psql() {
  local query="$1"

  if [[ -n "${SSH_TARGET}" ]]; then
    local ssh_opts="-o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes"
    [[ -n "${SSH_KEY}" ]] && ssh_opts="${ssh_opts} -i ${SSH_KEY}"

    ssh ${ssh_opts} "${SSH_TARGET}" "
      CONTAINER=\$(docker ps \
        --filter 'label=com.docker.compose.service=postgres' \
        --filter 'label=com.docker.compose.project=${COMPOSE_PROJECT}' \
        --format '{{.Names}}' | head -1)
      if [[ -z \"\${CONTAINER}\" ]]; then
        CONTAINER=\$(docker ps --filter 'name=postgres' --format '{{.Names}}' | head -1)
      fi
      if [[ -z \"\${CONTAINER}\" ]]; then
        echo 'Ошибка: postgres-контейнер не найден' >&2; exit 1
      fi
      docker exec -i \"\${CONTAINER}\" psql -U ${DB_USER} -d ${DB_NAME} -c \"${query}\"
    "
  else
    PGPASSWORD="${DB_PASS}" psql \
      -h "${DB_HOST}" -p "${DB_PORT}" \
      -U "${DB_USER}" -d "${DB_NAME}" \
      -c "${query}"
  fi
}

if [[ -n "${SSH_TARGET}" ]]; then
  log "Режим: удалённый сервер (${SSH_TARGET})"
else
  log "Режим: локальная БД (${DB_HOST}:${DB_PORT}/${DB_NAME})"
fi
log "Email: ${EMAIL}"

# Проверяем, что пользователь существует
RESULT=$(run_psql "SELECT id, email, role FROM users WHERE email = '${EMAIL}';" 2>&1)

if echo "${RESULT}" | grep -q "(0 rows)"; then
  echo ""
  echo "Пользователь с email '${EMAIL}' не найден в БД."
  echo "Убедитесь, что он зарегистрировался через API (/api/v1/users)."
  exit 1
fi

log "Текущее состояние:"
echo "${RESULT}"

# Устанавливаем роль admin
UPDATE=$(run_psql "UPDATE users SET role = 'admin' WHERE email = '${EMAIL}' RETURNING id, email, role;")

echo ""
log "Готово:"
echo "${UPDATE}"
