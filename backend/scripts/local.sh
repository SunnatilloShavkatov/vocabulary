#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
ENV_FILE="$ROOT_DIR/.env"
ENV_EXAMPLE="$ROOT_DIR/.env.docker.example"

cmd="${1:-up}"
dry_run="${DRY_RUN:-0}"

if [[ "${2:-}" == "--dry-run" ]]; then
  dry_run="1"
fi

run() {
  if [[ "$dry_run" == "1" ]]; then
    echo "[dry-run] $*"
    return 0
  fi
  "$@"
}

ensure_env_file() {
  if [[ -f "$ENV_FILE" ]]; then
    return 0
  fi

  if [[ -f "$ENV_EXAMPLE" ]]; then
    cp "$ENV_EXAMPLE" "$ENV_FILE"
    echo "Created .env from .env.docker.example"
    return 0
  fi

  echo "Missing .env and .env.docker.example" >&2
  return 1
}

compose() {
  run docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

up() {
  ensure_env_file
  compose up -d postgres redis rabbitmq migrate gateway auth-service
  echo "Stack up: gateway + auth-service + postgres + redis + rabbitmq + migrations"
  echo "Health: http://localhost:8080/healthz"
  echo "Auth Health: http://localhost:8081/healthz"
  echo "Metrics: http://localhost:8080/metrics"
}

down() {
  ensure_env_file
  compose down
  echo "Stack down"
}

status() {
  ensure_env_file
  compose ps
}

logs() {
  ensure_env_file
  compose logs -f --tail=150 gateway auth-service postgres redis rabbitmq
}

case "$cmd" in
  up)
    up
    ;;
  down)
    down
    ;;
  status|ps)
    status
    ;;
  logs)
    logs
    ;;
  *)
    echo "Usage: scripts/local.sh [up|down|status|logs] [--dry-run]" >&2
    exit 1
    ;;
esac
