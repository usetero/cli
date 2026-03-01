#!/usr/bin/env bash
set -euo pipefail

env_name="${TERO_ENV:-local}"
if [ "$env_name" = "local" ]; then
  exec "$@"
fi

exec doppler run -c "$env_name" -- "$@"
