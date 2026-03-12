#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [[ -f ".env" ]]; then
  # shellcheck disable=SC1091
  set -a
  source ".env"
  set +a
fi

: "${HTTP_ADDRESS:=:8081}"
: "${MAILER_DRIVER:=smtp}"
: "${SMTP_HOST:=smtp.gmail.com}"
: "${SMTP_PORT:=587}"
: "${SMTP_USERNAME:=viktorov.e1310@gmail.com}"
: "${SMTP_FROM:=viktorov.e1310@gmail.com}"
: "${SUPPORT_RECIPIENT_EMAIL:=viktorov.e1310@gmail.com}"

if [[ -z "${SMTP_PASSWORD:-}" ]]; then
  echo "error: SMTP_PASSWORD is not set." >&2
  echo "create $ROOT_DIR/.env with:" >&2
  echo "SMTP_PASSWORD=your_gmail_app_password" >&2
  exit 1
fi

export HTTP_ADDRESS
export MAILER_DRIVER
export SMTP_HOST
export SMTP_PORT
export SMTP_USERNAME
export SMTP_PASSWORD
export SMTP_FROM
export SUPPORT_RECIPIENT_EMAIL

echo "start: HTTP_ADDRESS=${HTTP_ADDRESS} SMTP_HOST=${SMTP_HOST} SMTP_PORT=${SMTP_PORT} SMTP_USERNAME=${SMTP_USERNAME} SUPPORT_RECIPIENT_EMAIL=${SUPPORT_RECIPIENT_EMAIL}"

exec go run ./cmd/api
