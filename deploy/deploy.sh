#!/usr/bin/env bash
set -euo pipefail

SERVER_HOST="${1:-87.242.119.199}"
SERVER_PORT="${2:-22}"
SERVER_USER="${SERVER_USER:-project-kdo}"
APP_DIR="${APP_DIR:-/opt/sport}"
SSH_KEY_PATH="${SSH_KEY_PATH:-$HOME/.ssh/cloud_key}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

if [[ -n "${POSTGRES_PASSWORD}" && ! "${POSTGRES_PASSWORD}" =~ ^[A-Za-z0-9_.-]+$ ]]; then
    echo "POSTGRES_PASSWORD must contain only A-Z, a-z, 0-9, underscore, dot, or dash." >&2
    echo "This keeps the DATABASE_URL valid without URL encoding." >&2
    exit 1
fi

if [[ ! -f "${SSH_KEY_PATH}" ]]; then
    echo "SSH key not found: ${SSH_KEY_PATH}" >&2
    exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SSH_ARGS=(-p "${SERVER_PORT}" -i "${SSH_KEY_PATH}")

ssh "${SSH_ARGS[@]}" "${SERVER_USER}@${SERVER_HOST}" "sudo mkdir -p '${APP_DIR}' && sudo chown '${SERVER_USER}:${SERVER_USER}' '${APP_DIR}'"

rsync -az --delete \
    -e "ssh -p ${SERVER_PORT} -i ${SSH_KEY_PATH}" \
    --exclude '.git/' \
    --exclude '.env.prod' \
    --exclude 'client/node_modules/' \
    --exclude 'client/dist/' \
    --exclude 'server/tmp/' \
    "${ROOT_DIR}/" "${SERVER_USER}@${SERVER_HOST}:${APP_DIR}/"

ssh "${SSH_ARGS[@]}" "${SERVER_USER}@${SERVER_HOST}" bash -s -- "${APP_DIR}" "${POSTGRES_PASSWORD}" <<'REMOTE'
set -euo pipefail

APP_DIR="$1"
POSTGRES_PASSWORD="$2"

cd "${APP_DIR}"

if [[ -z "${POSTGRES_PASSWORD}" && -f .env.prod ]]; then
    POSTGRES_PASSWORD="$(sed -n 's/^POSTGRES_PASSWORD=//p' .env.prod | tail -n 1)"
fi
if [[ -z "${POSTGRES_PASSWORD}" ]]; then
    POSTGRES_PASSWORD="$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')"
fi
if ! [[ "${POSTGRES_PASSWORD}" =~ ^[A-Za-z0-9_.-]+$ ]]; then
    echo "Stored POSTGRES_PASSWORD contains unsupported characters." >&2
    exit 1
fi

touch .env.prod
chmod 0600 .env.prod
printf 'POSTGRES_PASSWORD=%s\n' "${POSTGRES_PASSWORD}" > .env.prod

docker compose --env-file .env.prod -f deploy/docker-compose.prod.yml up -d --build postgres api client
docker compose --env-file .env.prod -f deploy/docker-compose.prod.yml --profile tools run --rm media-sync

if command -v caddy >/dev/null 2>&1; then
    sudo install -d -m 0755 /etc/caddy/conf.d
    sudo install -m 0644 deploy/Caddyfile.sport /etc/caddy/conf.d/sport.register.im.caddy

    if ! sudo grep -Eq '^[[:space:]]*import[[:space:]]+/etc/caddy/conf\.d/\*\.caddy[[:space:]]*$' /etc/caddy/Caddyfile; then
        sudo cp /etc/caddy/Caddyfile "/etc/caddy/Caddyfile.bak.$(date +%Y%m%d%H%M%S)"
        printf '\nimport /etc/caddy/conf.d/*.caddy\n' | sudo tee -a /etc/caddy/Caddyfile >/dev/null
    fi

    sudo caddy validate --config /etc/caddy/Caddyfile
    sudo systemctl reload caddy
else
    echo "Caddy is not installed. Install/configure Caddy and add this site block:" >&2
    cat deploy/Caddyfile.sport >&2
fi
REMOTE
